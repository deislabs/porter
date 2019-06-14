package porter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	cxt "github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/pkg/errors"
)

type BuildOptions struct {
	contextOptions
}

func (p *Porter) Build(opts BuildOptions) error {
	opts.Apply(p.Context)

	err := p.LoadManifest()
	if err != nil {
		return err
	}

	if err := p.prepareDockerFilesystem(); err != nil {
		return fmt.Errorf("unable to copy mixins: %s", err)
	}
	if err := p.generateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}
	err = p.buildInvocationImage(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to build CNAB invocation image")
	}

	return p.buildBundle(p.Config.Manifest.Image, "")
}

func (p *Porter) generateDockerFile() error {
	lines, err := p.buildDockerfile()
	if err != nil {
		return errors.Wrap(err, "error generating the Dockerfile")
	}

	fmt.Fprintf(p.Out, "\nWriting Dockerfile =======>\n")
	contents := strings.Join(lines, "\n")

	if p.IsVerbose() {
		fmt.Fprintln(p.Out, contents)
	}

	err = p.Config.FileSystem.WriteFile("Dockerfile", []byte(contents), 0644)
	return errors.Wrap(err, "couldn't write the Dockerfile")
}

func (p *Porter) buildDockerfile() ([]string, error) {
	fmt.Fprintf(p.Out, "\nGenerating Dockerfile =======>\n")

	lines, err := p.getBaseDockerfile()
	if err != nil {
		return nil, err
	}

	mixinLines, err := p.buildMixinsSection()
	if err != nil {
		return nil, errors.Wrap(err, "error generating Dockefile content for mixins")
	}
	lines = append(lines, mixinLines...)

	// The template dockerfile copies everything by default, but if the user
	// supplied their own, copy over cnab/ and porter.yaml
	if p.Manifest.Dockerfile != "" {
		lines = append(lines, p.buildCNABSection()...)
		lines = append(lines, p.buildPorterSection()...)
	}
	lines = append(lines, p.buildWORKDIRSection())
	lines = append(lines, p.buildCMDSection())

	if p.IsVerbose() {
		for _, line := range lines {
			fmt.Fprintln(p.Out, line)
		}
	}

	return lines, nil
}

