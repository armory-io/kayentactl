package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"text/template"

	"github.com/armory-io/kayentactl/pkg/kayenta"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/olekukonko/tablewriter"
)

var ErrNotComplete = errors.New("execution is still in analysis")

const AsciiKayenta string = `
   __ _______  _______  ___________  _____________ 
  / //_/ _ \ \/ / __/ |/ /_  __/ _ |/ ___/_  __/ / 
 / ,< / __ |\  / _//    / / / / __ / /__  / / / /__
/_/|_/_/ |_|/_/___/_/|_/ /_/ /_/ |_\___/ /_/ /____/
`
const asciiReport string = `

Analysis Report For Execution ID: {{ .ID }}

Summary
-------
Status: {{ .Status }}
Final Score: {{ .FinalScore }}
Message: {{ .Message }}
HasWarnings: {{ .HasWarnings }}

Measurements
{{ .Measurements }}

Group Results
{{ .Results }}

Stage Results
`

type asciiReportData struct {
	ID           string
	Status       string
	FinalScore   string
	Message      string
	HasWarnings  bool
	Results      string
	Measurements string
}

func resultToAsciiReportData(result kayenta.GetStandaloneCanaryAnalysisOutput) (asciiReportData, error) {
	scores := result.CanaryAnalysisExecutionResult.CanaryScores
	canaryResults := result.CanaryAnalysisExecutionResult.CanaryExecutionResults
	hasResults := len(canaryResults) > 0

	reportData := asciiReportData{
		ID:          color.GreenString(result.PipelineID),
		HasWarnings: result.CanaryAnalysisExecutionResult.HasWarnings,
	}

	if hasResults {
		lastResult := canaryResults[len(canaryResults)-1]
		resultsTable, err := tableFromJudgeResult(lastResult.Result.JudgeResult)
		if err != nil {
			return asciiReportData{}, err
		}
		measurementsTables, err := tableFromMeasurements(lastResult.Result.JudgeResult)
		if err != nil {
			return asciiReportData{}, err
		}
		reportData.Results = resultsTable
		reportData.Measurements = measurementsTables
	}

	execStatus := color.GreenString(result.ExecutionStatus)
	score := "0"
	if len(scores) > 0 {
		score = strconv.Itoa(int(scores[len(scores)-1]))
	}
	scoreStr := color.GreenString(score)
	finalMsg := color.GreenString(result.CanaryAnalysisExecutionResult.CanaryScoreMessage)
	if !result.CanaryAnalysisExecutionResult.DidPassThresholds {
		scoreStr = color.RedString(score)
		finalMsg = color.RedString(result.CanaryAnalysisExecutionResult.CanaryScoreMessage)
		execStatus = color.YellowString(result.ExecutionStatus)
	}
	reportData.FinalScore = scoreStr
	reportData.Measurements = finalMsg

	if result.ExecutionStatus == "TERMINAL" {
		execStatus = color.RedString(result.ExecutionStatus)
	}

	reportData.Status = execStatus
	return reportData, nil
}

func tableFromJudgeResult(result kayenta.JudgeResult) (string, error) {
	writer := table.NewWriter()
	writer.AppendHeader(table.Row{"Group", "Score"})
	for _, score := range result.GroupScores {
		writer.AppendRow(table.Row{score.Name, score.Score})
	}
	return writer.Render(), nil
}

func tableFromMeasurements(result kayenta.JudgeResult) (string, error) {

	wb := new(bytes.Buffer)
	table := tablewriter.NewWriter(wb)
	table.SetHeader([]string{"Name", "Groups", "Results", "Reason"})
	table.SetAutoWrapText(false)

	for _, score := range result.Results {
		groups := strings.Join(score.Groups, ",")
		statusColor := tablewriter.Colors{tablewriter.Bold, tablewriter.FgWhiteColor}
		reasonColor := tablewriter.Colors{tablewriter.Bold, tablewriter.FgWhiteColor}
		classification := strings.ToUpper(score.Classification)
		if classification == "PASS" {
			statusColor = tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor}
		} else if classification == "HIGH" || classification == "LOW" {
			statusColor = tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor}
			reasonColor = tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor}
		}
		r := []string{score.Name, groups, classification, score.ClassificationReason}
		table.Rich(r, []tablewriter.Colors{{}, {}, statusColor, reasonColor})
	}
	table.Render()
	return wb.String(), nil
}

func TableReport(result kayenta.GetStandaloneCanaryAnalysisOutput) ([]byte, error) {
	tmpl, err := template.New("asciiReport").Parse(asciiReport)
	if err != nil {
		return nil, err
	}
	input, err := resultToAsciiReportData(result)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, input); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func JsonReport(result kayenta.GetStandaloneCanaryAnalysisOutput) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

func Report(result kayenta.GetStandaloneCanaryAnalysisOutput, format string, writer io.Writer) error {
	if !result.Complete && format != "json" {
		return ErrNotComplete
	}

	var b []byte
	var err error
	switch format {
	case "json":
		b, err = JsonReport(result)
	default:
		b, err = TableReport(result)
	}
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, bytes.NewReader(b))
	return err

}
