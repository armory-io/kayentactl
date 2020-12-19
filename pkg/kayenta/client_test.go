package kayenta

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func SkipIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("This is an integraiton test. Skipping test in build/dev environment")
	}
}

func TestUpdateCanaryConfigs(t *testing.T) {
	//SkipIntegration(t)
	c := NewDefaultClient(ClientBaseURL("http://localhost:8090"))
	cc, err := c.UpdateCanaryConfig("1234", strings.NewReader(testConfig))
	assert.Nil(t, err)
	assert.NotEqual(t, cc, "")
}

func TestGetCanaryConfigs(t *testing.T) {
	SkipIntegration(t)
	c := NewDefaultClient(ClientBaseURL("http://localhost:8090"))
	cc, err := c.GetCanaryConfigs("someapps")
	assert.Nil(t, err)
	assert.NotNil(t, cc)
	assert.True(t, len(cc) > 0)
}

const testConfig string = `{
	"applications": [
	  "beats"
	],
	"metrics": [
	  {
		"analysisConfigurations": {
		  "canary": {
			"critical": true,
			"direction": "either"
		  }
		},
		"groups": [
		  "Group 1"
		],
		"name": "Latency",
		"query": {
		  "customFilterTemplate": "filter",
		  "customInlineTemplate": "",
		  "groupByFields": [],
		  "metricName": "custom_dummy_latency",
		  "serviceType": "prometheus",
		  "type": "prometheus"
		},
		"scopeName": "default"
	  }
	],
	"name": "hello-world2"
  }`
