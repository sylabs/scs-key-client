# SCS Key Client

<a href="https://circleci.com/gh/sylabs/workflows/scs-key-client"><img src="https://circleci.com/gh/sylabs/scs-key-client.svg?style=shield&circle-token=f2b6e1b11393ccadf9ec8712d76da4a48dc8630d"></a>
<a href="https://app.zenhub.com/workspace/o/sylabs/scs-key-client/boards"><img src="https://raw.githubusercontent.com/ZenHubIO/support/master/zenhub-badge.png"></a>

This project provides a Go client for the Singularity Container Services (SCS) Key Service.

## Quick Start

Install the [CircleCI Local CLI](https://circleci.com/docs/2.0/local-cli/). See the [Continuous Integration](#continuous-integration) section below for more detail.

To build and test:

```sh
circleci build
```

## Continuous Integration

This package uses [CircleCI](https://circleci.com) for Continuous Integration (CI). It runs automatically on commits and pull requests involving a protected branch. All CI checks must pass before a merge to a proected branch can be performed.

The CI checks are typically run in the cloud without user intervention. If desired, the CI checks can also be run locally using the [CircleCI Local CLI](https://circleci.com/docs/2.0/local-cli/).
