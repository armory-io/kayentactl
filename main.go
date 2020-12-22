package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/kayentactl/pkg/kayenta"
)

const (
	//TODO: This is being hosted in Isaac's personal github account, we'll need to move this somewhere better.
	defaultCanaryConfig string = "https://gist.githubusercontent.com/imosquera/399a89ad65e4f625fc2e0f0822dc5911/raw/2a0afe8fb482d57afdcb1188dfcdf8bf15403b8c/canary_config.json"
)

func main() {
	var (
		scope, configLocation, control, experiment, kayentaURL, startTimeIso, endTimeIso string
		checkInterval, timeout                                                           time.Duration
	)
	flag.StringVar(&configLocation, "canary-config-url", defaultCanaryConfig, "location of the canary config to use")
	flag.StringVar(&scope, "scope", "", "Scope to use for both experiment & control metrics. Example: kube_deploy:myappname")
	flag.StringVar(&control, "control", "", "application to use as the experiment control (i.e. baseline)")
	flag.StringVar(&experiment, "experiment", "", "application to use as the experiment  (i.e. canary)")
	flag.StringVar(&kayentaURL, "kayenta-url", "http://localhost:8090", "URL for kayenta service")
	flag.StringVar(&startTimeIso, "start-time-iso", "", "start time for the analysis in ISO format. Ex: 2020-12-20T14:49:31.647Z")
	flag.StringVar(&endTimeIso, "end-time-iso", "", "end time for the analysis in ISO format. Ex: 2020-12-20T15:49:31.647Z")

	flag.DurationVar(&checkInterval, "interval", time.Second*10, "polling interval")
	flag.DurationVar(&timeout, "timeout", 1*time.Hour, "timeout")
	flag.Parse()

	log.Printf("fetching canary config from %s", configLocation)
	resp, err := http.Get(configLocation)
	if err != nil {
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

	input.ExecutionRequest.Scopes = kayenta.UpdateScopes(input.ExecutionRequest.Scopes, scope, startTimeIso, endTimeIso)

	kc := kayenta.NewDefaultClient(kayenta.ClientBaseURL(kayentaURL))

	// start standalone canary
	log.Info("starting canary analysis")
	output, err := kc.StartStandaloneCanaryAnalysis(input)
	if err != nil {
		log.Fatalf("error starting canary analysis: %s", err.Error())
	}

	analysisID := output.CanaryAnalysisExecutionID
	log.Info(fmt.Sprintf("canary analysis started with id %s", analysisID))

	// poll until standalone canary is complete
	log.Info("polling until canary analysis is complete")
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	ticker := time.NewTicker(checkInterval)
	if err := kayenta.WaitForComplete(ctx, analysisID, kc, ticker); err != nil {
		log.Fatalf(err.Error())
	}
	// generate some kind of report
	log.Info("canary analysis complete.")
}
