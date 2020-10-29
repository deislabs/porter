package datastore

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromConfigFile(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetHomeDir("/root/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

	c.DataLoader = FromConfigFile
	err := c.LoadData()
	require.NoError(t, err, "dataloader failed")
	require.NotNil(t, c.Data, "config.Data was not populated")
	assert.True(t, c.Debug, "config.Debug was not set correctly")
}

func TestFromFlagsThenEnvVarsThenConfigFile(t *testing.T) {
	buildCommand := func(c *config.Config) *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().BoolVar(&c.Debug, "debug", false, "debug")
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return c.LoadData()
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
		c.DataLoader = FromFlagsThenEnvVarsThenConfigFile(cmd)
		return cmd
	}

	t.Run("no flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()
		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.False(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.True(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug flag overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug=false"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.False(t, c.Debug, "config.Debug should have been set by the flag and not the config")
	})

	t.Run("debug env var", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		os.Setenv("PORTER_DEBUG", "true")
		defer os.Unsetenv("PORTER_DEBUG")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.True(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug env var overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		os.Setenv("PORTER_DEBUG", "false")
		defer os.Unsetenv("PORTER_DEBUG")
		c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.False(t, c.Debug, "config.Debug should have been set by the env var and not the config")
	})

	t.Run("flag overrides debug env var overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		os.Setenv("PORTER_DEBUG", "false")
		defer os.Unsetenv("PORTER_DEBUG")
		c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug", "true"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.True(t, c.Debug, "config.Debug should have been set by the flag and not the env var or config")
	})
}

func TestData_Marshal(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetHomeDir("/root/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

	c.DataLoader = FromConfigFile
	err := c.LoadData()
	require.NoError(t, err, "LoadData failed")

	require.NotNil(t, c.Data, "Data was not populated by LoadData")

	// Check Storage Attributes
	assert.Equal(t, "dev", c.Data.DefaultStorage, "DefaultStorage was not loaded properly")
	assert.Equal(t, "azure.blob", c.Data.DefaultStoragePlugin, "DefaultStoragePlugin was not loaded properly")

	require.Len(t, c.Data.CrudStores, 1, "CrudStores was not loaded properly")
	devStore := c.Data.CrudStores[0]
	assert.Equal(t, "dev", devStore.Name, "CrudStores.Name was not loaded properly")
	assert.Equal(t, "azure.blob", devStore.PluginSubKey, "CrudStores.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"env": "DEV_AZURE_STORAGE_CONNECTION_STRING"}, devStore.Config, "CrudStores.Config was not loaded properly")

	// Check Secret Attributes
	assert.Equal(t, "red-team", c.Data.DefaultSecrets, "DefaultSecrets was not loaded properly")
	assert.Equal(t, "azure.keyvault", c.Data.DefaultSecretsPlugin, "DefaultSecretsPlugin was not loaded properly")

	require.Len(t, c.Data.SecretSources, 1, "SecretSources was not loaded properly")
	teamSource := c.Data.SecretSources[0]
	assert.Equal(t, "red-team", teamSource.Name, "SecretSources.Name was not loaded properly")
	assert.Equal(t, "azure.keyvault", teamSource.PluginSubKey, "SecretSources.PluginSubKey was not loaded properly")
	assert.Equal(t, map[string]interface{}{"vault": "teamsekrets"}, teamSource.Config, "SecretSources.Config was not loaded properly")
}
