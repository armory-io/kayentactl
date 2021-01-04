package analysis

import (
	"context"
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
