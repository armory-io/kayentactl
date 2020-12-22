package kayenta

import (
	"context"
	"time"
)

//UpdateScopes is a helper function that will replace the values in an ExecRequest to produce the proper request
func UpdateScopes(scopes []Scope, scope, startTimeIso, endTimeIso string, controlOffset time.Duration) []Scope {
	updatedScopes := []Scope{}
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

//WaitForComplete this issues a call out to kayenta and then loops as it waits for it to finish. Likely we'll have to refactor the
//call signature to inject dependencies, for example, a ticker perhaps should live otuside of this function
func WaitForComplete(ctx context.Context, executionID string, client StandaloneCanaryAnalysisAPI, ticker *time.Ticker, logger StdLogger) error {
	if logger == nil {
		logger = NoopStdLogger{}
	}
	done := make(chan bool, 1)
	var err1 error
	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- true
				return
			case <-ticker.C:
				logger.Println("checking on execution status")
				res, err := client.GetStandaloneCanaryAnalysis(executionID)
				if err != nil {
					logger.Println(err.Error())
					err1 = err
					done <- true
					return
				}
				if res.Complete {
					logger.Printf("execution is complete with status %s", res.Status)
					done <- true
					return
				}

				logger.Printf("execution is still running with status %s", res.Status)
			}
		}
	}()
	<-done
	return err1
}
