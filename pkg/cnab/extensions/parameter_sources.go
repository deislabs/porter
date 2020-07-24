package extensions

import (
	"encoding/json"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

const (
	// ParameterSourcesExtensionKey represents the full key for the Parameter Sources Extension.
	ParameterSourcesKey = "io.cnab.parameter-sources"
	// ParameterSourcesExtensionSchema represents the schema for the Docker Extension.
	ParameterSourcesSchema = "https://cnab.io/v1/parameter-sources.schema.json"
	// ParameterSourceTypeOutput defines a type of parameter source that is provided by a bundle output.
	ParameterSourceTypeOutput = "output"
)

// ParameterSourcesExtension represents a required extension that specifies how
// to default parameter values.
var ParameterSourcesExtension = RequiredExtension{
	Shorthand: "parameter-sources",
	Key:       ParameterSourcesKey,
	Schema:    ParameterSourcesSchema,
	Reader:    ParameterSourcesReader,
}

// ParameterSources describes the set of custom extension metadata associated
// with the Parameter Sources extension
type ParameterSources map[string]ParameterSource

// SetParameterFromOutput creates an entry in the parameter sources section setting
// the parameter's value using the specified output's value.
func (ps *ParameterSources) SetParameterFromOutput(parameter string, output string) {
	if *ps == nil {
		*ps = ParameterSources{}
	}

	(*ps)[parameter] = ParameterSource{
		Priority: []string{ParameterSourceTypeOutput},
		Sources: ParameterSourceMap{
			ParameterSourceTypeOutput: OutputParameterSource{OutputName: output},
		},
	}
}

type ParameterSource struct {
	// Priority is an array of source types in the priority order that they should be used to
	// populated the parameter.
	Priority []string `json:"priority" mapstructure:"priority"`

	// Sources is a map of key/value pairs of a source type and definition for
	// the parameter value.
	Sources ParameterSourceMap `json:"sources" mapstructure:"sources"`
}

// ListSourcesByPriority returns the parameter sources by the requested priority,
// if none is specified, they are unsorted.
func (s ParameterSource) ListSourcesByPriority() []ParameterSourceDefinition {
	sources := make([]ParameterSourceDefinition, len(s.Sources))
	if len(s.Priority) == 0 {
		for _, source := range s.Sources {
			sources = append(sources, source)
		}
	} else {
		for _, sourceType := range s.Priority {
			sources = append(sources, s.Sources[sourceType])
		}
	}
	return sources
}

type ParameterSourceMap map[string]ParameterSourceDefinition

func (m *ParameterSourceMap) UnmarshalJSON(data []byte) error {
	if *m == nil {
		*m = ParameterSourceMap{}
	}

	var rawMap map[string]interface{}
	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	for sourceKey, sourceDef := range rawMap {
		rawDef, err := json.Marshal(sourceDef)
		if err != nil {
			return errors.Wrapf(err, "error re-marshaling parameter source definition")
		}

		switch sourceKey {
		case ParameterSourceTypeOutput:
			var output OutputParameterSource
			err := json.Unmarshal(rawDef, &output)
			if err != nil {
				return errors.Wrapf(err, "invalid parameter source definition for key output")
			}
			(*m)[ParameterSourceTypeOutput] = output
		default:
			return errors.Errorf("unsupported parameter source key %s", sourceKey)
		}
	}

	return nil
}

type ParameterSourceDefinition interface {
}

// OutputParameterSource represents a parameter that is set using the value from
// a bundle output.
type OutputParameterSource struct {
	OutputName string `json:"name" mapstructure:"name"`
}

// ReadParameterSources is a convenience method for returning a bonafide
// ParameterSources reference after reading from the applicable section from
// the provided bundle
func ReadParameterSources(bun bundle.Bundle) (ParameterSources, error) {
	raw, err := ParameterSourcesReader(bun)
	if err != nil {
		return nil, err
	}

	ps, ok := raw.(ParameterSources)
	if !ok {
		return nil, errors.New("unable to read parameter sources extension data")
	}

	return ps, nil
}

// ParameterSourcesReader is a Reader for the ParameterSourcesExtension,
// which reads from the applicable section in the provided bundle and
// returns a the raw data in the form of an interface
func ParameterSourcesReader(bun bundle.Bundle) (interface{}, error) {
	data, ok := bun.Custom[ParameterSourcesKey]
	if !ok {
		return nil, errors.Errorf("attempted to read parameter sources from bundle but none are defined")
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal the untyped %q extension data %q",
			ParameterSourcesKey, string(dataB))
	}

	ps := ParameterSources{}
	err = json.Unmarshal(dataB, &ps)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal the %q extension %q",
			ParameterSourcesKey, string(dataB))
	}

	return ps, nil
}

// GetParameterSources checks if the parameter sources extension is present and returns its
// extension configuration.
func (e ProcessedExtensions) GetParameterSources() (ParameterSources, bool, error) {
	rawExt, required := e[ParameterSourcesKey]

	ext, ok := rawExt.(ParameterSources)
	if !ok && required {
		return ParameterSources{}, required, errors.Errorf("unable to parse Parameter Sources extension config: %+v", rawExt)
	}

	return ext, required, nil
}

// HasParameterSources returns whether or not the bundle has parameter sources defined.
func HasParameterSources(b bundle.Bundle) bool {
	_, ok := b.Custom[ParameterSourcesKey]
	return ok
}
