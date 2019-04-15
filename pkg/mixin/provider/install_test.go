package mixinprovider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_Install(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a mixxin\n")
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	c.SetupPorterHome()
	p := NewFileSystem(c.Config)

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	m, err := p.Install(opts)

	require.NoError(t, err)
	assert.Equal(t, "mixxin", m.Name)
	assert.Equal(t, "/root/.porter/mixins/mixxin", m.Dir)
	assert.Equal(t, "/root/.porter/mixins/mixxin/mixxin", m.ClientPath)

	clientExists, _ := p.FileSystem.Exists("/root/.porter/mixins/mixxin/mixxin")
	assert.True(t, clientExists)
	runtimeExists, _ := p.FileSystem.Exists("/root/.porter/mixins/mixxin/mixxin-runtime")
	assert.True(t, runtimeExists)
}

func TestFileSystem_Install_Rollback(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a mixxin\n")
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config)
	// Hit the real file system for this test because Afero doesn't enforce file permissions and that's how we are
	// sabotaging the install
	c.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}

	// bin is my home now
	binDir := c.TestContext.FindBinDir()
	os.Setenv(config.EnvHOME, binDir)
	defer os.Unsetenv(config.EnvHOME)

	// Make the install fail
	mixinsDir, _ := p.GetMixinsDir()
	mixinDir := path.Join(mixinsDir, "mixxin")
	p.FileSystem.MkdirAll(mixinDir, 0755)
	f, err := p.FileSystem.OpenFile(path.Join(mixinDir, "mixxin"), os.O_CREATE, 0400)
	require.NoError(t, err)
	f.Close()

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	_, err = p.Install(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not create the mixin at")

	// Make sure the mixin directory was removed
	mixinDirExists, _ := p.FileSystem.DirExists(mixinDir)
	assert.False(t, mixinDirExists)
}

func TestFileSystem_Install_RollbackBadDownload(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a mixxin\n")
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config)
	// Hit the real file system for this test because Afero doesn't enforce file permissions and that's how we are
	// sabotaging the install
	c.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}

	// bin is my home now
	binDir := c.TestContext.FindBinDir()
	os.Setenv(config.EnvHOME, binDir)
	defer os.Unsetenv(config.EnvHOME)

	// Make the install fail
	mixinsDir, _ := p.GetMixinsDir()
	mixinDir := path.Join(mixinsDir, "mixxin")
	p.FileSystem.MkdirAll(mixinDir, 0755)
	f, err := p.FileSystem.OpenFile(path.Join(mixinDir, "mixxin"), os.O_CREATE, 0400)
	require.NoError(t, err)
	f.Close()

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	_, err = p.Install(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not create the mixin at")

	// Make sure the mixin directory was removed
	mixinDirExists, _ := p.FileSystem.DirExists(mixinDir)
	assert.False(t, mixinDirExists)
}

func TestFileSystem_Install_RollbackMissingRuntime(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "runtime") {
			w.WriteHeader(400)
		} else {
			fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a client mixxin\n")
		}
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config)

	mixinsDir, _ := p.GetMixinsDir()
	mixinDir := path.Join(mixinsDir, "mixxin")

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	_, err := p.Install(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad status returned when downloading the mixin")

	// Make sure the mixin directory was removed
	mixinDirExists, _ := p.FileSystem.DirExists(mixinDir)
	assert.False(t, mixinDirExists)
}
