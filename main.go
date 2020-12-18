package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/armory-io/kayenta-ctl/pkg/kayenta"
)

func main() {
	var (
		configLocation, control, experiment, kayentaURL string
		checkInterval, timeout                          time.Duration
	)
	flag.StringVar(&configLocation, "canary-config", "", "location of the canary config to use")
	flag.StringVar(&control, "control", "", "application to use as the experiment control (i.e. baseline)")
	flag.StringVar(&experiment, "experiment", "", "application to use as the experiment  (i.e. canary)")
	flag.StringVar(&kayentaURL, "url", "http://localhost:8090", "URL for kayenta service")

	flag.DurationVar(&checkInterval, "interval", time.Second*10, "polling interval")
	flag.DurationVar(&timeout, "timeout", 1*time.Hour, "timeout")
	flag.Parse()

	kayentaClient := kayenta.NewDefaultClient(
		kayenta.ClientBaseURL(kayentaURL))

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
	config := defaultCanaryConfig()
	if configLocation != "" {
		// read in local canary configuration is the user supplied one
		log.Printf("using canary config located at %s", configLocation)
		b, err := ioutil.ReadFile(configLocation)
		if err != nil {
			return emptyResp, err
		}
		var c map[string]interface{}
		if err := json.NewDecoder(bytes.NewReader(b)).Decode(&c); err != nil {
			return emptyResp, fmt.Errorf("failed to parse canary config: %w", err)
		}
		config = c
	}

	if config == nil {
		return emptyResp, errors.New("canary config missing one must be supplied to continue")
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

// stub method for supplying a default canary config
func defaultCanaryConfig() map[string]interface{} {
	return nil
}
