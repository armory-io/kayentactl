package kayenta

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func SkipIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("This is an integraiton test. Skipping test in build/dev environment")
	}
}

func TestGetCanaryConfigs(t *testing.T) {
	//SkipIntegration(t)
	c := NewDefaultClient(ClientBaseURL("http://localhost:8090"))
	cc, err := c.GetCanaryConfigs()
	assert.Nil(t, err)
	assert.NotNil(t, cc)
	assert.True(t, len(cc) > 0)
}
