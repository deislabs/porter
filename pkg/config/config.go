package config

import (
	"os"
	"path/filepath"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

const (
	// Name is the file name of the porter configuration file.
	Name = "porter.yaml"

	// RunScript is the path to the CNAB run script.
	RunScript = "cnab/app/run"

	// EnvHOME is the name of the environment variable containing the porter home directory path.
	EnvHOME = "PORTER_HOME"

	// EnvACTION is the request
	EnvACTION = "CNAB_ACTION"
)

type Config struct {
	*context.Context
	Manifest *Manifest
}

// New Config initializes a default porter configuration.
func New() *Config {
	return &Config{
		Context: context.New(),
	}
}

// GetHomeDir determines the path to the porter home directory.
func (c *Config) GetHomeDir() (string, error) {
	home, ok := os.LookupEnv(EnvHOME)
	if ok {
		return home, nil
	}

	porterPath, err := os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "could not get path to the executing porter binary")
	}

	porterDir := filepath.Dir(porterPath)

	return porterDir, nil
}

// GetTemplatesDir determines the path to the templates directory.
func (c *Config) GetTemplatesDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "templates"), nil
}

// GetPorterConfigTemplate reads templates/porter.yaml from the porter home directory.
func (c *Config) GetPorterConfigTemplate() ([]byte, error) {
	tmplDir, err := c.GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	tmplPath := filepath.Join(tmplDir, Name)
	return c.FileSystem.ReadFile(tmplPath)
}

// GetRunScriptTemplate reads templates/run from the porter home directory.
func (c *Config) GetRunScriptTemplate() ([]byte, error) {
	tmplDir, err := c.GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	tmplPath := filepath.Join(tmplDir, filepath.Base(RunScript))
	return c.FileSystem.ReadFile(tmplPath)
}

func (c *Config) GetMixinsDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "mixins"), nil
}

func (c *Config) GetMixinPath(mixin string) (string, error) {
	mixinsDir, err := c.GetMixinsDir()
	if err != nil {
		return "", err
	}

	executablePath := filepath.Join(mixinsDir, mixin, mixin)
	return executablePath, nil
}
