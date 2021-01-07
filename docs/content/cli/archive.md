---
title: "porter archive"
slug: porter_archive
url: /cli/porter_archive/
---
## porter archive

Archive a bundle from a reference

### Synopsis

Archives a bundle by generating a gzipped tar archive containing the bundle, invocation image and any referenced images.

```
porter archive FILENAME --reference PUBLISHED_BUNDLE [flags]
```

### Examples

```
  porter archive mybun.tgz --reference getporter/porter-hello:v0.1.0
  porter archive mybun.tgz --reference localhost:5000/getporter/porter-hello:v0.1.0 --force

```

### Options

```
      --force               Force a fresh pull of the bundle
  -h, --help                help for archive
      --insecure-registry   Don't require TLS for the registry
      --reference string    Use a bundle in an OCI registry specified by the given reference.
```

### Options inherited from parent commands

```
      --debug           Enable debug logging
      --debug-plugins   Enable plugin debug logging
```

### SEE ALSO

* [porter](/cli/porter/)	 - I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool

