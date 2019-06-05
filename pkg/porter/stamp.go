package porter

import (
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pkg/errors"
)

// EnsureBundleIsUpToDate ensures that the bundle is up to date with the porter manifest,
// if it is out-of-date, performs a build of the bundle.
func (p *Porter) EnsureBundleIsUpToDate(opts bundleFileOptions) error {
	if opts.File == "" {
		return nil
	}

	upToDate, err := p.IsBundleUpToDate(opts)
	if err != nil {
		fmt.Fprintln(p.Err, "warning", err)
	}

	if !upToDate {
		fmt.Fprintln(p.Out, "Building bundle ===>")
		return p.Build()
	}
	return nil
}

// IsBundleUpToDate checks the hash of the manifest against the hash in cnab/bundle.json.
func (p *Porter) IsBundleUpToDate(opts bundleFileOptions) (bool, error) {
	if exists, _ := p.FileSystem.Exists(opts.CNABFile); exists {
		bunData, err := p.FileSystem.ReadFile(opts.CNABFile)
		if err != nil {
			return false, errors.Wrapf(err, "could not read data from %s", opts.CNABFile)
		}

		bun, err := bundle.Unmarshal(bunData)
		if err != nil {
			return false, errors.Wrapf(err, "could not marshal data from %s", opts.CNABFile)
		}

		oldStamp, err := p.LoadStamp(bun)
		if err != nil {
			return false, errors.Wrapf(err, "could not load stamp from %s", opts.CNABFile)
		}

		newStamp := p.GenerateStamp(p.Manifest)
		return oldStamp.ManifestDigest == newStamp.ManifestDigest, nil
	}

	return false, nil
}