func (p *Porter) getBaseDockerfile() ([]string, error) {
	var reader io.Reader
	if p.Manifest.Dockerfile != "" {
		exists, err := p.FileSystem.Exists(p.Manifest.Dockerfile)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking if Dockerfile exists: %q", p.Manifest.Dockerfile)
		}
		if !exists {
			return nil, errors.Errorf("the Dockerfile specified in the manifest doesn't exist: %q", p.Manifest.Dockerfile)
		}

		file, err := p.FileSystem.Open(p.Manifest.Dockerfile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file

	} else {
		contents, err := p.Templates.GetDockerfile()
		if err != nil {
			return nil, errors.Wrap(err, "error loading default Dockerfile template")
		}
		reader = bytes.NewReader(contents)
	}

	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func (p *Porter) buildPorterSection() []string {
	return []string{
		`COPY porter.yaml /cnab/app/porter.yaml`,
	}
}

func (p *Porter) buildCNABSection() []string {
	return []string{
		`COPY cnab/ /cnab/`,
	}
}

func (p *Porter) buildWORKDIRSection() string {
	return `WORKDIR /cnab/app`
}

func (p *Porter) buildCMDSection() string {
	return `CMD ["/cnab/app/run"]`
}

func (p *Porter) buildMixinsSection() ([]string, error) {
	lines := make([]string, 0)
	for _, m := range p.Manifest.Mixins {
		mixinDir, err := p.GetMixinDir(m)
		if err != nil {
			return nil, err
		}

		r := mixin.NewRunner(m, mixinDir, false)
		r.Command = "build"
		r.Input = "" // TODO: let the mixin know about which steps will be executed so that it can be more selective about copying into the invocation image

		// Copy the existing context and tweak to pipe the output differently
		mixinStdout := &bytes.Buffer{}
		var mixinContext cxt.Context
		mixinContext = *p.Context
		mixinContext.Out = mixinStdout   // mixin stdout -> dockerfile lines
		mixinContext.Err = p.Context.Out // mixin stderr -> logs
		r.Context = &mixinContext

		err = r.Validate()
		if err != nil {
			return nil, err
		}

		err = r.Run()
		if err != nil {
			return nil, err
		}

		l := strings.Split(mixinStdout.String(), "\n")
		lines = append(lines, l...)
	}
	return lines, nil
}

func (p *Porter) prepareDockerFilesystem() error {
	fmt.Fprintf(p.Out, "Copying dependencies ===> \n")
	for _, dep := range p.Manifest.Dependencies {
		err := p.copyDependency(dep.Name)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(p.Out, "Copying porter runtime ===> \n")

	runTmpl, err := p.Templates.GetRunScript()
	if err != nil {
		return err
	}

	err = p.FileSystem.MkdirAll(filepath.Dir(config.RunScript), 0755)
	if err != nil {
		return err
	}

	err = p.FileSystem.WriteFile(config.RunScript, runTmpl, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", config.RunScript)
	}

	pr, err := p.GetPorterRuntimePath()
	if err != nil {
		return err
	}
	err = p.CopyFile(pr, "cnab/app/porter-runtime")
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Copying mixins ===> \n")
	for _, mixin := range p.Manifest.Mixins {
		err := p.copyMixin(mixin)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Porter) copyDependency(bundle string) error {
	fmt.Fprintf(p.Out, "Copying bundle dependency %s ===> \n", bundle)
	bundleDir, err := p.GetBundleDir(bundle)
	if err != nil {
		return err
	}

	err = p.Context.CopyDirectory(bundleDir, "cnab/app/bundles", true)
	return errors.Wrapf(err, "could not copy bundle directory contents for %s", bundle)
}

func (p *Porter) copyMixin(mixin string) error {
	fmt.Fprintf(p.Out, "Copying mixin %s ===> \n", mixin)
	mixinDir, err := p.GetMixinDir(mixin)
	if err != nil {
		return err
	}

	err = p.Context.CopyDirectory(mixinDir, "cnab/app/mixins", true)
	return errors.Wrapf(err, "could not copy mixin directory contents for %s", mixin)
}

func (p *Porter) buildInvocationImage(ctx context.Context) error {
	fmt.Fprintf(p.Out, "\nStarting Invocation Image Build =======> \n")
	path, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "could not get current working directory")
	}
	buildOptions := types.ImageBuildOptions{
		SuppressOutput: false,
		PullParent:     false,
		Tags:           []string{p.Config.Manifest.Image},
		Dockerfile:     "Dockerfile",
	}
	tar, err := archive.TarWithOptions(path, &archive.TarOptions{})
	if err != nil {
		return err
	}

	cli, err := command.NewDockerCli()
	if err != nil {
		return errors.Wrap(err, "could not create new docker client")
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}

	response, err := cli.Client().ImageBuild(context.Background(), tar, buildOptions)
	if err != nil {
		return err
	}

	dockerOutput := ioutil.Discard
	if p.IsVerbose() {
		dockerOutput = p.Out
	}

	termFd, _ := term.GetFdInfo(dockerOutput)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
	err = jsonmessage.DisplayJSONMessagesStream(response.Body, dockerOutput, termFd, isTerm, nil)
	if err != nil {
		return errors.Wrap(err, "failed to stream docker build output")
	}
	return nil
}

func (p *Porter) buildBundle(invocationImage string, digest string) error {
	fmt.Fprintf(p.Out, "\nGenerating Bundle File with Invocation Image %s =======> \n", invocationImage)
	b := bundle.Bundle{
		Name:        p.Config.Manifest.Name,
		Description: p.Config.Manifest.Description,
		Version:     p.Config.Manifest.Version,
		Custom:      make(map[string]interface{}, 1),
	}
	image := bundle.InvocationImage{
		BaseImage: bundle.BaseImage{
			Image:     invocationImage,
			ImageType: "docker",
			Digest:    digest,
		},
	}

	b.InvocationImages = []bundle.InvocationImage{image}

	b.Images = p.generateBundleImages()
	b.Parameters = p.generateBundleParameters()
	b.Credentials = p.generateBundleCredentials()
	b.Custom[config.CustomBundleKey] = p.GenerateStamp(p.Manifest)

	return p.writeBundle(b)
}

func (p Porter) writeBundle(b bundle.Bundle) error {
	f, err := p.Config.FileSystem.OpenFile("cnab/bundle.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return errors.Wrapf(err, "error creating cnab/bundle.json")
	}
	_, err = b.WriteTo(f)
	return errors.Wrap(err, "error writing to cnab/bundle.json")
}

func (p *Porter) generateBundleImages() map[string]bundle.Image {
	images := make(map[string]bundle.Image, len(p.Manifest.ImageMap))

	for i, refImage := range p.Manifest.ImageMap {
		img := bundle.Image{
			Description: refImage.Description,
			BaseImage: bundle.BaseImage{
				Image:         refImage.Image,
				Digest:        refImage.Digest,
				ImageType:     refImage.ImageType,
				MediaType:     refImage.MediaType,
				OriginalImage: refImage.OriginalImage,
				Size:          refImage.Size,
			},
		}
		if refImage.Platform != nil {
			img.Platform = &bundle.ImagePlatform{
				Architecture: refImage.Platform.Architecture,
				OS:           refImage.Platform.OS,
			}
		}
		images[i] = img
	}

	return images
}

func (p *Porter) generateBundleParameters() map[string]bundle.ParameterDefinition {
	params := map[string]bundle.ParameterDefinition{}
	for _, param := range append(p.Manifest.Parameters, p.buildDefaultPorterParameters()...) {
		fmt.Fprintf(p.Out, "Generating parameter definition %s ====>\n", param.Name)
		p := bundle.ParameterDefinition{
			DataType:      param.DataType,
			DefaultValue:  param.DefaultValue,
			AllowedValues: param.AllowedValues,
			MinValue:      param.MinValue,
			MaxValue:      param.MaxValue,
			MinLength:     param.MinLength,
			MaxLength:     param.MaxLength,
		}

		// If the default is empty, set required to true.
		if param.DefaultValue == nil {
			p.Required = true
		}

		if param.Metadata.Description != "" {
			p.Metadata = &bundle.ParameterMetadata{Description: param.Metadata.Description}
		}

		if param.Destination != nil {
			p.Destination = &bundle.Location{
				EnvironmentVariable: param.Destination.EnvironmentVariable,
				Path:                param.Destination.Path,
			}
		} else {
			p.Destination = &bundle.Location{
				EnvironmentVariable: strings.ToUpper(param.Name),
			}
		}
		params[param.Name] = p
	}
	return params
}

func (p *Porter) buildDefaultPorterParameters() []config.ParameterDefinition {
	return []config.ParameterDefinition{
		{
			Name: "porter-debug",
			Destination: &config.Location{
				EnvironmentVariable: "PORTER_DEBUG",
			},
			DataType:     "bool",
			DefaultValue: false,
			Metadata: config.ParameterMetadata{
				Description: "Print debug information from Porter when executing the bundle"},
		},
	}
}

func (p *Porter) generateBundleCredentials() map[string]bundle.Location {
	params := map[string]bundle.Location{}
	for _, cred := range p.Manifest.Credentials {
		fmt.Fprintf(p.Out, "Generating credential %s ====>\n", cred.Name)
		l := bundle.Location{
			Path:                cred.Path,
			EnvironmentVariable: cred.EnvironmentVariable,
		}
		params[cred.Name] = l
	}
	return params
}
