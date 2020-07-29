package cnabprovider

import (
	"testing"
	"time"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_loadCredentials(t *testing.T) {
	r := NewTestRuntime(t)

	r.TestCredentials.TestSecrets.AddSecret("password", "mypassword")
	r.TestCredentials.TestSecrets.AddSecret("db-password", "topsecret")

	r.TestConfig.TestContext.AddTestFile("testdata/db-creds.json", "/db-creds.json")

	cs1 := credentials.CredentialSet{
		Name:     "mycreds",
		Created:  time.Now(),
		Modified: time.Now(),
		Credentials: []valuesource.Strategy{
			{
				Name: "password",
				Source: valuesource.Source{
					Key:   secrets.SourceSecret,
					Value: "password",
				},
			},
		},
	}
	err := r.credentials.Save(cs1)
	require.NoError(t, err, "Save credential set failed")

	b := bundle.Bundle{
		Credentials: map[string]bundle.Credential{
			"password": {
				Location: bundle.Location{
					EnvironmentVariable: "PASSWORD",
				},
			},
			"db-password": {
				Location: bundle.Location{
					EnvironmentVariable: "DB_PASSWORD",
				},
			},
		},
	}

	gotValues, err := r.loadCredentials(b, []string{"mycreds", "/db-creds.json"})
	require.NoError(t, err, "loadCredentials failed")

	wantValues := valuesource.Set{
		"password":    "mypassword",
		"db-password": "topsecret",
	}
	assert.Equal(t, wantValues, gotValues, "resolved unexpected credential values")
}
