---
title: "porter bundles explain"
slug: porter_bundles_explain
url: /cli/porter_bundles_explain/
---
## porter bundles explain

Explain a bundle

### Synopsis

Explain how to use a bundle by printing the parameters, credentials, outputs, actions.

```
porter bundles explain [flags]
```

### Examples

```
  porter bundle explain
  porter bundle explain --file another/porter.yaml
  porter bundle explain --cnab-file some/bundle.json
  porter bundle explain --tag getporter/porter-hello:v0.1.0
		  
```

### Options

```
      --cnab-file string   Path to the CNAB bundle.json file.
  -f, --file porter.yaml   Path to the Porter manifest. Defaults to porter.yaml in the current directory.
      --force              Force a fresh pull of the bundle
  -h, --help               help for explain
  -o, --output string      Specify an output format.  Allowed values: table, json, yaml (default "table")
  -t, --tag string         Use a bundle in an OCI registry specified by the given tag
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter bundles](/cli/porter_bundles/)	 - Bundle commands

