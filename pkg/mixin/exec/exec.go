package exec

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/davecgh/go-spew/spew"

	"github.com/deislabs/porter/pkg/config"

	"github.com/deislabs/porter/pkg/context"
	"gopkg.in/yaml.v2"
)

// Exec is the logic behind the exec mixin
type Mixin struct {
	*context.Context

	step Step
}

type Step struct {
	Description string              `yaml:"description"`
	Outputs     []config.StepOutput `yaml:"outputs"`
	Instruction Instruction         `yaml:"exec"`
}

type Instruction struct {
	Name       string            `yaml:"name"`
	Command    string            `yaml:"command"`
	Arguments  []string          `yaml:"arguments"`
	Parameters map[string]string `yaml:"parameters"`
}

// New exec mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: context.New(),
	}
}

func (m *Mixin) LoadInstruction(commandFile string) error {
	contents, err := m.getCommandFile(commandFile, m.Out)
	if err != nil {
		return fmt.Errorf("there was an error getting commands: %s", err)
	}
	return yaml.Unmarshal(contents, &m.step)
}

func (m *Mixin) Execute() error {
	cmd := m.NewCommand(m.step.Instruction.Command, m.step.Instruction.Arguments...)
	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	err := cmd.Start()
	if err != nil {
		spew.Dump(m)
		return fmt.Errorf("failed to start...%s", err)
	}

	return cmd.Wait()
}

func (m *Mixin) getCommandFile(commandFile string, w io.Writer) ([]byte, error) {
	if commandFile == "" {
		reader := bufio.NewReader(m.In)
		return ioutil.ReadAll(reader)
	}
	return ioutil.ReadFile(commandFile)
}
