package porter

import (
	"fmt"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	BundleLifecycleOpts
}

func (o *InstallOptions) Validate(args []string, cxt *context.Context) error {
	if o.Tag != "" {
		err := o.validateTag()
		if err != nil {
			return err
		}
	}
	return o.sharedOptions.Validate(args, cxt)
}

// ToDuffleArgs converts this instance of user-provided installation options
// to duffle installation arguments.
func (o *InstallOptions) ToDuffleArgs() cnabprovider.InstallArguments {
	args := cnabprovider.InstallArguments{
		ActionArguments: cnabprovider.ActionArguments{
			Claim:                 o.Name,
			BundleIdentifier:      o.CNABFile,
			BundleIsFile:          true,
			Insecure:              o.Insecure,
			Params:                make(map[string]string, len(o.combinedParameters)),
			CredentialIdentifiers: make([]string, len(o.CredentialIdentifiers)),
			Driver:                o.Driver,
		},
	}

	// Do a safe copy so that modifications to the duffle args aren't also made to the
	// original options, which is confusing to debug
	for k, v := range o.combinedParameters {
		args.Params[k] = v
	}
	copy(args.CredentialIdentifiers, o.CredentialIdentifiers)

	return args
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(opts InstallOptions) error {
	if opts.Tag != "" {
		o := &opts
		err := o.populateOptsFromBundlePull(p)
		if err != nil {
			return errors.Wrap(err, "unable to pull bundle before installation")
		}
	}
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	err = p.EnsureBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "installing %s...\n", opts.Name)
	return p.CNAB.Install(opts.ToDuffleArgs())
}
