# Development

## Gitflow

This repo is using [gitflow](https://nvie.com/posts/a-successful-git-branching-model/).

## Makefile

There's a `Makefile` to help building and running the cli during development.

## Building a Local Binary

Building a local binary for development purposes can be done with:

    make build

The binary can be found in `bin/a9s`.

## Building All Binaries

Building binaries via cross-compilation for all selected platforms can be done with:

    make build-all

The binaries can be found in `bin/`.

## Testing the CLI

### Unit Tests
The state of unit tests is currently very poor.

Execute the unit tests:

    make tests

Or:

    make test_failfast

### End to End Tests

End-to-end testing can be done using the external Ruby/RSpec test suite located at `e2e-tests` directory.

Ensure that you've setup everything is properly setup by running `a9s create cluster` at least once.

Then execute:

    cd e2s-tests
    rspec

## Version Bump

**Note**: The version of the a9s CLI is maintained in the file `VERSION` and used in the `Makefile` and passed via `ldflags` into the binary. Therefore, when issuing new releases the `VERSION` file needs to be updated before building the binaries.

## Making Release

Example: Release `v0.10.0`.

1. Ensure that all tests are run including `e2e-tests`.
1. Ensure that the documentation is updated incl. 
    1. adding a copy of the docs to the corresponding version directory under `docs/versioned_docs`
    2. updating `docs/versions.json` by adding the new version to the array
1. Ensure the Readme is updated.
1. Ensure the Changelog is updated.
1. Ensure the `main` branch is up to date and clean with all necessary changes comitted.
1. Ensure the release state of the `main` branch is tagged with the tag `v0.10.0`.
    * This can be done using `git flow release start 0.10.0` and `git flow release finish 0.10.0`.
1. Cross-compile binaries with `make build_all`.
1. Upload binaries to the release folder in the S3 bucket
1. Run `run-ci.bash` on the CI VM executing `e2e-tests` on linux.
1. Update and upload the `releases.json` file to the S3 bucket.
1. Deploy the documentation 

## GitHub Release with GoReleaser

This section contains targets for creating GitHub releases using GoReleaser.
GoReleaser automates the build, packaging and release process for Go projects,
integrating with GitHub's release functionality.

> **Note:** This approach is intended for providing releases to end users and
> exists in parallel with our current release process. It will be maintained
> alongside the existing approach until we decide to fully implement the
> release process with GitHub and GitHub Actions.

### Creating GitHub Releases with GoReleaser

This guide outlines the process for creating GitHub releases using GoReleaser.
Following these steps will automate the build, packaging, and release process
for your Go project, streamlining the creation of consistent and professional
releases on GitHub.

#### Prerequisites

1. Install [GoReleaser](https://goreleaser.com/).
2. Set up a GitHub Personal Access Token with appropriate permissions.

#### Steps to Create a Release

1. Set your GitHub token as an environment variable:

    ```bash
    export GITHUB_TOKEN="your_personal_access_token_here"
    ```

2. Add a git tag to the main branch:

    ```bash
    git tag -a v1.0.0 -m "Release v1.0.0"
    ```

3. Run GoReleaser:

    ```bash
    goreleaser release
    ```

#### Testing the Release Process

To test the release process without creating an actual release, use:

```bash
goreleaser release --snapshot --clean
```

This command will simulate the release process without publishing or creating
any permanent artifacts.

# Design Principles / Ideals
* The CLI acts like a personal assistent who knows how to install and use certain Kubernetes extensions including a list of anynines products facilitating their use.
    * The CLI helps with installing a cluster.
    * The CLI helps with writing YAML manifests, e.g. so that users do not have to lookup attributes in the documentation.
* The CLI should not need a tight synchronization with product releases.
    * The release of a new a8s Postgres version, for example, should be working with an existing CLI version.
* Minimize the code owned in the CLI for achieving a specified goal.
    * The use of other CLI tools is preferred over implementing existing functionality again. Just like a human assistent would use existing tools to achieve a certain goal, so does the CLI.
* Automation over documentation.
    * If the CLI can do a task, the task should be automated instead of adding a paragraph to the documentation.
        * Documentation is wonderful but the CLI can interact much better with the user than any documentation could.

# Known Issues / Limitations
* Currently releases are tested on MacOS and Linux.
* Windows binaries are available but they have not been tested.
* Creating a backup for non-existing service instances falsely suggests that the backup has been successful.
* Deletion of backups with `kubectl delete backup ...` get stuck and the deletion doesn't succeed.
* When applying a sql file to an a8s Postgres database using `a9s pg apply --file` ensure that there is no change of the primary pod for clustered instances as otherwise the file might be copied to the wrong pod. There's a slight delay between determining the primary pod and uploading the file to it. 