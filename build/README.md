These are assets used by the CI build.

* testdata - test data used for the integration tests, such as credentials and test bundles.
* images - source for any docker images that we publish. 

## Azure Pipelines

See all Porter related [pipelines across all repositories](https://dev.azure.com/deislabs/porter/_build).

Our pipeline is broken into a few discrete builds so that we can control how and when they are triggered:

* **porter-release: azure-pipelines.release.yml**: Validates canary and tag releases. This can be tested in a pull request 
  using `/azp run porter-release` though steps that require credentials will fail. [View Latest Builds](https://dev.azure.com/deislabs/porter/_build?definitionId=23)
* **porter-check-install: azure-pipelines.install.yml**: Validates our install scripts against canary and tag releases. [View Latest Builds](https://dev.azure.com/deislabs/porter/_build?definitionId=16)
* **porter: azure-pipelines.pr-automatic.yml**: Validates everything we can without a live environment. [View Latest Builds](https://dev.azure.com/deislabs/porter/_build?definitionId=6)
* **porter-integration: azure-pipelines.pr-manual.yml**: Validates a pull request using a live environment. Requires manual triggering 
  using `/azp run porter-integration` by a maintainer because this accesses secrets in the environment, 
  e.g. kubeconfig. We don't want this accessible to anyone who submits a PR without a code review first. [View Latest Builds](https://dev.azure.com/deislabs/porter/_build?definitionId=22)

### Documentation Only Builds

The `porter` and `porter-integration` builds [detect if the build is
documentation only](doc-only-build.sh) and will short circuit the build. They
look for changes to the website, markdown files, workshop materials and
repository metadata files that do not affect the build.