# Contributing Guidelines

The following is a set of guidelines for contributing to NGINX Kubernetes Gateway. We really appreciate that you are
considering contributing!

#### Table Of Contents

[Ask a Question](#ask-a-question)

[Getting Started](#getting-started)

[Contributing](#contributing)

* [Issues and Discussions](#issues-and-discussions)
* [Development Guide](#development-guide)

[Style Guides](#style-guides)

* [Git Style Guide](#git-style-guide)
* [Go Style Guide](#go-style-guide)

[Code of Conduct](CODE_OF_CONDUCT.md)

[Contributor License Agreement](#contributor-license-agreement)

## Ask a Question

To ask a question please use [Github Discussions](https://github.com/nginxinc/nginx-kubernetes-gateway/discussions).

[NGINX Community Slack](https://community.nginx.org/joinslack) has a dedicated channel for this
project -- `#nginx-kubernetes-gateway`.

Please reserve GitHub issues for feature requests and bugs rather than general questions.

## Getting Started

Follow our [Installation Instructions](docs/installation.md) to get the NGINX Kubernetes Gateway up and running.

### Project Structure

* NGINX Kubernetes Gateway is written in Go and uses the open source NGINX software as the data plane.
* The project follows a standard Go project layout
    * The main code is found at `cmd/gateway/`
    * The internal code is found at `internal/`
    * Build files for Docker are found under `build/`
    * Deployment yaml files are found at `deploy/`
    * External APIs, clients, and SDKs can be found under `pkg/`
* We use [Go Modules](https://github.com/golang/go/wiki/Modules) for managing dependencies.
* We use [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) for our BDD style unit
  tests.

## Contributing

### Issues and Discussions

#### Open a Discussion

If you have any questions, ideas, or simply want to engage in a conversation with the community and maintainers, we
encourage you to open a [discussion](https://github.com/nginxinc/nginx-kubernetes-gateway/discussions) on GitHub. We
highly recommend that you open a discussion about a potential enhancement before opening an issue. This enables the
maintainers to gather valuable insights regarding the idea and its use cases, while also giving the community an
opportunity to provide valuable feedback.

#### Report a Bug

To report a bug, open an issue on GitHub with the label `bug` using the available bug report issue template. Please
ensure the issue has not already been reported.

#### Suggest an Enhancement

To suggest an enhancement, please create an issue on GitHub with the label `proposal` using the available feature issue
template.

#### Issue lifecycle

When an issue or PR is created, it will be triaged by the core development team and assigned a label to indicate the
type of issue it is (bug, feature request, etc) and to determine the milestone. Please see
the [Issue Lifecycle](ISSUE_LIFECYCLE.md) document for more information.

### Development Guide

Before beginning development, please familiarize yourself with the following documents:

- [Developer Quickstart](docs/developer/quickstart.md): This guide provides a quick and easy walkthrough of setting up
  your development environment and executing tasks required when submitting a PR.
- [Branching and Workflow](docs/developer/branching-and-workflow.md): This document outlines the project's specific
  branching and workflow practices, including instructions on how to name a branch.
- [Implement a Feature](docs/developer/implementing-a-feature.md): A step-by-step guide on how to implement a feature or
  bug.
- [Testing](docs/developer/testing.md): The project's testing guidelines, includes both unit testing and manual testing
  procedures. This document explains how to write and run unit tests, and how to manually verify changes.
- [Pull Request Guidelines](docs/developer/pull-request.md): A guide for both PR submitters and reviewers, outlining
  guidelines and best practices to ensure smooth and efficient PR processes.

## Style Guides

### Git Style Guide

* Keep a clean, concise and meaningful git commit history on your branch, rebasing locally and squashing before
  submitting a PR
* Follow the guidelines of writing a good commit message as described [here](https://chris.beams.io/posts/git-commit/)
  and summarized in the next few points
    * In the subject line, use the present tense ("Add feature" not "Added feature")
    * In the subject line, use the imperative mood ("Move cursor to..." not "Moves cursor to...")
    * Limit the subject line to 72 characters or less
    * Reference issues and pull requests liberally after the subject line
    * Add more detailed description in the body of the git message (`git commit -a` to give you more space and time in
      your text editor to write a good message instead of `git commit -am`)

### Go Style Guide

* Run `gofmt` over your code to automatically resolve a lot of style issues. Most editors support this running
  automatically when saving a code file.
* Run `go lint` and `go vet` on your code too to catch any other issues.
* Follow this guide on some good practice and idioms for Go -  https://github.com/golang/go/wiki/CodeReviewComments
* To check for extra issues, install [golangci-lint](https://github.com/golangci/golangci-lint) and run `make lint`
  or `golangci-lint run`

## Contributor License Agreement

Individuals or business entities who contribute to this project must have completed and submitted
the [F5® Contributor License Agreement](F5ContributorLicenseAgreement.pdf) prior to their code submission being included
in this project. To submit, please print out the [F5® Contributor License Agreement](F5ContributorLicenseAgreement.pdf),
fill in the required sections, sign, scan, and send executed CLA to kubernetes@nginx.com. Please include your github
handle in the CLA email.
