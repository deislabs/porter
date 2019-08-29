package builder

import (
	"bytes"
	"encoding/json"

	"github.com/PaesslerAG/jsonpath"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type OutputJsonPath interface {
	Output
	GetJsonPath() string
}

// ProcessJsonPathOutputs evaluates the specified output buffer as JSON, looks through the outputs for
// any that implement the OutputJsonPath and extracts their output.
func ProcessJsonPathOutputs(cxt *context.Context, step StepWithOutputs, outputB *bytes.Buffer) error {
	outputs := step.GetOutputs()

	if len(outputs) == 0 {
		return nil
	}

	if outputB.Len() == 0 {
		return nil
	}

	var outputJson interface{}
	err := json.Unmarshal(outputB.Bytes(), &outputJson)
	if err != nil {
		return errors.Wrapf(err, "error unmarshaling json %s", outputB.String())
	}

	for _, o := range outputs {
		output, ok := o.(OutputJsonPath)
		if !ok {
			continue
		}

		outputPath := output.GetJsonPath()
		outputName := output.GetName()

		value, err := jsonpath.Get(outputPath, outputJson)
		if err != nil {
			return errors.Wrapf(err, "error evaluating jsonpath %q for output %q against %s", outputPath, outputName, outputB.String())
		}

		valueB, err := json.Marshal(value)
		if err != nil {
			return errors.Wrapf(err, "error marshaling jsonpath result %v for output %q", valueB, outputName)
		}

		err = cxt.WriteMixinOutputToFile(outputName, valueB)
		if err != nil {
			return errors.Wrapf(err, "error writing mixin output for %q", outputName)
		}
	}

	return nil
}
