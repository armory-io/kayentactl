package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/kayenta-ctl/pkg/kayenta"
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

	resp, err := http.Get(configLocation)
	if err != nil {
		log.Fatal("Could not get default canary config json at locations: " + configLocation)
	}

	kayentaClient := kayenta.NewDefaultClient(
		kayenta.ClientBaseURL(kayentaURL))

	log.Println("updating canary configs")
	_, err = kayenta.UpsertCanaryConfigs(kayentaClient, appName, resp.Body)
	if err != nil {
		log.Println(err)
		log.Fatalf("could not update canary config, exiting")
	}
	log.Println("successfully upserted config")

	//TODO - build canary request based on user inputs
	input, err := generateAnalysisRequest(control, experiment, configLocation)
	if err != nil {
		log.Fatalf("unable to start canary analysis: %s", err.Error())
	}

	// start standalone canary
	log.Println("starting canary analysis")
	output, err := kayentaClient.StartStandaloneCanaryAnalysis(input)
	if err != nil {
		log.Fatalf("error starting canary analysis: %s", err.Error())
	}

	analysisID := output.CanaryAnalysisExecutionID
	log.Println(fmt.Sprintf("canary analysis started with id %s", analysisID))

	// poll until standalone canary is complete
	log.Println("polling until canary analysis is complete")
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	if err := waitForComplete(ctx, analysisID, kayentaClient, checkInterval); err != nil {
		log.Fatalf(err.Error())
	}
	// generate some kind of report
	log.Println("canary analysis complete.")
}

func waitForComplete(ctx context.Context, executionID string, client kayenta.Client, interval time.Duration) error {
	ticker := time.NewTicker(interval)
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

				log.Printf("execution is still running with status %s", res.Status)
			}
		}
	}()
	<-done
	return err1
}

func isComplete(status kayenta.GetStandaloneCanaryAnalysisOutput) bool {
	for _, s := range []string{"canceled", "stopped", "succeeded", "failed_continue", "terminal"} {
		if status.Status == s {
			return true
		}
	}
	return false
}

func generateAnalysisRequest(control, experiment, configLocation string) (kayenta.StandaloneCanaryAnalysisInput, error) {
	emptyResp := kayenta.StandaloneCanaryAnalysisInput{}

	// decide which canary configuration to use
	config := kayenta.CanaryConfig{}
	if configLocation != "" {
		// read in local canary configuration is the user supplied one
		log.Printf("using canary config located at %s", configLocation)
		b, err := ioutil.ReadFile(configLocation)
		if err != nil {
			return emptyResp, err
		}
		var c kayenta.CanaryConfig
		if err := json.NewDecoder(bytes.NewReader(b)).Decode(&c); err != nil {
			return emptyResp, fmt.Errorf("failed to parse canary config: %w", err)
		}
		config = c
	}
	return kayenta.StandaloneCanaryAnalysisInput{
		CanaryConfig: config,
		ExecutionRequest: kayenta.ExecutionRequest{
			Scopes: []kayenta.Scope{
				// TODO - possibly split these to get location i.e. namespace/deploymentName == location, scope
				{ControlScope: control, ExperimentScope: experiment},
			},
			// TODO - decide on defaults here
			LifetimeDurationMins: 60,
			BeginAfterMins:       5,
		},
	}, nil
}
