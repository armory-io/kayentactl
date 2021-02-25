package kayenta

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	canaryConfigEndpoint             = "/canaryConfig"
	standaloneCanaryAnalysisEndpoint = "/standalone_canary_analysis"
	credentialsEndpoint              = "/credentials"
)

//StandaloneCanaryAnalysisInput is used to create an api request to kayenta for a standalone analysis
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

type CanaryConfig struct {
	Name             string           `json:"name"`
	Id               string           `json:"id"`
	Applications     []string         `json:"canaryConfig"`
	ConfigVersion    string           `json:"configVersion"`
	CreatedTimestamp int              `json:"createdTimestamp"`
	Judge            JudgeConfig      `json:"judge"`
	Metrics          []Metric         `json:"metrics"`
	Classifier       CanaryClassifier `json:"classifier"`
}

type JudgeConfig struct {
	Name string `json:"name"`
}

type Metric struct {
	Groups    []string          `json:"groups"`
	Name      string            `json:"name"`
	Query     map[string]string `json:"query"`
	ScopeName string            `json:"scopeName"`

	AnalysisConfigurations AnalysisConfiguration `json:"analysisConfigurations"`
}

type AnalysisConfiguration map[string]interface{}

type CanaryClassifier struct {
	GroupWeights map[string]int `json:"groupWeights"`
}

type ExecutionRequest struct {
	Scopes               []Scope `json:"scopes"`
	LifetimeDurationMins int     `json:"lifetimeDurationMins"`
	BeginAfterMins       int     `json:"beginAfterMins"`
	AnalysisIntervalMins int     `json:"analysisIntervalMins"`

	Thresholds Threshold `json:"thresholds"`
}

type Threshold struct {
	Marginal string `json:"marginal"`
	Pass     string `json:"pass"`
}

type Scope struct {
	ScopeName              string `json:"scopeName"`
	ControlScope           string `json:"controlScope"`
	ControlLocation        string `json:"controlLocation"`
	ControlOffsetInMinutes int    `json:"controlOffsetInMinutes"`
	ExperimentScope        string `json:"experimentScope"`
	ExperimentLocation     string `json:"experimentLocation"`
	Step                   int    `json:"step"`

	StartTimeIso string `json:"startTimeIso,omitempty"`
	EndTimeIso   string `json:"endTimeIso,omitempty"`

	ExtendedScopeParams map[string]string `json:"extendedScopeParams"`

	// TODO - omitted some propoerties, add if required
}

type StandaloneCanaryAnalysisOutput struct {
	CanaryAnalysisExecutionID string `json:"canaryAnalysisExecutionId"`
}

type StageStatus struct {
	StageType   string `json:"type"`
	Name        string `json:"name"`
	Status      string `json:status"`
	ExecutionID string `json:executionId"`
}
type GetStandaloneCanaryAnalysisOutput struct {
	Status                        string                        `json:"status"`
	ExecutionStatus               string                        `json:"executionStatus"`
	PipelineID                    string                        `json:"pipelineId"`
	Complete                      bool                          `json:"complete"`
	Stages                        []StageStatus                 `json:"stageStatus"`
	CanaryAnalysisExecutionResult CanaryAnalysisExecutionResult `json:"canaryAnalysisExecutionResult"`

	// TODO - there are more things we want here
}

func (g GetStandaloneCanaryAnalysisOutput) IsSuccessful() bool {
	return g.Status == "succeeded"
}

type CanaryAnalysisExecutionResult struct {
	DidPassThresholds      bool                    `json:"didPassThresholds"`
	HasWarnings            bool                    `json:"hasWarnings"`
	CanaryScoreMessage     string                  `json:"canaryScoreMessage"`
	CanaryScores           []float64               `json:"canaryScores"`
	CanaryExecutionResults []CanaryExecutionResult `json:"canaryExecutionResults"`
}

type CanaryExecutionResult struct {
	Result struct {
		JudgeResult    JudgeResult `json:"judgeResult"`
		CanaryDuration string      `json:"canaryDuration"`
	} `json:"result"`
}

type JudgeResult struct {
	JudgeName string `json:"judgeName"`
	Results   []struct {
		Name                 string   `json:"name"`
		Classification       string   `json:"classification"`
		ClassificationReason string   `json:"classificationReason"`
		Groups               []string `json:"groups"`
	} `json:"results"`
	GroupScores []MetricGroup `json:"groupScores"`
}

type MetricGroup struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

//ServerError is returned whenever there is a problem
type ServerError struct {
	Message string `json:"message"`
	Code    int
}

func (e ServerError) Error() string {
	msg := e.Message
	if e.Message == "" {
		msg = "no message included in response from server"
	}
	return fmt.Sprintf("%d : %v", e.Code, msg)
}

