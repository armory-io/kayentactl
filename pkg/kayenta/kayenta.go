package kayenta

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type StandaloneCanaryAnalysisInput map[string]interface{}
type StandaloneCanaryAnalysisOutput map[string]interface{}

type GetStandaloneCanaryAnalysisOutput map[string]interface{}

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
	BaseURL string
	ClientFactory HTTPClientFactory
}

func NewDefaultClient() *DefaultClient {
	return &DefaultClient{
		// TODO: replace with actual kayenta port
		BaseURL:       "http://localhost:9999",
		ClientFactory: DefaultHTTPClientFactory,
	}
}

func (d *DefaultClient) getEndpoint(endpoint string) string {
	return ""
}

func (d *DefaultClient) StartStandaloneCanaryAnalysis(input StandaloneCanaryAnalysisInput) (StandaloneCanaryAnalysisOutput, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return StandaloneCanaryAnalysisOutput{}, err
	}
	req, err := http.NewRequest(
		http.MethodPost, d.getEndpoint("/standalone_canary_analysis"), bytes.NewReader(b))

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
		http.MethodPost, d.getEndpoint("/standalone_canary_analysis/" + id), nil)

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

func deserializeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
