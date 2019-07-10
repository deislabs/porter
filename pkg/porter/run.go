package porter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
)

type RunOptions struct {
	config *config.Config

	File         string
	Action       string
	parsedAction config.Action
}

func NewRunOptions(c *config.Config) RunOptions {
	return RunOptions{
		config: c,
	}
}

func (o *RunOptions) Validate() error {
	err := o.defaultDebug()
	if err != nil {
		return err
	}

	err = o.validateAction()
	if err != nil {
		return err
	}

	return nil
}

func (o *RunOptions) validateAction() error {
	if o.Action == "" {
		o.Action = os.Getenv(config.EnvACTION)
		if o.config.Debug {
			fmt.Fprintf(o.config.Err, "DEBUG: defaulting action to %s (%s)\n", config.EnvACTION, o.Action)
		}
	}

	var err error
	o.parsedAction, err = config.ParseAction(o.Action)
	return err
}

func (o *RunOptions) defaultDebug() error {
	// if debug was manually set, leave it
	if o.config.Debug {
		return nil
	}

	rawDebug, set := os.LookupEnv(config.EnvDEBUG)
	if !set {
		return nil
	}

	debug, err := strconv.ParseBool(rawDebug)
	if err != nil {
		return errors.Wrapf(err, "invalid PORTER_DEBUG, expected a bool (true/false) but got %s", rawDebug)
	}

	if debug {
		fmt.Fprintf(o.config.Err, "DEBUG: defaulting debug to %s (%t)\n", config.EnvDEBUG, debug)
		o.config.Debug = debug
	}

	return nil
}

func (p *Porter) Run(opts RunOptions) error {
	fmt.Fprintf(p.Out, "executing porter %s configuration from %s\n", opts.parsedAction, opts.File)

	err := p.Config.LoadManifestFrom(opts.File)
	if err != nil {
		return err
	}

	steps, err := p.Manifest.GetSteps(opts.parsedAction)
	if err != nil {
		return err
	}

	err = steps.Validate(p.Manifest)
	if err != nil {
		return errors.Wrap(err, "invalid action configuration")
	}

	mixinsDir, err := p.GetMixinsDir()
	if err != nil {
		return err
	}

	err = p.FileSystem.MkdirAll(mixin.OutputsDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not create outputs directory %s", mixin.OutputsDir)
	}

	for _, step := range steps {
		if step != nil {
			err := p.Manifest.ResolveStep(step)
			if err != nil {
				return errors.Wrap(err, "unable to resolve sourced values")
			}
			// Hand over values needing masking in context output streams
			p.Context.SetSensitiveValues(p.Manifest.GetSensitiveValues())

			runner := p.loadRunner(step, opts.parsedAction, mixinsDir)

			err = runner.Validate()
			if err != nil {
				return errors.Wrap(err, "mixin validation failed")
			}

			description, _ := step.GetDescription()
			fmt.Fprintln(p.Out, description)
			err = runner.Run()
			if err != nil {
				return errors.Wrap(err, "mixin execution failed")
			}

			outputs, err := p.readOutputs()
			if err != nil {
				return errors.Wrap(err, "could not read step outputs")
			}

			err = p.Manifest.ApplyStepOutputs(step, outputs)
			if err != nil {
				return err
			}

			// Apply any Bundle Outputs declared in this step
			err = p.ApplyBundleOutputs(opts, outputs)
			if err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(p.Out, "execution completed successfully!")
	return nil
}

type ActionInput struct {
	action config.Action
	Steps  []*config.Step `yaml:"steps"`
}

// MarshalYAML marshals the step nested under the action
// install:
// - helm:
//   ...
// Solution from https://stackoverflow.com/a/42547226
func (a *ActionInput) MarshalYAML() (interface{}, error) {
	// encode the original
	b, err := yaml.Marshal(a.Steps)
	if err != nil {
		return nil, err
	}

	// decode it back to get a map
	var tmp interface{}
	err = yaml.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	stepMap := tmp.([]interface{})
	actionMap := map[string]interface{}{
		string(a.action): stepMap,
	}
	return actionMap, nil
}

func (p *Porter) loadRunner(s *config.Step, action config.Action, mixinsDir string) *mixin.Runner {
	name := s.GetMixinName()
	mixinDir := filepath.Join(mixinsDir, name)

	r := mixin.NewRunner(name, mixinDir, true)
	r.Command = string(action)
	r.Context = p.Context

	input := &ActionInput{
		action: action,
		Steps:  []*config.Step{s},
	}
	inputBytes, _ := yaml.Marshal(input)
	r.Input = string(inputBytes)

	return r
}

func (p *Porter) readOutputs() ([]string, error) {
	var outputs []string

	outfiles, err := p.FileSystem.ReadDir(mixin.OutputsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list %s", mixin.OutputsDir)
	}

	for _, outfile := range outfiles {
		if outfile.IsDir() {
			continue
		}

		outpath := filepath.Join(mixin.OutputsDir, outfile.Name())
		contents, err := p.FileSystem.ReadFile(outpath)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read output file %s", outpath)
		}

		for _, line := range strings.Split(string(contents), "\n") {
			// remove any empty lines from the split process
			if len(line) > 0 {
				outputs = append(outputs, line)
			}
		}
		err = p.FileSystem.Remove(outpath)
		if err != nil {
			return nil, err
		}
	}

	return outputs, nil
}

// ApplyBundleOutputs writes the provided outputs to the proper location
// in the execution environment
func (p *Porter) ApplyBundleOutputs(opts RunOptions, outputs []string) error {
	for _, output := range outputs {
		// Iterate through bundle outputs declared in the manifest
		for _, bundleOutput := range p.Manifest.Outputs {
			// Currently, outputs are all transfered in one file, delimited by newlines
			// We therefore have to check if a given output (line) corresponds to this bundle output
			// TODO: refactor once outputs are transferred in the form of files
			outputSplit := strings.SplitN(output, "=", 2)
			outputKey := outputSplit[0]
			outputValue := outputSplit[1]

			// If a given step output matches a bundle output, proceed
			if outputKey == bundleOutput.Name {
				doApply := true

				// If ApplyTo array non-empty, default doApply to false
				// and only set to true if at least one entry matches current Action
				if len(bundleOutput.ApplyTo) > 0 {
					doApply = false
					for _, applyTo := range bundleOutput.ApplyTo {
						if opts.Action == applyTo {
							doApply = true
						}
					}
				}

				if doApply {
					// Ensure outputs directory exists
					if err := p.FileSystem.MkdirAll(config.BundleOutputsDir, 0755); err != nil {
						return errors.Wrap(err, "unable to ensure CNAB outputs directory exists")
					}

					outpath := filepath.Join(config.BundleOutputsDir, bundleOutput.Name)

					// Create data structure with relevant data for use in listing/showing later
					output := Output{
						Name:      bundleOutput.Name,
						Sensitive: bundleOutput.Sensitive,
						Type:      bundleOutput.Type,
						Value:     outputValue,
					}

					data, err := json.MarshalIndent(output, "", "  ")
					if err != nil {
						return err
					}

					err = p.FileSystem.WriteFile(outpath, data, 0755)
					if err != nil {
						return errors.Wrapf(err, "unable to write output file %s", outpath)
					}
				}
			}
		}
	}
	return nil
}
