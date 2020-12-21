package kayenta

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpsertCanaryConfigs(t *testing.T) {

	c := NewDefaultClient(ClientBaseURL("http://localhost:8090"))
	var cc CanaryConfig
	json.Unmarshal([]byte(testConfig), &cc)
	UpsertCanaryConfigs(c, "somename", cc)
	assert.Equal(t, cc.Id, "hello")
}
