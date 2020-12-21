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
	defaultCanaryConfig string = "https://gist.githubusercontent.com/imosquera/399a89ad65e4f625fc2e0f0822dc5911/raw/canary_config.json"
)

func main() {
	var (
		appName, configLocation, control, experiment, kayentaURL string
		checkInterval, timeout                                   time.Duration
	)
	flag.StringVar(&configLocation, "canary-config", defaultCanaryConfig, "location of the canary config to use")
	flag.StringVar(&appName, "application", "", "name of the application to use")
	flag.StringVar(&control, "control", "", "application to use as the experiment control (i.e. baseline)")
	flag.StringVar(&experiment, "experiment", "", "application to use as the experiment  (i.e. canary)")
	flag.StringVar(&kayentaURL, "url", "http://localhost:8090", "URL for kayenta service")

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

	var er kayenta.ExecutionRequest
	err = json.NewDecoder(resp.Body).Decode(&er)
	if err != nil {
		log.Error(err)
		log.Fatal("could not decode canary config JSON, exiting")
	}

	kc := kayenta.NewDefaultClient(kayenta.ClientBaseURL(kayentaURL))

	// start standalone canary
	log.Info("starting canary analysis")
	output, err := kc.StartStandaloneCanaryAnalysis(kayenta.StandaloneCanaryAnalysisInput{ExecutionRequest: er})
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
