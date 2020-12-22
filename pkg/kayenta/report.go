package kayenta

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"text/template"

	"github.com/jedib0t/go-pretty/table"
)

var ErrNotComplete = errors.New("execution is still in progress")

const asciiReport = `Analysis Report - {{ .ID }}

Summary
-------
Status: {{ .Status }}
Final Score: {{ .FinalScore }}
Message: {{ .Message }}
HasWarnings: {{ .HasWarnings }}

Results
-------
{{ .Results }}
`

type asciiReportData struct {
	ID          string
	Status      string
	FinalScore  float64
	Message     string
	HasWarnings bool
	Results     string
}

func resultToAsciiReportData(result GetStandaloneCanaryAnalysisOutput) (asciiReportData, error) {
	scores := result.CanaryAnalysisExecutionResult.CanaryScores
	canaryResults := result.CanaryAnalysisExecutionResult.CanaryExecutionResults
	lastResult := canaryResults[len(canaryResults)-1]

	resultsTable, err := tableFromJudgeResult(lastResult.Result.JudgeResult)
	if err != nil {
		return asciiReportData{}, err
	}
	return asciiReportData{
		ID:          result.PipelineID,
		Status:      result.ExecutionStatus,
		FinalScore:  scores[len(scores)-1],
		Message:     result.CanaryAnalysisExecutionResult.CanaryScoreMessage,
		HasWarnings: result.CanaryAnalysisExecutionResult.HasWarnings,
		Results:     resultsTable,
	}, nil
}

func tableFromJudgeResult(result JudgeResult) (string, error) {
	writer := table.NewWriter()
	writer.AppendHeader(table.Row{"Group", "Score"})
	for _, score := range result.GroupScores {
		writer.AppendRow(table.Row{score.Name, score.Score})
	}
	return writer.Render(), nil
}

func TableReport(result GetStandaloneCanaryAnalysisOutput) ([]byte, error) {
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

func JsonReport(result GetStandaloneCanaryAnalysisOutput) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

func Report(result GetStandaloneCanaryAnalysisOutput, format string, writer io.Writer) error {
	if !result.Complete {
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
