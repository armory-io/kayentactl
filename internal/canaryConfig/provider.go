package canaryConfig

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/armory-io/kayentactl/pkg/kayenta"
)

// GetCanaryConfig fetches a canary config from a remote or local source
// and converts it to a StandaloneCanaryAnalysisInput. It supports both YAML
// and JSON data formats
func GetCanaryConfig(location string) (*kayenta.CanaryConfig, error) {
	var configProvider func(string) ([]byte, error)
	if startsWithProtocol(location, []string{"file://", "http://", "https://"}) {
		configProvider = httpConfigProvider
	} else {
		configProvider = ioutil.ReadFile
	}

	b, err := configProvider(location)
	if err != nil {
		return nil, err
	}

	var input kayenta.CanaryConfig
	if err := parseYamlOrJson(b, &input); err != nil {
		return nil, fmt.Errorf("failed to deserialize canary config: %w", err)
	}
	return &input, nil
}

func httpConfigProvider(location string) ([]byte, error) {
	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("attempt to fetch config results in code %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func startsWithProtocol(location string, protocols []string) bool {
	for _, p := range protocols {
		if strings.HasPrefix(location, p) {
			return true
		}
	}
	return false
}

func parseYamlOrJson(b []byte, dest interface{}) error {
	return yaml.Unmarshal(b, dest)
}
