package porter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	yaml "gopkg.in/yaml.v2"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_readOutputs(t *testing.T) {
	p := NewTestPorter(t)

	type readOutputsTest struct {
		Name    string
		Remove  bool
		FailMsg string
	}

	testcases := []readOutputsTest{
		{
			Name:    "remove: true",
			Remove:  true,
			FailMsg: "files should not exist after reading outputs",
		},
		{
			Name:    "remove: false",
			Remove:  false,
			FailMsg: "files should exist after reading outputs",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			p.TestConfig.TestContext.AddTestFile("testdata/outputs1.txt", filepath.Join(mixin.OutputsDir, "myoutput1"))
			p.TestConfig.TestContext.AddTestFile("testdata/outputs2.txt", filepath.Join(mixin.OutputsDir, "myoutput2"))

			gotOutputs, err := p.readOutputs(tc.Remove)
			require.NoError(t, err)

			for _, file := range []string{filepath.Join(mixin.OutputsDir, "myoutput1"), filepath.Join(mixin.OutputsDir, "myoutput2")} {
				if exists, _ := p.FileSystem.Exists(file); exists == tc.Remove {
					require.Fail(t, tc.FailMsg)
				}
			}

			wantOutputs := []string{
				"FOO=BAR",
				"BAZ=QUX",
				"A=B",
			}
			assert.Equal(t, wantOutputs, gotOutputs)
		})
	}
}

func TestPorter_defaultDebugToOff(t *testing.T) {
	p := New() // Don't use the test porter, it has debug on by default
	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)
	assert.False(t, p.Config.Debug)
}

func TestPorter_defaultDebugUsesEnvVar(t *testing.T) {
	os.Setenv(config.EnvDEBUG, "true")
	defer os.Unsetenv(config.EnvDEBUG)

	p := New() // Don't use the test porter, it has debug on by default
	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)

	assert.True(t, p.Config.Debug)
}

func TestActionInput_MarshalYAML(t *testing.T) {
	s := &config.Step{
		Data: map[string]interface{}{
			"exec": map[string]interface{}{
				"command": "echo hi",
			},
		},
	}

	input := &ActionInput{
		action: config.ActionInstall,
		Steps:  []*config.Step{s},
	}

	b, err := yaml.Marshal(input)
	require.NoError(t, err)
	wantYaml := "install:\n- exec:\n    command: echo hi\n"
	assert.Equal(t, wantYaml, string(b))
}
