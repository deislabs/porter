package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/stretchr/testify/require"
)

func TestNewDriver_Docker(t *testing.T) {
	c := config.NewTestConfig(t)
	claimStorage := storage.NewTestClaimProvider()
	credentialStorage := credentials.NewTestCredentialProvider(t, c)
	d := NewRuntime(c.Config, claimStorage, credentialStorage)

	driver, err := d.newDriver("docker", "myclaim", ActionArguments{})
	require.NoError(t, err)

	if _, ok := driver.(*docker.Driver); ok {
		// TODO: check dockerConfigurationOptions to verify expected bind mount setup,
		// once we're able to (add ability to dockerdriver pkg)
	} else {
		t.Fatal("expected driver to be of type *dockerdriver.Driver")
	}
}
