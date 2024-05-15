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

The state of unit tests is currently very poor.

End-to-end testing can be done using the external Ruby/RSpec test suite located at `e2e-tests` directory.

## Version Bump

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



# Design Principles / Ideals
* The CLI acts like a personal assistent who knows the a9s products and helps to use them more easily.
    * The CLI helps with installing a cluster.
    * The CLI helps with writing YAML manifests, e.g. so that users do not have to lookup attributes in the documentation.
* The CLI should not need a tight synchronization with product releases.
    * The release of a new a8s Postgres version, for example, should be working with an existing CLI version.

# Known Issues / Limitations
* Currently releases are tested on MacOS and Linux.
* Windows binaries are available but they have not been tested.
* Creating a backup for non-existing service instances falsely suggests that the backup has been successful.
* Deletion of backups with `kubectl delete backup ...` get stuck and the deletion doesn't succeed.
* When applying a sql file to an a8s Postgres database using `a9s pg apply --file` ensure that there is no change of the primary pod for clustered instances as otherwise the file might be copied to the wrong pod. There's a slight delay between determining the primary pod and uploading the file to it. 