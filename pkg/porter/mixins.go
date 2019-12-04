package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/mixin/feed"
	"get.porter.sh/porter/pkg/printer"
)

// PrintMixinsOptions represent options for the PrintMixins function
type PrintMixinsOptions struct {
	printer.PrintOptions
}

func (p *Porter) PrintMixins(opts PrintMixinsOptions) error {
	mixins, err := p.ListMixins()
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatTable:
		printMixinRow :=
			func(v interface{}) []interface{} {
				m, ok := v.(mixin.Metadata)
				if !ok {
					return nil
				}
				return []interface{}{m.Name, m.VersionInfo.Version, m.VersionInfo.Author}
			}
		return printer.PrintTable(p.Out, mixins, printMixinRow, "Name", "Version", "Author")
	case printer.FormatJson:
		return printer.PrintJson(p.Out, mixins)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, mixins)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) ListMixins() ([]mixin.Metadata, error) {
	// List out what is installed on the file system
	mixins, err := p.Mixins.List()
	if err != nil {
		return nil, err
	}

	// Query each mixin and fill out their version metadata, if available
	for i := range mixins {
		m := &mixins[i]
		v, err := p.Mixins.GetVersionMetadata(*m)
		if err != nil {
			// For now, while we transition from mixins not supporting version --output json, ignore it if a mixin
			// doesn't handle this call
			continue
		}

		m.VersionInfo = *v
	}

	return mixins, nil
}

func (p *Porter) InstallMixin(opts mixin.InstallOptions) error {
	m, err := p.Mixins.Install(opts)
	if err != nil {
		return err
	}

	// TODO: Once we can extract the version from the mixin with json (#263), then we can print it out as installed mixin @v1.0.0
	confirmedVersion, err := p.Mixins.GetVersion(*m)
	if err != nil {
		return err
	}
	if p.Debug {
		fmt.Fprintf(p.Out, "installed %s mixin to %s\n%s", m.Name, m.Dir, confirmedVersion)
	} else {
		fmt.Fprintf(p.Out, "installed %s mixin\n%s", m.Name, confirmedVersion)
	}

	return nil
}

func (p *Porter) UninstallMixin(opts mixin.UninstallOptions) error {
	m, err := p.Mixins.Uninstall(opts)
	if err != nil {
		return err
	}

	if p.Debug {
		fmt.Fprintf(p.Out, "Uninstalled %s mixin from %s", m.Name, m.Dir)
	} else {
		fmt.Fprintf(p.Out, "Uninstalled %s mixin", m.Name)
	}

	return nil
}

func (p *Porter) GenerateMixinFeed(opts feed.GenerateOptions) error {
	f := feed.NewMixinFeed(p.Context)

	err := f.Generate(opts)
	if err != nil {
		return err
	}

	return f.Save(opts)
}

func (p *Porter) CreateMixinFeedTemplate() error {
	return feed.CreateTemplate(p.Context)
}
