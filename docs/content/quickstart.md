---
title: QuickStart Guide
descriptions: Get started using Porter
---

## Getting Porter

First make sure Porter is installed.
Please see the [installation instructions](/install/) for more info.

## Create a new bundle
Use the `porter create` command to start a new project:
```
mkdir -p my-bundle/ && cd my-bundle/
porter create
```

This will create a file called `porter.yaml` which contains the configuration for your bundle.
Modify and customize this file for your application's needs.

Here is a very basic `porter.yaml` example:
```
name: my-bundle
version: 0.1.0
description: "this application is extremely important"

invocationImage: my-dockerhub-user/my-bundle:latest

mixins:
  - exec

install:
  - description: "Install Hello World"
    exec:
      command: bash
      arguments:
        - -c
        - echo Hello World

uninstall:
  - description: "Uninstall Hello World"
    exec:
      command: bash
      arguments:
        - -c
        - echo Goodbye World
```

## Build the bundle

The `porter build` command will create a
[CNAB-compliant](https://github.com/deislabs/cnab-spec/blob/master/101-bundle-json.md) `bundle.json`,
as well as build and push the associated invocation image:
```
porter build
```

Note: Make sure that the `invocationImage` listed in you `porter.yaml`  is a reference that you are
able to `docker push` to and that your end-users are able to `docker pull` from.

## Install the bundle

_Wondering the differences between Duffle and Porter? Please see [this page](/porter-or-duffle/)._

First, make sure Duffle is installed
(see [install instructions](https://github.com/deislabs/duffle/blob/master/README.md#getting-started)).

You can then use `duffle install` to install your bundle ("demo" is the unique installation name):
```
duffle install demo -f bundle.json
```

The `duffle list` command can be used to show all installed bundles.

If you wish to uninstall the bundle, you can use `duffle uninstall`:
```
duffle uninstall demo
```