package kayenta

import (
	"context"
	"log"
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
		s.ControlOffsetInMins = int(controlOffset.Minutes())
		updatedScopes = append(updatedScopes, s)
	}
	return updatedScopes
}

//WaitForComplete this issues a call out to kayenta and then loops as it waits for it to finish. Likely we'll have to refactor the
//call signature to inject dependencies, for example, a ticker perhaps should live otuside of this function
func WaitForComplete(ctx context.Context, executionID string, client *DefaultClient, ticker *time.Ticker) error {
	done := make(chan bool, 1)
	var err1 error
	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- true
				return
			case <-ticker.C:
				log.Println("checking on execution status")
				res, err := client.GetStandaloneCanaryAnalysis(executionID)
				if err != nil {
					log.Println(err.Error())
					err1 = err
					done <- true
					return
				}
				if isComplete(res) {
					log.Printf("execution is complete with status %s\n", res.Status)
					done <- true
					return
				}

				log.Printf("execution is still running with status %s\n", res.Status)
			}
		}
	}()
	<-done
	return err1
}

func isComplete(status GetStandaloneCanaryAnalysisOutput) bool {
	for _, s := range []string{"canceled", "stopped", "succeeded", "failed_continue", "terminal"} {
		if status.Status == s {
			return true
		}
	}
	return false
}
