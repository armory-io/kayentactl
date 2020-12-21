package kayenta

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func SkipIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("This is an integraiton test. Skipping test in build/dev environment")
	}
}

func TestUpdateCanaryConfigs(t *testing.T) {
	SkipIntegration(t)
	c := NewDefaultClient(ClientBaseURL("http://localhost:8090"))

	var cc CanaryConfig
	json.Unmarshal([]byte(testConfig), &cc)

	id, err := c.UpdateCanaryConfig(cc)
	assert.Nil(t, err)
	assert.NotEqual(t, id, "")
}

func TestGetCanaryConfigs(t *testing.T) {
	//SkipIntegration(t)
	c := NewDefaultClient(ClientBaseURL("http://localhost:8090"))
	cc, err := c.GetCanaryConfigs("somename")
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
