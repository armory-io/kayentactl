package analysis

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/armory-io/kayentactl/pkg/kayenta"
	"github.com/briandowns/spinner"
	"github.com/olekukonko/tablewriter"
)

// GraphicalProgressPrinter implements a kayenta.ProgressFunc
// that outputs the analysis of an analysis while we wait for
// the final result
type GraphicalProgressPrinter struct {
	s         *spinner.Spinner
	numChecks int
	out       io.Writer
}

func NewDefaultGraphicalProgressPrinter() *GraphicalProgressPrinter {
	return &GraphicalProgressPrinter{
		s:         spinner.New(spinner.CharSets[9], 100*time.Millisecond),
		numChecks: 0,
		out:       os.Stdout,
	}
}

func (pp *GraphicalProgressPrinter) PrintProgress(res kayenta.GetStandaloneCanaryAnalysisOutput) {
	pp.numChecks = pp.numChecks + 1
	printInPlace(TableStatus(res), pp.numChecks == 1, pp.out)
}

func (pp *GraphicalProgressPrinter) Start() {
	pp.s.Start()
}

func (pp *GraphicalProgressPrinter) Stop() {
	pp.s.Stop()
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

func printInPlace(o string, createTermSpace bool, out io.Writer) {
	newLines := countRune(o, '\n')
	if createTermSpace {
		for i := 0; i < newLines; i++ {
			fmt.Fprintln(out)
		}
	}
	fmt.Fprintf(out, "\033[%dA\033[0G\033[0J", newLines)
	fmt.Fprintf(out, o)

}

func TableStatus(o kayenta.GetStandaloneCanaryAnalysisOutput) string {
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
