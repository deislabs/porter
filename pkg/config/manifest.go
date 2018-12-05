package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/reflectwalk"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	// path where the manifest was loaded, used to resolve local bundle references
	path string

	Name         string                 `yaml:"image,omitempty"`
	Version      string                 `yaml:"version,omitempty"`
	Image        string                 `yaml:"invocationImage,omitempty"`
	Mixins       []string               `yaml:"mixins,omitempty"`
	Install      Steps                  `yaml:"install"`
	Uninstall    Steps                  `yaml:"uninstall"`
	Parameters   []ParameterDefinition  `yaml:"parameters,omitempty"`
	Credentials  []CredentialDefinition `yaml:"credentials,omitempty"`
	Dependencies []Dependency           `yaml:"dependencies,omitempty"`
}

// ParameterDefinition defines a single parameter for a CNAB bundle
type ParameterDefinition struct {
	Name          string            `yaml:"name"`
	DataType      string            `yaml:"type"`
	DefaultValue  interface{}       `yaml:"default,omitempty"`
	AllowedValues []interface{}     `yaml:"allowed,omitempty"`
	Required      bool              `yaml:"required"`
	MinValue      *int              `yaml:"minValue,omitempty"`
	MaxValue      *int              `yaml:"maxValue,omitempty"`
	MinLength     *int              `yaml:"minLength,omitempty"`
	MaxLength     *int              `yaml:"maxLength,omitempty"`
	Metadata      ParameterMetadata `yaml:"metadata,omitempty"`
	Destination   *Location         `yaml:"destination,omitempty"`
}

type CredentialDefinition struct {
	Name                string `yaml:"name"`
	Path                string `yaml:"path"`
	EnvironmentVariable string `yaml:"env"`
}

type Location struct {
	Path                string `yaml:"path"`
	EnvironmentVariable string `yaml:"env"`
}

// ParameterMetadata contains metadata for a parameter definition.
type ParameterMetadata struct {
	Description string `yaml:"description,omitempty"`
}

type Dependency struct {
	// Stores the results of a bundle after it has been executed
	outputs []BundleOutput

	Name        string             `yaml:"name"`
	Parameters  map[string]string  `yaml:"parameters,omitempty"`
	Connections []BundleConnection `yaml:"connections",omitempty`
}

func (d *Dependency) Validate() error {
	if d.Name == "" {
		return errors.New("dependency name is required")
	}
	return nil
}

type BundleOutput struct {
	Name                string `yaml:"name"`
	Path                string `yaml:"path"`
	EnvironmentVariable string `yaml:"env"`
}

type BundleConnection struct {
	Source      string `yaml:source`
	Destination string `yaml:destination`
	// TODO: Need to add type once it's completed in #20
}

func (c *Config) ReadManifest(path string) (*Manifest, error) {
	data, err := c.FileSystem.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read manifest at %q", path)
	}

	m := &Manifest{}
	err = yaml.Unmarshal(data, m)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse manifest yaml in %q", path)
	}
	m.path = path

	return m, nil
}

func (c *Config) LoadManifest() error {
	return c.LoadManifestFrom(Name)
}

func (c *Config) LoadManifestFrom(file string) error {
	m, err := c.ReadManifest(file)
	if err != nil {
		return err
	}

	c.Manifest = m

	err = c.Manifest.Validate()
	if err != nil {
		return err
	}

	return c.LoadDependencies()
}

// GetManifestDir returns the path to the directory that contains the manifest.
func (m *Manifest) GetManifestDir() string {
	return filepath.Dir(m.path)
}

