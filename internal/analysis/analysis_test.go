package analysis

import (
	"testing"

	"github.com/armory-io/kayentactl/pkg/kayenta"

	"github.com/stretchr/testify/assert"
)

func TestConfigureExecRequest(t *testing.T) {
	type testTable struct {
		scope    string
		startIso string
		endIso   string
	}
	tests := []testTable{
		{scope: "MOCK_SCOPE", startIso: "MOCK_START_ISO", endIso: "MOCK_END_ISO"},
		{scope: "MOCK_SCOPE", startIso: "", endIso: ""},
	}
	for _, test := range tests {
		input := []kayenta.Scope{
			{
				ControlScope:    "PLACEHOLDER_CTRL",
				ExperimentScope: "PLACEHOLDER_EXPR",
				StartTimeIso:    "SOME_FAKE_ISO",
				EndTimeIso:      "SOME_FAKE_ISO",
			},
		}

		expectedOutput := []kayenta.Scope{
			{
				ControlScope:    test.scope,
				ExperimentScope: test.scope,
				StartTimeIso:    test.startIso,
				EndTimeIso:      test.endIso,
			},
		}
		output := UpdateScopes(input, test.scope, test.startIso, test.endIso, 10)
		assert.Equal(t, output, expectedOutput)

	}
}
