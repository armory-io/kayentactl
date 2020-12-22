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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/armory-io/kayentactl/pkg/kayenta"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	//TODO: This is being hosted in Isaac's personal github account, we'll need to move this somewhere better.
	defaultCanaryConfig string = "https://gist.githubusercontent.com/imosquera/399a89ad65e4f625fc2e0f0822dc5911/raw/2a0afe8fb482d57afdcb1188dfcdf8bf15403b8c/canary_config.json"
)

// TODO: get rid of these package global variables. it was easier to port existing code by using them.
var (
	scope, configLocation, control, experiment, startTimeIso, endTimeIso string
	lifetimeDuration, checkInterval, timeout                             time.Duration
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		kc := kayenta.NewDefaultClient(kayenta.ClientBaseURL(kayentaURL))

		log.Printf("fetching canary config from %s", configLocation)
		resp, err := http.Get(configLocation)
		if err != nil {
			log.Error(err)
			log.Fatalf("Could not get default canary config json at locations: %s", configLocation)
		}
		defer resp.Body.Close()
		log.Println("canary config fetched successfully")

		var input kayenta.StandaloneCanaryAnalysisInput
		err = json.NewDecoder(resp.Body).Decode(&input)
		if err != nil {
			log.Error(err)
			log.Fatal("could not decode canary config JSON, exiting")
		}

		input.ExecutionRequest.LifetimeDurationMins = int(lifetimeDuration.Minutes())
		input.ExecutionRequest.Scopes = kayenta.UpdateScopes(input.ExecutionRequest.Scopes, scope, startTimeIso, endTimeIso)

		// start standalone canary
		log.Infof("starting canary analysis with kayenta host: %v", kayentaURL)
		output, err := kc.StartStandaloneCanaryAnalysis(input)
		if err != nil {
			log.Fatalf("error starting canary analysis: %s", err.Error())
		}

		analysisID := output.CanaryAnalysisExecutionID
		log.Info(fmt.Sprintf("canary analysis started with id %s", analysisID))

		// poll until standalone canary is complete
		log.Info("polling until canary analysis is complete")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ticker := time.NewTicker(checkInterval)
		if err := kayenta.WaitForComplete(ctx, analysisID, kc, ticker); err != nil {
			log.Fatalf(err.Error())
		}
		// generate some kind of report
		log.Info("canary analysis complete.")
		log.Info("getting analysis result")
		result, err := kc.GetStandaloneCanaryAnalysis(analysisID)
		if err != nil {
			log.Fatalf("failed to get analysis result: %s", err.Error())
		}

		exitCode := 0
		msg := "analysis was successful"
		if !result.IsSuccessful() {
			msg = fmt.Sprintf("analysis failed. result: %s", result.Status)
			exitCode = 1
		}
		log.Println(msg)
		os.Exit(exitCode)
	},
}

func init() {
	analysisCmd.AddCommand(startCmd)
	flags := startCmd.Flags()
	flags.StringVar(&configLocation, "canary-config-url", defaultCanaryConfig, "location of canary configuration")
	flags.StringVarP(&scope, "scope", "s", "", "name of the scope to use")
	flags.StringVarP(&control, "control", "c", "", "application to use as the experiment control (i.e. baseline)")
	flags.StringVarP(&experiment, "experiment", "e", "", "application to use as the experiment  (i.e. canary)")
	flags.StringVar(&startTimeIso, "start-time-iso", "", "start time for the analysis in ISO format. Ex: 2020-12-20T14:49:31.647Z")
	flags.StringVar(&endTimeIso, "end-time-iso", "", "end time for the analysis in ISO format. Ex: 2020-12-20T15:49:31.647Z")

	flags.DurationVar(&lifetimeDuration, "lifetime-duration", time.Minute*5, "Total duration time for the analysis")
	flags.DurationVar(&checkInterval, "interval", time.Second*10, "polling interval")
	flags.DurationVar(&timeout, "timeout", time.Hour*1, "timeout")
}
