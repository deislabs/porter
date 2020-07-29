// +build integration

package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestInvokeCustomAction(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Install a bundle with a custom action defined
	err := p.Create()
	require.NoError(t, err)

	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/bundle-with-custom-action.yaml"), "porter.yaml")
	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/helpers.sh"), "helpers.sh")

	installOpts := porter.InstallOptions{}
	err = installOpts.Validate([]string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(installOpts)
	require.NoError(t, err)

	// Invoke the custom action
	invokeOpts := porter.InvokeOptions{Action: "zombies"}
	err = invokeOpts.Validate([]string{}, p.Porter)
	require.NoError(t, err)
	err = p.InvokeBundle(invokeOpts)
	require.NoError(t, err, "invoke should have succeeded")

	// TODO: We can't check this yet because docker driver is printing directly to stdout instead of to the given writer
	//gotOutput := p.TestConfig.TestContext.GetOutput()
	//assert.Contains(t, gotOutput, "oh noes my brains", "invoke should have printed a cry for halp")

	// Verify that the custom action was recorded properly
	i, err := p.Claims.ReadInstallationStatus(p.Manifest.Name)
	require.NoError(t, err, "could not fetch installation")
	c, err := i.GetLastClaim()
	require.NoError(t, err, "GetLastClaim failed")
	assert.Equal(t, "zombies", c.Action, "the custom action wasn't recorded in the installation")
}
