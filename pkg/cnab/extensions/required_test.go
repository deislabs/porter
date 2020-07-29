package extensions

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/require"
)

func TestProcessRequiredExtensions(t *testing.T) {
	t.Run("supported", func(t *testing.T) {
		bun := cnab.ReadTestBundle(t, "testdata/bundle.json")
		exts, err := ProcessRequiredExtensions(bun)
		require.NoError(t, err, "could not process required extensions")

		expected := ProcessedExtensions{
			"io.cnab.dependencies": Dependencies{
				Requires: map[string]Dependency{
					"storage": Dependency{
						Bundle: "somecloud/blob-storage",
					},
					"mysql": Dependency{
						Bundle: "somecloud/mysql",
						Version: &DependencyVersion{
							AllowPrereleases: true,
							Ranges:           []string{"5.7.x"},
						},
					},
				},
			},
			"io.cnab.parameter-sources": ParameterSources{
				"tfstate": ParameterSource{
					Priority: []string{ParameterSourceTypeOutput},
					Sources: ParameterSourceMap{
						ParameterSourceTypeOutput: OutputParameterSource{"tfstate"},
					},
				},
			},
		}
		require.Equal(t, expected, exts)
	})

	t.Run("supported unprocessable", func(t *testing.T) {
		bun := cnab.ReadTestBundle(t, "testdata/bundle-supported-unprocessable.json")
		_, err := ProcessRequiredExtensions(bun)
		require.EqualError(t, err, "unable to process extension: io.cnab.docker: no custom extension configuration found")
	})

	t.Run("unsupported", func(t *testing.T) {
		bun := cnab.ReadTestBundle(t, "testdata/bundle-unsupported-required.json")
		_, err := ProcessRequiredExtensions(bun)
		require.EqualError(t, err, "unsupported required extension: donuts")
	})
}

func TestGetSupportedExtension(t *testing.T) {
	for _, supported := range SupportedExtensions {
		t.Run(fmt.Sprintf("%s - shorthand", supported.Shorthand), func(t *testing.T) {
			ext, err := GetSupportedExtension(supported.Shorthand)
			require.NoError(t, err)
			require.Equal(t, supported.Key, ext.Key)
		})

		t.Run(fmt.Sprintf("%s - key", supported.Key), func(t *testing.T) {
			ext, err := GetSupportedExtension(supported.Key)
			require.NoError(t, err)
			require.Equal(t, supported.Key, ext.Key)
		})
	}

	t.Run("unsupported", func(t *testing.T) {
		_, err := GetSupportedExtension("donuts")
		require.EqualError(t, err, "unsupported required extension: donuts")
	})
}
