/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/armory-io/kayentactl/internal/canaryConfig"

	"github.com/armory-io/kayentactl/internal/report"

	"github.com/armory-io/kayentactl/internal/analysis"

	"github.com/armory-io/kayentactl/pkg/kayenta"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// TODO: get rid of these package global variables. it was easier to port existing code by using them.
var (
	scope, configLocation, control, experiment, startTimeIso, endTimeIso      string
	controlOffset, lifetimeDuration, analysisInterval, checkInterval, timeout time.Duration
	noWait                                                                    bool
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if !noColor {
			fmt.Printf("%v\n", color.HiMagentaString(report.AsciiKayenta))
		}
		kc := kayenta.NewDefaultClient(kayenta.ClientBaseURL(kayentaURL))

		log.Debugf("Fetching canary config from: %s", color.BlueString(configLocation))
		input, err := canaryConfig.GetCanaryConfig(configLocation)
		if err != nil {
			log.Fatalf("failed to fetch and parse canary config: %s", err.Error())
		}

		input.ExecutionRequest.AnalysisIntervalMins = int(analysisInterval.Minutes())
		input.ExecutionRequest.Scopes = analysis.UpdateScopes(input.ExecutionRequest.Scopes, scope, startTimeIso, endTimeIso, controlOffset)
		input.ExecutionRequest.LifetimeDurationMins = int(lifetimeDuration.Minutes())

		// start standalone canary
		log.Debugf("Analysis Execution starting with kayenta host: %v", color.BlueString(kayentaURL))
		output, err := kc.StartStandaloneCanaryAnalysis(*input)
		if err != nil {
			log.Fatalf("error starting canary analysis: %s", err.Error())
		}
		analysisID := output.CanaryAnalysisExecutionID
		log.Info(fmt.Sprintf("Analysis Execution ID: %s", color.GreenString(analysisID)))

		// if the no-wait flag is set, we exit early. this enables users
		// to implement their own wait login within scripts
		if noWait {
			return
		}

		// poll until standalone canary is complete
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ticker := time.NewTicker(checkInterval)
		progressPrinter := analysis.NewDefaultGraphicalProgressPrinter()
		progressPrinter.Start()
		if err := analysis.WaitForComplete(ctx, analysisID, kc, ticker, progressPrinter.PrintProgress); err != nil {
			log.Fatalf(err.Error())
		}
		progressPrinter.Stop()

		// generate some kind of report
		result, err := kc.GetStandaloneCanaryAnalysis(analysisID)
		if err != nil {
			log.Fatalf("Failed to get analysis result: %s", err.Error())
		}

		if err := report.Report(result, "pretty", os.Stdout); err != nil {
			log.Fatalf("error generating analysis report: %s", err.Error())
		}

		fmt.Println(analysis.TableStatus(result))

		exitCode := 1
		if result.IsSuccessful() {
			exitCode = 0
		}
		os.Exit(exitCode)
	},
}

func init() {
	analysisCmd.AddCommand(startCmd)
	flags := startCmd.Flags()
	flags.StringVar(&configLocation, "canary-config", "canary.json", "location of canary configuration")
	flags.StringVarP(&scope, "scope", "s", "", "name of the scope to use")
	flags.StringVarP(&control, "control", "c", "", "application to use as the experiment control (i.e. baseline)")
	flags.StringVarP(&experiment, "experiment", "e", "", "application to use as the experiment  (i.e. canary)")
	flags.StringVar(&startTimeIso, "start-time-iso", "", "start time for the analysis in ISO format. Ex: 2020-12-20T14:49:31.647Z")
	flags.StringVar(&endTimeIso, "end-time-iso", "", "end time for the analysis in ISO format. Ex: 2020-12-20T15:49:31.647Z")

	flags.DurationVar(&analysisInterval, "analysis-interval", 1*time.Minute, "Minutes between each analysis. Default is once per minute")
	flags.DurationVar(&lifetimeDuration, "lifetime-duration", time.Minute*5, "Total duration time for the analysis")
	flags.DurationVar(&checkInterval, "interval", time.Second*5, "polling interval")
	flags.DurationVar(&timeout, "timeout", time.Hour, "timeout")
	flags.DurationVar(&controlOffset, "control-offset", time.Hour, "The control offset to compare against the experiment, by default is your new deployment")

	flags.BoolVar(&noWait, "no-wait", false, "don't wait for canary execution to complete before exiting")
}
