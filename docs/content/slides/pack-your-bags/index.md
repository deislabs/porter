---
title: "Pack Your Bags: Managing Distributed Applications with CNAB"
description: |
  Learn how to use Cloud Native Application Bundles (CNAB) and Porter to bundle up your
  application and its infrastructure so that it is easier to manage.
url: "/pack-your-bags/"
---
class: center, middle

# Pack your bags
##  Managing Distributed Applications with CNAB

---
name: setup

# Workshop Setup

It can take a while for things to download and install over the workshop wifi,
so please go to the workshop materials directory and follow the setup instructions
to get all the materials ready.

.center[👩🏽‍✈️ https://porter.sh/pack-your-bags/#setup 👩🏽‍✈️ ]

* Clone the workshop repository
  ```
  git clone https://github.com/deislabs/porter.git
  cd porter/workshop
  ```
* [Install Porter](https://porter.sh/install)
* Create a Kubernetes Cluster on [macOS](https://docs.docker.com/docker-for-mac/kubernetes/) or [Windows](https://docs.docker.com/docker-for-windows/kubernetes/)
* [Install Helm 2](https://helm.sh/docs/install/)
* Initialize Helm on your cluster by running `helm init`


---
name: agenda

# Agenda

1. What is CNAB?
2. Manage Bundles with Porter
3. Authoring Bundles

---
name: introductions

# Introductions

<div id="introductions">
  <div class="left">
    <img src="/images/carolynvs.jpg" width="150px" />
    <p>Carolyn Van Slyck</p>
    <p>Senior Software Engineer</p>
    <p>Microsoft Azure</p>
  </div>
  <div class="right">
    <img src="/images/jerrycar.jpg" width="200px" />
    <p>Jeremy Rickard</p>
    <p>Senior Software Engineer</p>
    <p>Microsoft Azure</p>
  </div>
</div>

---
name: kickoff
class: center, middle
# First A Quick Demo!

---
name: cnab
# What's a CNAB???

---
class: center, middle

# Let's Answer That With A Story!

---
class: center, middle

# The Cast

---
class: center, middle

# You!
.center[
  ![you, a developer](/images/pack-your-bags/you-a-developer.jpg)
]

---
class: center, middle

# Your friend!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle

# Your friend!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle

# Your App!
.center[
  ![it's the journey that matters](/images/pack-your-bags/mcguffin.png)
]

---
class: center, middle

# Your Fans!
.center[
  ![trending](/images/pack-your-bags/your-fans.jpg)
]

---
class: center, middle

# Act One!
# You Built an App
.center[
  ![you again](/images/pack-your-bags/you-a-developer.jpg)
  ![it's the journey that matters](/images/pack-your-bags/mcguffin.png)
]

---
class: center, middle

# It Runs Happily In The Cloud
# ....your cloud
.center[
  ![that's a bingo](/images/pack-your-bags/cloud-bingo.png)
]

---
class: center, middle

# Act Two
# Your Friend Wants To Run It!
.center[
  ![your friend, a computer user](/images/pack-your-bags/your-friend-a-user.jpg)
]

---
class: center, middle

# You write impressive docs
.center[
  ![you fight for the users](/images/pack-your-bags/scroll-of-truth.png)
]

---
class: center, middle

# You write impressive docs
.center[
  ![you fight for the users](/images/pack-your-bags/scroll-of-truth.png)
]

---
class: center, middle

# Your friend does not thank you....
.center[
  ![you fight for the users](/images/pack-your-bags/Spongebob-patrick-crying.jpg)
]

.footnote[http://vignette3.wikia.nocookie.net/spongebob/images/f/f0/Spongebob-patrick-crying.jpg/revision/latest?cb=20140713205315]

---
class: center, middle

# So you work together....
.center[
  ![pair programming](/images/pack-your-bags/working-together.jpg)
]

---
class: center, middle

# Finally your friend has McGuffin in his cloud...then you help a few more people
.center[
  ![go team](/images/pack-your-bags/working-together.jpg)
]

---
class: center, middle

# Suddenly McGuffin has FANS!
.center[
  ![all the github stars!!!](/images/pack-your-bags/your-fans.jpg)
]

---
class: center, middle

# Your impressive docs don't really scale though...
.center[
  ![nobody wants to do this](/images/pack-your-bags/scroll-of-truth.png)
]

---
class: center, middle

# Docker made us rethink how we ship the bits of our app...
.center[
  ![ship it](/images/pack-your-bags/container-ship.jpg)
]

---
class: center, middle

# But containers don't really solve this...
.center[
  ![half way there](/images/pack-your-bags/scroll-of-sad-truth.png)
]

---
class: center, middle

# So what do we do...
.center[
  ![this is my thinking face](/images/pack-your-bags/thinking.jpg)
]

---
class: center, middle

# This is the problem CNAB wants to solve

---
class: center, middle

# Hashtag Goals

* Package All The Logic To Make Your App Happen
* Allow Consumer To Verify Everything They Will Install
* Distribute Them In Verifiable Way

---
class: center, middle

# How that works

.center[
  ![workflow](/images/pack-your-bags/he-workflow.png)
  ![magic](/images/pack-your-bags/magic.gif)
]

.footnote[_http://www.reactiongifs.com/magic-3_]

---
name: anatomy
class: center, middle

# Anatomy of a Bundle

.center[
  ![so what is it](/images/pack-your-bags/anatomy.png)
]

---
class: center, middle

# Your App...

---
class: center, middle

# The Invocation Image

* MSI for the Cloud?
* It's a Docker Image
* It contains all the tools you need to install your app
* It contains configuration, metadata, templates, etc

---
class: center, middle

# The Invocation Image

.center[
  ![so what is it](/images/pack-your-bags/easy-bake-oven-image.png)
]

---
class: center, middle

# The Bundle Descriptor

* JSON!
* List of the invocation image(s) (with digests!)
* List of the application image(s) (with digests!)
* Definitions for inputs and outputs
* Can be signed

---
name: sharing
class: center, middle

# So we can install things
# So we can verify what we are going to install
--
# How do we distribute bundles?

---
class: center, middle

## Sharing Images With OCI Registries (Docker Registries)

.center[
  ![how docker shares](/images/pack-your-bags/ship-it.png)
]

---
class: center, middle

## Distributing App and Invocation Images is solved
--
## So what about the bundle?

---
class: center, middle

## Sharing Bundles With OCI Registries (Docker Registries)

.center[
  ![how oci shares bundles](/images/pack-your-bags/share-bundles.png)
]

---
class: center, middle

## OCI Registries Can Store Lots of Things

* CNAB today is working within the OCI Spec (not optimal)
* CNAB Spec group working with OCI to improve this

---

# CNAB Specification

### The Bundle format
--
### Defines how things are passed into and out of the invocation image
--
### A required entrypoint in invocation image
--

### Well-defined verbs
--
* Install
* Upgrade
* Uninstall
--

---

# Breakdown of Azure MySQL Wordpress

---
class: center, middle

# Manage Bundles with Porter

.center[
  🚨 Not Setup Yet? 🚨

  https://porter.sh/pack-your-bags/#setup
  
  ]
---
name: hello
class: center

# Tutorial
# Hello World

.center[
  ![whale saying hello](/images/whale-hello.png)
]
---

## porter create

```console
$ porter create --help
Create a bundle. This generates a porter bundle in the current directory.
```

---

### porter.yaml
**The most important file**.  Edit and check this in. Everything else is optional.

### README.md
Explains the other files in detail

### Dockerfile.tmpl
Optional template for your bundle's invocation image

### .gitignore
Suggested set of files to ignore in git

### .dockerignore
Suggested set of files to not include in your bundle

---

# porter.yaml

```yaml
mixins:
  - exec

name: HELLO
version: 0.1.0
description: "An example Porter configuration"
invocationImage: porter-hello:latest

install:
  - exec:
      description: "Install Hello World"
      command: bash
      arguments:
        - -c
        - echo Hello World
```

---

## Try it out: porter create

```console
$ mkdir hello
$ porter create
creating porter configuration in the current directory
$ ls
Dockerfile.tmpl  README.md  porter.yaml
```

---

## porter build

```console
$ porter build --help
Builds the bundle in the current directory by generating a Dockerfile 
and a CNAB bundle.json, and then building the invocation image.
```

---

## Try it out: porter build

```console
$ porter build

Copying dependencies ===>
Copying porter runtime ===>
Copying mixins ===>
Copying mixin exec ===>
Generating Dockerfile =======>
Writing Dockerfile =======>
Starting Invocation Image Build =======>
Generating Bundle File with Invocation Image porter-hello:latest =======>
Generating parameter definition porter-debug ====>
```

---

# What did Porter do? 🔎

---
### Dockerfile
```Dockerfile
FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY . /cnab/app
RUN mv /cnab/app/cnab/app/* /cnab/app && rm -r /cnab/app/cnab
# exec mixin has no buildtime dependencies
```
.footnote[🚨 Generated by Porter]

---

### cnab/
```console
$ tree cnab/
cnab
├── app
│   ├── mixins
│   │   └── exec
│   │       ├── exec
│   │       └── exec-runtime
│   ├── porter-runtime
│   └── run
└── bundle.json
```

### cnab/app/run

```bash
#!/usr/bin/env bash
exec /cnab/app/porter-runtime run -f /cnab/app/porter.yaml
```
.footnote[🚨 Generated by Porter]

---

### bundle.json
```json
{
    "description": "An example Porter configuration",
    "invocationImages": [
        {
            "image": "porter-hello:latest",
            "imageType": "docker"
        }
    ],
    "name": "HELLO",
    "parameters": {
        "porter-debug": {
            "destination": {
                "env": "PORTER_DEBUG"
            },
            "metadata": {
                "description": "Print debug information from Porter when executing the bundle"
            },
            "type": "bool"
        }
    },
    "version": "0.1.0"
}
```
.footnote[🚨 Generated by Porter]

---

## porter install

```console
$ porter install --help
Install a bundle.

The first argument is the name of the claim to create for the installation. 
The claim name defaults to the name of the bundle.

Flags:
  -c, --cred strings         Credential to use when installing the bundle. 
  -f, --file string          Path to the bundle file to install.
      --param strings        Define an individual parameter in the form NAME=VALUE.
      --param-file strings   Path to a parameters definition file for the bundle
```

---
## CNAB: What Executes Where

TODO PICTURE

---

## Try it out: porter install

```console
$ porter install

installing HELLO...
executing porter install configuration from /cnab/app/porter.yaml
Install Hello World
Hello World
execution completed successfully!
```

---
class: center, middle

# BREAK

---
name: mellamo
class: center

# Tutorial
# Hi, My Name is _

.center[
  ![como se llamo me llamo llama](/images/me-llamo.jpg)
]

---
name: parameters

## Parameters

Variables in your bundle that you can specify when you execute the bundle
and are loaded into the bundle either as environment variables or files.

### Define a Paramer
```yaml
parameters:
- name: name
  type: string
  default: llama
```

### Use a Parameter
```yaml
- "echo Hello, {{ bundle.parameters.name }}"
```

* Needs double quotes around the yaml entry
* Needs double curly braces around the templating
* Uses the format `bundle.parameters.PARAMETER_NAME`

???
Explain defaults and when parameters are required

---

## Try it out: Print Your Name

Modify the hello bundle to print "Hello, YOUR NAME", for example "Hello, Aarti", using a parameter.

1. Edit the porter.yaml to define a parameter named `name`.
1. Use the parameter in the `install` action and echo your name.
1. Rebuild your bundle with `porter build`.
1. Finally run `porter install -p name=YOUR_NAME` and look for your name in the output.

---

### porter bundle list

```console
$ porter bundle list
NAME          CREATED         MODIFIED        LAST ACTION   LAST STATUS
HELLO_LLAMA   5 seconds ago   3 seconds ago   install       success
HELLO         8 minutes ago   8 minutes ago   install       success
```

???
Ask them to list their bundles

---
name: claims

### Claims

Claims are records of any actions performed by CNAB compliant tools on a bundle.

```console
$ cat ~/.porter/claims/HELLO.json
{
  "name": "HELLO",
  "revision": "01DCFCN6AH00SM8E1968XHTSJ5",
  "created": "2019-06-03T14:22:00.952704-05:00",
  "modified": "2019-06-03T14:22:02.449355-05:00",
  "result": {
    "message": "",
    "action": "install",
    "status": "success"
  },
  "parameters": {
    "porter-debug": false,
    "name": "llama"
  },
  ...
}
```

---
name: cleanup-hello
## Cleanup Hello World

First run `porter uninstall` without any arguments:
```console
$ porter uninstall
uninstalling HELLO...
executing porter uninstall configuration from /cnab/app/porter.yaml
Uninstall Hello World
Goodbye World
execution completed successfully!
```

Now run `porter uninstall` with the name you used for the modified bundle:
```console
$ porter uninstall HELLO_LLAMA
uninstalling HELLO_LLAMA...
executing porter uninstall configuration from /cnab/app/porter.yaml
Uninstall Hello llama
Goodbye llama
execution completed successfully!
```

---
name: wordpress
class: center

# Tutorial
# Wordpress

---
name: credentials

## Credentials

Variables that can be specified when the bundle is executed that are _associated with the identity 
of the user executing the bundle_, and are loaded into the bundle either as environment variables or files.

They are mapped from the local system using named credential sets, instead of specified on the command-line.

---
name: creds-v-params
## Credentials vs. Parameters

### Parameters
* Application Configuration
* Stored in the claim
* 🚨 Available in **plaintext** on the local filesystem

### Credentials
* Identity of the user executing the bundle
* Is not stored in the claim
* Has to be presented every time you perform an action

---
name: passwords

## Credentials, Passwords and Sensitive Data

* Credentials are for data identifying data associated with a user. They are 
re-specified every time you run a bundle, and are not stored in the claim.
* Parameters can store sensitive data using the `sensitive` flag. This prevents 
the value from being printed to the console.
* We (porter) and the CNAB spec are working on more robust storage mechanisms for 
claims with sensitive data, and better ways to pull data from secret stores so that 
they don't end up on the file system unencrypted.

In all honesty this area is a work in progress. I would shove as everything in a 
credential for now but be aware of the distinction and where the CNAB spec is moving.

---

## porter credentials generate

```console
$ porter credentials generate --help
Generate a named set of credentials.

The first argument is the name of credential set you wish to generate. If not
provided, this will default to the bundle name. By default, Porter will
generate a credential set for the bundle in the current directory. You may also
specify a bundle with --file.

Bundles define 1 or more credential(s) that are required to interact with a
bundle. The bundle definition defines where the credential should be delivered
to the bundle, i.e. at /root/.kube. A credential set, on the other hand,
represents the source data that you wish to use when interacting with the
bundle. These will typically be environment variables or files on your local
file system.

When you wish to install, upgrade or delete a bundle, Porter will use the
credential set to determine where to read the necessary information from and
will then provide it to the bundle in the correct location.
```

---

## Wordpress Credential Mapping

### ~/.porter/credentials/wordpress.yaml
```yaml
name: wordpress
credentials:
- name: kubeconfig
  source:
    path: /Users/carolynvs/.kube/config
```

### porter.yaml
```yaml
credentials:
- name: kubeconfig
  path: /root/.kube/config
```

---

## Try it out: porter credentials generate

Generate a set of credentials for the wordpress bundle in this repository.

1. Change to the `wordpress` directory under the workshop materials
1. Run `porter credentials generate` and follow the interactive prompts to create a set of credentials
for the wordpress bundle.

???
we all do this together

---

## Try it out: porter install --cred

Install the wordpress bundle and pass it the named set of credentials that you generated.

```console
$ porter install --cred wordpress
```

---
name: cleanup-wordpress

## Cleanup Wordpress

```console
$ porter uninstall --cred wordpress
```

???
Explain why --cred is required again for uninstall 

---
class: center, middle

# BREAK

---
class: center, middle

# Authoring Bundles

---

# Porter Manifest In-Depth

---

# Steps and Actions

---

# Wiring

---

# Templating

---
class: center, middle

# Mixins

---

# Step Outputs

---

# Make Your Own Mixin

---

# Break Glass

---
class: center, middle

# CNAB Best Practices

---

# What would you really put into a bundle?

---

# What does a real bundle look like?

???
Look at the azure examples and quick starts

---

# How does this fit into a CI/C pipeline?

---
class: center, middle

# Tooling

---

# CNAB Tooling Ecosystem

???
Explain where porter shines, what it is good at vs. say docker app

---

# Duffle

???
Mention duffle as a ops tool for managing bundles from multiple orgins at runtime

---
class: center, middle

# Beyond!

---

# Roadmap

???

Both CNAB and Porter for the next 3 months and rest of the year

---

# Next Steps

???
What should someone do if they are interested in CNAB for their work or personal projects?
What is the timeline for the project and how should they be thinking about beginning to incorporate it?

---

# Contribute!

---
class: center, middle

# Choose your own adventure

* Cloud + Break Glass
* Order a pizza with Porter
* Make a mixin