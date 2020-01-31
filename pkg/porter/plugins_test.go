package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/instance-storage/claimstore"
	"get.porter.sh/porter/pkg/instance-storage/filesystem"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInternalPluginOpts_Validate(t *testing.T) {
	cfg := config.NewTestConfig(t)
	var opts RunInternalPluginOpts

	t.Run("no key", func(t *testing.T) {
		err := opts.Validate(nil, cfg.Config)
		require.Error(t, err)
		assert.Equal(t, err.Error(), "The positional argument KEY was not specified")
	})

	t.Run("too many keys", func(t *testing.T) {
		err := opts.Validate([]string{"foo", "bar"}, cfg.Config)
		require.Error(t, err)
		assert.Equal(t, err.Error(), "Multiple positional arguments were specified but only one, KEY is expected")
	})

	t.Run("valid key", func(t *testing.T) {
		err := opts.Validate([]string{filesystem.PluginKey}, cfg.Config)
		require.NoError(t, err)
		assert.Equal(t, opts.selectedInterface, claimstore.PluginInterface)
		assert.NotNil(t, opts.selectedPlugin)
	})

	t.Run("invalid key", func(t *testing.T) {
		err := opts.Validate([]string{"foo"}, cfg.Config)
		require.Error(t, err)
		assert.Equal(t, err.Error(), `invalid plugin key specified: "foo"`)
	})
}

func TestPorter_PrintPlugins(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	opts := PrintPluginsOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatTable,
		},
	}
	err := p.PrintPlugins(opts)

	require.Nil(t, err)
	wantOutput := `Name      Type               Implementation   Version   Author
plugin1   instance-storage   blob             v1.0      Deis Labs
plugin1   instance-storage   mongo            v1.0      Deis Labs
plugin2   instance-storage   blob             v1.0      Deis Labs
plugin2   instance-storage   mongo            v1.0      Deis Labs
plugin3   instance-storage   blob             v1.0      Deis Labs
plugin3   instance-storage   mongo            v1.0      Deis Labs
unknown   N/A                N/A              v1.0      Deis Labs
`
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, wantOutput, gotOutput)
}
