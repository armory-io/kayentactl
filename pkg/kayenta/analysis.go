package kayenta

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/olekukonko/tablewriter"
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

func countRune(s string, r rune) int {
	count := 0
	for _, c := range s {
		if c == r {
			count++
		}
	}
	return count
}

func printInPlace(o string, createTermSpace bool) {
	newLines := countRune(o, '\n')
	if createTermSpace {
		for i := 0; i < newLines; i++ {
			fmt.Println()
		}
	}
	fmt.Printf("\033[%dA\033[0G\033[0J", newLines)
	fmt.Printf(o)

}

func TableStatus(o GetStandaloneCanaryAnalysisOutput) string {
	//termLinesNeeded := len(o.Stages) + fixedTableWriterLines

	// if numChecks == 1 {
	// 	//create space in terminal for the table
	// 	for i := 0; i < termLinesNeeded; i++ {
	// 		fmt.Println()
	// 	}
	// }
	wb := new(bytes.Buffer)
	table := tablewriter.NewWriter(wb)
	table.SetHeader([]string{"Name", "TYPE", "STATUS"})
	table.SetAutoWrapText(false)

	for _, s := range o.Stages {
		// name := s.Name
		// if len(name) > 30 {
		// 	name = name[0:30]
		// }

		statusColor := tablewriter.Colors{tablewriter.Bold, tablewriter.FgWhiteColor}
		if s.Status == "RUNNING" {
			statusColor = tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor}
		} else if s.Status == "SUCCEEDED" {
			statusColor = tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor}
		} else if s.Status == "TERMINAL" {
			statusColor = tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor}
		}
		d := []string{s.Name, s.StageType, s.Status}
		table.Rich(d, []tablewriter.Colors{{}, {}, statusColor})
	}

	//fmt.Printf("\033[%dA\033[0G\033[0J", termLinesNeeded)
	table.Render() // Sends bytes to our buffer
	return wb.String()
}

//WaitForComplete this issues a call out to kayenta and then loops as it waits for it to finish. Likely we'll have to refactor the
//call signature to inject dependencies, for example, a ticker perhaps should live otuside of this function
func WaitForComplete(ctx context.Context, executionID string, client StandaloneCanaryAnalysisAPI, ticker *time.Ticker, logger StdLogger) error {
	if logger == nil {
		logger = NoopStdLogger{}
	}
	done := make(chan bool, 1)
	var err1 error
	numChecks := 0

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond) // Build our new spinner
	s.Start()

	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- true
				return
			case <-ticker.C:
				//logger.Println("checking on execution status")
				res, err := client.GetStandaloneCanaryAnalysis(executionID)
				if err != nil {
					logger.Println(err.Error())
					err1 = err
					done <- true
					return
				}

				if res.Complete {
					s.Stop()
					done <- true
					return
				}
				numChecks++
				o := TableStatus(res)
				printInPlace(o, numChecks == 1)
			}
		}
	}()
	<-done
	return err1
}
