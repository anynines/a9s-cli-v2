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

### Version Bump

**Note**: The version of the a9s CLI is maintained in the file `VERSION` and used in the `Makefile` and passed via `ldflags` into the binary. Therefore, when issuing new releases the `VERSION` file needs to be updated before building the binaries.

## Making Release

Example: Release `v0.10.0`.

1. Ensure that all tests are run including `e2e-tests`.
1. Ensure the `main` branch is up to date and clean with all necessary changes comitted.
1. Ensure the release state of the `main` branch is tagged with the tag `v0.10.0`.
    * This can be done using `git flow release start 0.10.0` and `git flow release finish 0.10.0`.
1. Cross-compile binaries with `make build_all`.
1. Upload binaries to the release folder in the S3 bucket
1. Run `run-ci.bash` on the CI VM executing `e2e-tests` on linux.
1. Update and upload the `releases.json` file to the S3 bucket.
