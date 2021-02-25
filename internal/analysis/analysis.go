package analysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/armory-io/kayentactl/pkg/kayenta"
)

//UpdateScopes is a helper function that will replace the values in an ExecRequest to produce the proper request
func UpdateScopes(scopes []kayenta.Scope, scope, startTimeIso, endTimeIso string, controlOffset time.Duration) []kayenta.Scope {
	updatedScopes := []kayenta.Scope{}
	for _, s := range scopes {
		s.ExperimentScope = scope
		s.ControlScope = scope
		s.StartTimeIso = startTimeIso
		s.EndTimeIso = endTimeIso
		s.ControlOffsetInMinutes = int(controlOffset.Minutes())
		updatedScopes = append(updatedScopes, s)
	}
	return updatedScopes
}

type scopeCoordinates struct {
	scope, location string
}

func coordinates(scope string) (*scopeCoordinates, error) {
	splitScope := strings.Split(scope, "/")
	if len(splitScope) == 0 {
		return nil, fmt.Errorf("scope could not be determined")
	}

	if len(splitScope) == 1 {
		return &scopeCoordinates{scope: splitScope[0], location: ""}, nil
	}

	return &scopeCoordinates{
		scope:    splitScope[1],
		location: splitScope[0],
	}, nil
}

type ExecutionRequestContext struct {
	ControlScope, ExperimentScope string
	StartTimeIso, EndTimeIso      string

	ControlOffset                              time.Duration
	AnalysisIntervalMins, LifetimeDurationMins time.Duration

	Thresholds kayenta.Threshold
}

func BuildExecutionRequest(ctx ExecutionRequestContext) (*kayenta.ExecutionRequest, error) {
	scope, err := BuildScope(ctx.ControlScope, ctx.ExperimentScope)
	if err != nil {
		return nil, fmt.Errorf("could not construct execution request: %w", err)
	}

	scope.StartTimeIso = ctx.StartTimeIso
	scope.EndTimeIso = ctx.EndTimeIso
	scope.ControlOffsetInMinutes = int(ctx.ControlOffset.Minutes())
	request := kayenta.ExecutionRequest{
		Scopes:               []kayenta.Scope{*scope},
		AnalysisIntervalMins: int(ctx.AnalysisIntervalMins.Minutes()),
		LifetimeDurationMins: int(ctx.LifetimeDurationMins.Minutes()),
		Thresholds:           ctx.Thresholds,
	}
	return &request, nil
}

func BuildScope(control, experiment string) (*kayenta.Scope, error) {
	scope := kayenta.Scope{ScopeName: "default"}
	{
		coord, err := coordinates(control)
		if err != nil {
			return nil, fmt.Errorf("could not build scope for control: %w", err)
		}
		scope.ControlScope = coord.scope
		scope.ControlLocation = coord.location
	}
	{
		coord, err := coordinates(experiment)
		if err != nil {
			return nil, fmt.Errorf("could not build scope for experiment: %w", err)
		}
		scope.ExperimentScope = coord.scope
		scope.ExperimentLocation = coord.location
	}

	return &scope, nil
}

type ProgressFunc func(res kayenta.GetStandaloneCanaryAnalysisOutput)

//WaitForComplete this issues a call out to kayenta and then loops as it waits for it to finish. Likely we'll have to refactor the
//call signature to inject dependencies, for example, a ticker perhaps should live otuside of this function
// progressFunc is called on ever interval where the execution is not complete. if the execution is complete,
// progressFunc will not be called and the function will terminate.
func WaitForComplete(ctx context.Context, executionID string, client kayenta.StandaloneCanaryAnalysisAPI, ticker *time.Ticker, progressFunc ProgressFunc) error {
	done := make(chan bool, 1)
	var err1 error

	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- true
				return
			case <-ticker.C:
				res, err := client.GetStandaloneCanaryAnalysis(executionID)
				if err != nil {
					err1 = err
					done <- true
					return
				}

				if res.Complete {
					done <- true
					return
				}
				if progressFunc != nil {
					progressFunc(res)
				}
			}
		}
	}()
	<-done
	return err1
}