func (c *Config) LoadDependencies() error {
	for _, dep := range c.Manifest.Dependencies {
		path, err := c.GetBundleManifestPath(dep.Name)
		if err != nil {
			return err
		}

		m, err := c.ReadManifest(path)
		if err != nil {
			return err
		}

		err = m.Validate()
		if err != nil {
			return err
		}

		err = c.Manifest.MergeDependency(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manifest) Validate() error {
	var result error

	if len(m.Mixins) == 0 {
		result = multierror.Append(result, errors.New("no mixins declared"))
	}

	if m.Install == nil {
		result = multierror.Append(result, errors.New("no install action defined"))
	}
	err := m.Install.Validate(m)
	if err != nil {
		result = multierror.Append(result, err)
	}

	if m.Uninstall == nil {
		result = multierror.Append(result, errors.New("no uninstall action defined"))
	}
	err = m.Uninstall.Validate(m)
	if err != nil {
		result = multierror.Append(result, err)
	}

	for _, dep := range m.Dependencies {
		err = dep.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (m *Manifest) GetSteps(action Action) (Steps, error) {
	var steps Steps
	switch action {
	case ActionInstall:
		steps = m.Install
	case ActionUninstall:
		steps = m.Uninstall
	}

	if len(steps) == 0 {
		return nil, errors.Errorf("unsupported action: %q", action)
	}

	return steps, nil
}

func (m *Manifest) MergeDependency(dep *Manifest) error {
	// include any unique credentials from the dependency
	for i, cred := range dep.Credentials {
		dupe := false
		for _, x := range m.Credentials {
			if cred.Name == x.Name {
				result, err := mergeCredentials(x, cred)
				if err != nil {
					return err
				}

				// Allow for having the same credential populated both as an env var and a file
				dep.Credentials[i].EnvironmentVariable = result.EnvironmentVariable
				dep.Credentials[i].Path = result.Path
				dupe = true
				break
			}
		}
		if !dupe {
			m.Credentials = append(m.Credentials, cred)
		}
	}

	// prepend the dependency's mixins
	m.Mixins = prependMixins(dep.Mixins, m.Mixins)

	// prepend dependency's install steps
	m.Install = m.Install.Prepend(dep.Install)

	// append uninstall steps so that we unroll it in dependency order (i.e. uninstall wordpress before we delete the database)
	m.Uninstall = append(m.Uninstall, dep.Uninstall...)

	return nil
}

func prependMixins(m1, m2 []string) []string {
	mixins := make([]string, len(m1), len(m1)+len(m2))
	copy(mixins, m1)
	for _, m := range m2 {
		dupe := false
		for _, x := range m1 {
			if m == x {
				dupe = true
				break
			}
		}
		if !dupe {
			mixins = append(mixins, m)
		}
	}
	return mixins
}

func mergeCredentials(c1, c2 CredentialDefinition) (CredentialDefinition, error) {
	result := CredentialDefinition{Name: c1.Name}

	if c1.Name != c2.Name {
		return result, fmt.Errorf("cannot merge credentials that don't have the same name: %s and %s", c1.Name, c2.Name)
	}

	if c1.Path != "" && c2.Path != "" && c1.Path != c2.Path {
		return result, fmt.Errorf("cannot merge credential %s: conflict on path", c1.Name)
	}
	result.Path = c1.Path
	if result.Path == "" {
		result.Path = c2.Path
	}

	if c1.EnvironmentVariable != "" && c2.EnvironmentVariable != "" && c1.EnvironmentVariable != c2.EnvironmentVariable {
		return result, fmt.Errorf("cannot merge credential %s: conflict on environment variable", c1.Name)
	}
	result.EnvironmentVariable = c1.EnvironmentVariable
	if result.EnvironmentVariable == "" {
		result.EnvironmentVariable = c2.EnvironmentVariable
	}

	return result, nil
}

type Steps []*Step

func (s Steps) Prepend(s1 Steps) Steps {
	result := make(Steps, len(s)+len(s1))
	copy(result[:len(s)], s1)
	copy(result[len(s):], s)
	return result
}

func (s Steps) Validate(m *Manifest) error {
	for _, step := range s {
		err := step.Validate(m)
		if err != nil {
			return err
		}
	}
	return nil
}

type Step struct {
	runner *mixin.Runner

	Description string                 `yaml:"description"`
	Outputs     []StepOutput           `yaml:"outputs"`
	Data        map[string]interface{} `yaml:",inline"`
}

type StepOutput struct {
	Name string                 `yaml:"name"`
	Data map[string]interface{} `yaml:",inline"`
}

func (s *Step) Validate(m *Manifest) error {
	if len(s.Data) == 0 {
		return errors.New("no mixin specified")
	}
	if len(s.Data) > 1 {
		return errors.New("more than one mixin specified")
	}

	mixinDeclared := false
	mixinType := s.GetMixinName()
	for _, mixin := range m.Mixins {
		if mixin == mixinType {
			mixinDeclared = true
			break
		}
	}
	if !mixinDeclared {
		return errors.Errorf("mixin (%s) was not declared", mixinType)
	}

	return nil
}

func (s *Step) GetMixinName() string {
	var mixinName string
	for k := range s.Data {
		mixinName = k
	}
	return mixinName
}

// ResolveStep will walk through the Step's data and resolve any placeholder
// data using the definitions in the manifest, like parameters or credentials.
func (m *Manifest) ResolveStep(step *Step) error {
	return reflectwalk.Walk(step, m)
}

// Map is a NOOP but implements the github.com/mitchellh/reflectwalk MapWalker interface
func (m *Manifest) Map(val reflect.Value) error {
	return nil
}

// MapElem implements the github.com/mitchellh/reflectwalk MapWalker interface and handles
// individual map elements. It will resolve source references to their value within a
// porter bundle and replace the value
func (m *Manifest) MapElem(mp, k, v reflect.Value) error {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	// If the value is is a map, check to see if it's a
	// single entry map with the key "source".
	if kind := v.Kind(); kind == reflect.Map {
		if len(v.MapKeys()) == 1 {
			sk := v.MapKeys()[0]
			if sk.Kind() == reflect.Interface {
				sk = sk.Elem()
			}
			//if the key is a string, and the string is source, then we should try
			//and replace this
			if sk.Kind() == reflect.String && sk.String() == "source" {
				kv := v.MapIndex(sk)
				if kv.Kind() == reflect.Interface {
					kv = kv.Elem()
					value := kv.String()
					replacement, err := m.resolveValue(value)
					if err != nil {
						return errors.Wrap(err, fmt.Sprintf("unable to set value for %s", k.String()))
					}
					mp.SetMapIndex(k, reflect.ValueOf(replacement))
				}
			}
		}
	}
	return nil
}

// Slice is a NOOP but implements the github.com/mitchellh/reflectwalk SliceWalker interface
func (m *Manifest) Slice(val reflect.Value) error {
	return nil
}

// SliceElem implements the github.com/mitchellh/reflectwalk SliceWalker interface and handles
// individual slice elements. It will resolve source references to their value within a
// porter bundle and replace the value
func (m *Manifest) SliceElem(index int, val reflect.Value) error {
	v, ok := val.Interface().(string)
	if ok {
		//if the array entry is a string that matches source:...., we should replace it
		re := regexp.MustCompile("source:\\s?(.*)")
		matches := re.FindStringSubmatch(v)
		if len(matches) > 0 {
			source := matches[1]
			r, err := m.resolveValue(source)
			if err != nil {
				return errors.Wrap(err, "unable to source value")
			}
			val.Set(reflect.ValueOf(r))
		}
	}
	return nil
}

func (m *Manifest) resolveValue(key string) (interface{}, error) {
	source := strings.Split(key, ".")
	var replacement interface{}
	if source[1] == "parameters" {
		for _, param := range m.Parameters {
			if param.Name == source[2] {
				if param.Destination == nil {
					// Porter by default sets CNAB params to name.ToUpper()
					pe := strings.ToUpper(source[2])
					replacement = os.Getenv(pe)
				} else if param.Destination.EnvironmentVariable != "" {
					replacement = os.Getenv(param.Destination.EnvironmentVariable)
				} else if param.Destination == nil && param.Destination.Path != "" {
					replacement = param.Destination.Path
				} else {
					return nil, errors.New(
						"unknown parameter definition, no environment variable or path specified",
					)
				}
			}
		}
	} else if source[1] == "credentials" {
		for _, cred := range m.Credentials {
			if cred.Name == source[2] {
				if cred.Path != "" {
					replacement = cred.Path
				} else if cred.EnvironmentVariable != "" {
					replacement = os.Getenv(cred.EnvironmentVariable)
				} else {
					return nil, errors.New(
						"unknown credential definition, no environment variable or path specified",
					)
				}
			}
		}
	} else {
		return nil, errors.New(
			fmt.Sprintf("unknown source specification: %s", key),
		)
	}
	if replacement == nil {
		return nil, errors.New(fmt.Sprintf("no value found for source specification: %s", key))
	}
	return replacement, nil
}
