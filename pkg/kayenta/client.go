package kayenta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	canaryConfigEndpoint = "/canaryConfig"
)

type CanaryConfig struct {
	Id               string                   `json:"id"`
	Applications     []string                 `json:"canaryConfig"`
	ConfigVersion    string                   `json:"configVersion"`
	CreatedTimestamp int                      `json:"createdTimestamp"`
	Metrics          []map[string]interface{} `json:"metrics"`
}

type StandaloneCanaryAnalysisInput struct {
	// Optional query parameters
	User               string `json:"-"`
	Application        string `json:"-"`
	MetricsAccountName string `json:"-"`
	StorageAccountName string `json:"-"`

	// Request body
	CanaryConfig     CanaryConfig     `json:"canaryConfig"`
	ExecutionRequest ExecutionRequest `json:"executionRequest"`
}

type ExecutionRequest struct {
	Scopes               []Scope `json:"scopes"`
	LifetimeDurationMins int     `json:"lifetimeDurationMins"`
	BeginAfterMins       int     `json:"beginAfterMins"`
}

type Scope struct {
	ScopeName          string `json:"scopeName"`
	ControlScope       string `json:"controlScope"`
	ControlLocation    string `json:"controlLocation"`
	ExperimentScope    string `json:"experimentScope"`
	ExperimentLocation string `json:"experimentLocation"`
	Step               int    `json:"step"`

	// TODO - omitted some propoerties, add if required
}

type StandaloneCanaryAnalysisOutput struct {
	CanaryAnalysisExecutionID string `json:"canaryAnalysisExecutionId"`
}

type GetStandaloneCanaryAnalysisOutput struct {
	Status          string `json:"status"`
	ExecutionStatus string `json:"executionStatus"`
	PipelineID      string `json:"pipelineId"`

	// TODO - there are more things we want here
}

//ServerError is returned whenever there is a problem
type ServerError struct {
	Message string `json:"message"`
	Code    int
}

func (e ServerError) Error() string {
	return fmt.Sprintf("%d : %v", e.Code, e.Message)
}

type Client interface {
	StartStandaloneCanaryAnalysis(input StandaloneCanaryAnalysisInput) (StandaloneCanaryAnalysisOutput, error)
	GetStandaloneCanaryAnalysis(id string) (GetStandaloneCanaryAnalysisOutput, error)
}

// HTTPClientFactory returns an http.Client that
// can be used to make requests and can be used
// to customize the client when needed
type HTTPClientFactory func() *http.Client

func DefaultHTTPClientFactory() *http.Client {
	return &http.Client{}
}

type DefaultClient struct {
	BaseURL       string
	ClientFactory HTTPClientFactory
}

func ClientBaseURL(baseURL string) func(dc *DefaultClient) {
	return func(dc *DefaultClient) {
		dc.BaseURL = baseURL
	}
}

func ClientHTTPClientFactory(factory HTTPClientFactory) func(dc *DefaultClient) {
	return func(dc *DefaultClient) {
		dc.ClientFactory = factory
	}
}

func NewDefaultClient(opts ...func(dc *DefaultClient)) *DefaultClient {
	c := &DefaultClient{
		// TODO: replace with actual kayenta port
		BaseURL:       "http://localhost:8090",
		ClientFactory: DefaultHTTPClientFactory,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (d *DefaultClient) getEndpoint(endpoint string, params map[string]string) string {
	return d.BaseURL + endpoint
}

func (d *DefaultClient) StartStandaloneCanaryAnalysis(input StandaloneCanaryAnalysisInput) (StandaloneCanaryAnalysisOutput, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return StandaloneCanaryAnalysisOutput{}, err
	}
	req, err := http.NewRequest(
		http.MethodPost, d.getEndpoint("/standalone_canary_analysis", nil), bytes.NewReader(b))

	if err != nil {
		return StandaloneCanaryAnalysisOutput{}, err
	}
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return StandaloneCanaryAnalysisOutput{}, err
	}
	var output StandaloneCanaryAnalysisOutput
	if err := deserializeResponse(resp, &output); err != nil {
		return StandaloneCanaryAnalysisOutput{}, err
	}

	return output, nil
}

func (d *DefaultClient) GetStandaloneCanaryAnalysis(id string) (GetStandaloneCanaryAnalysisOutput, error) {
	req, err := http.NewRequest(
		http.MethodPost, d.getEndpoint("/standalone_canary_analysis/"+id, nil), nil)

	if err != nil {
		return GetStandaloneCanaryAnalysisOutput{}, err
	}
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return GetStandaloneCanaryAnalysisOutput{}, err
	}
	var output GetStandaloneCanaryAnalysisOutput
	if err := deserializeResponse(resp, &output); err != nil {
		return GetStandaloneCanaryAnalysisOutput{}, err
	}

	return output, nil
}

//UpdateCanaryConfig updates an existing config
func (d *DefaultClient) UpdateCanaryConfig(configID string, config io.Reader) (string, error) {
	req, err := http.NewRequest(
		http.MethodPut, d.getEndpoint(canaryConfigEndpoint+"/"+configID, nil), config)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", deserializeErrorResponse(resp)
	}
	var result map[string]string
	if err := deserializeResponse(resp, &result); err != nil {
		return "", err
	}
	return result["canaryConfigId"], nil
}

//CreateCanaryConfig writes a canary config to object storage
func (d *DefaultClient) CreateCanaryConfig(config io.Reader) (string, error) {
	req, err := http.NewRequest(
		http.MethodPost, d.getEndpoint(canaryConfigEndpoint, nil), config)
	if err != nil {
		return "", err
	}
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return "", err
	}
	var result map[string]string
	if err := deserializeResponse(resp, &result); err != nil {
		return "", err
	}
	return result["canaryConfigId"], nil
}

//GetCanaryConfigs gets a list of canary configs from the Kayenta server
func (d *DefaultClient) GetCanaryConfigs(application string) ([]CanaryConfig, error) {
	log.Info("Getting Canary Configs")
	req, err := http.NewRequest(
		http.MethodGet, d.getEndpoint(canaryConfigEndpoint, nil), nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return nil, err
	}
	var output []CanaryConfig
	if err := deserializeResponse(resp, &output); err != nil {
		return nil, err
	}
	log.Infof("Found %d canary configs", len(output))
	return output, nil

}

func deserializeErrorResponse(resp *http.Response) error {
	defer resp.Body.Close()
	var e ServerError
	err := json.NewDecoder(resp.Body).Decode(&e)
	if err != nil {
		return err
	}
	e.Code = resp.StatusCode
	return e
}
func deserializeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