type CanaryConfigAPI interface {
	UpdateCanaryConfig(cc CanaryConfig) (string, error)
	CreateCanaryConfig(cc CanaryConfig) (string, error)
	GetCanaryConfigs(application string) ([]CanaryConfig, error)
}
type StandaloneCanaryAnalysisAPI interface {
	StartStandaloneCanaryAnalysis(input StandaloneCanaryAnalysisInput) (StandaloneCanaryAnalysisOutput, error)
	GetStandaloneCanaryAnalysis(id string) (GetStandaloneCanaryAnalysisOutput, error)
}

type CredentialsAPI interface {
	GetCredentials() []AccountCredential
}

type AccountCredential struct {
	Name           string   `json:"name"`
	SupportedTypes []string `json:"supportedTypes"`
	Type           string   `json:"type"`
}

type Client interface {
	StandaloneCanaryAnalysisAPI
	CanaryConfigAPI
	CredentialsAPI
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
	u := d.BaseURL + endpoint
	// TODO - handle error, who cares for now
	parsed, _ := url.Parse(u)
	q := parsed.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func requestFactory(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}
	req.Header.Add("Content-Type", "application/json")
	return req, err

}

//StartStandaloneCanaryAnalysis - starts a canary analysis
func (d *DefaultClient) StartStandaloneCanaryAnalysis(input StandaloneCanaryAnalysisInput) (StandaloneCanaryAnalysisOutput, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return StandaloneCanaryAnalysisOutput{}, fmt.Errorf("failed to marshal request input: %w", err)
	}
	startQueryParams := map[string]string{
		"storageAccountName": input.StorageAccountName,
		"metricsAccountName": input.MetricsAccountName,
		// TODO - there are still some params missing from this
	}

	req, err := requestFactory(
		http.MethodPost, d.getEndpoint(standaloneCanaryAnalysisEndpoint, startQueryParams), bytes.NewReader(b))

	if err != nil {
		return StandaloneCanaryAnalysisOutput{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return StandaloneCanaryAnalysisOutput{}, fmt.Errorf("failed to execute request: %w", err)
	}
	if resp.StatusCode >= 400 {
		return StandaloneCanaryAnalysisOutput{}, deserializeErrorResponse(resp)
	}

	var output StandaloneCanaryAnalysisOutput
	if err := deserializeResponse(resp, &output); err != nil {
		return StandaloneCanaryAnalysisOutput{}, fmt.Errorf("error deserializing response: %w", err)
	}

	return output, nil
}

func (d *DefaultClient) GetStandaloneCanaryAnalysis(id string) (GetStandaloneCanaryAnalysisOutput, error) {
	req, err := requestFactory(
		http.MethodGet, d.getEndpoint(standaloneCanaryAnalysisEndpoint+"/"+id, nil), nil)

	if err != nil {
		return GetStandaloneCanaryAnalysisOutput{}, err
	}
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return GetStandaloneCanaryAnalysisOutput{}, err
	}
	if resp.StatusCode >= 400 {
		return GetStandaloneCanaryAnalysisOutput{}, deserializeErrorResponse(resp)
	}

	var output GetStandaloneCanaryAnalysisOutput
	if err := deserializeResponse(resp, &output); err != nil {
		return GetStandaloneCanaryAnalysisOutput{}, err
	}

	return output, nil
}

//UpdateCanaryConfig updates an existing config
func (d *DefaultClient) UpdateCanaryConfig(cc CanaryConfig) (string, error) {
	if cc.Id == "" {
		return "", errors.New("Canary Config ID cannot be empty value")
	}
	ccBytes, err := json.Marshal(cc)
	if err != nil {
		return "", fmt.Errorf("could not marshal canary config: %w", err)
	}

	req, err := requestFactory(
		http.MethodPut, d.getEndpoint(canaryConfigEndpoint+"/"+cc.Id, nil), bytes.NewReader(ccBytes))
	if err != nil {
		return "", err
	}
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
func (d *DefaultClient) CreateCanaryConfig(cc CanaryConfig) (string, error) {
	ccBytes, err := json.Marshal(cc)
	if err != nil {
		return "", fmt.Errorf("could not marshal canary config: %w", err)
	}
	req, err := requestFactory(
		http.MethodPost, d.getEndpoint(canaryConfigEndpoint, nil), bytes.NewReader(ccBytes))
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
	return output, nil

}

func (d *DefaultClient) GetCredentials() ([]AccountCredential, error) {
	req, err := http.NewRequest(
		http.MethodGet, d.getEndpoint(credentialsEndpoint, nil), nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.ClientFactory().Do(req)
	if err != nil {
		return nil, err
	}
	var output []AccountCredential
	if err := deserializeResponse(resp, &output); err != nil {
		return nil, err
	}
	return output, nil

}

func deserializeErrorResponse(resp *http.Response) error {
	var e ServerError
	if err := deserializeResponse(resp, &e); err != nil {
		return err
	}
	e.Code = resp.StatusCode
	return e
}
func deserializeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	if resp.Body == http.NoBody {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
