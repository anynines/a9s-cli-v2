# Backlog

## Next 

* Rename `a9s create demo a8s` to `a9s create local-dev-env` or something similar.
    * [DONE] Decide which word to use instead of `demo`
        * `a9s create stack` 
        * `a9s create environment` &&|| `a9s create env`
        * `a9s create cluster`
    * Rename to `a9s create cluster`
        * Rename create cmd
        * Rename delete cmd
        * Adapt Readme
        * Adapt unit tests
        * Adapt e2e tests
        * Write changelog
        * Adapt implementation notes


* BUG: When there's an exec format error and thus the a8s-system can't start, the a9s create demo a8s command does not recognize the failing pods but falsely thinks that the a8s-system is running.

## Release: KubeCon Pre-Release

### a8s PG Demo
**OBJECTIVE**: Given a local computer and the `a9s` CLI, a local a8s PG demo can be performed.

* **Github authentication**: A non-anynines user must be able to perform the a8s PG demo
    * There mustn't be a requirement to have certain access privileges for any anynines github repository.
    * Possible solution: Include github authentication into the CLI
        * Currently it is assumed that `gh auth login` is performed.

### Public a8s PG Self-Demo
**OBJECTIVE**: A visitor can perform a a8s PG demo on a local computer without the involvement of a sales engineer and/or online documentation.

* Create and publish `a9s` binaries for supported OSes
    * Run test suite on linux
        * Perform cold test run
        * Run test suite
    * Run test suite on windows
        * Perform cold test run
        * Run test suite
    * Run test suite test on macos
        * Perform cold test run
        * Run test suite
* Create CI/CD pipeline that
    * Creates binaries for each supported OS
    * Runs tests for each supported OS
    * Publishes binaries for each supported OS
        * Upload binary
        * Publish docs / changelog
* Create and deploy and a8s PG self-demo landing page
* Create a marketing campaign with ads to feed the landing page
* Establish a funnel and funnel monitoring to measure success

## Release: KubeCon Final

## Unassigned

* CHORE: Harmonize variable declaration for params. Use package config. Maybe use viper.


* BUGFIX: `a9s create demo a8s` fails if minikube has never started before.

```sh
    minikube profile list -o json
{"error":{"Op":"open","Path":"/home/ubuntu/.minikube/profiles","Err":2}}
```

* Feature: In order to script the usage of a9s I want to pass all information to a9s create demo a8s to avoid any user dialog including the creation of a work direction as well as bucket credentials.


* Evaluate to use viper to handle config options: https://github.com/spf13/viper

* Backup: A failed backup should be indicated to the user.
* Restore: A failed restore should be indicated to the user.

* Epic: POC on AWS
* Epic: Allow Minio as Object Store
    * OBJECTIVE: Allow a local demo without a depedency to external object stores such as AWS S3.

* Epic: Use case: **Deploy the demo app**
    * Feature: a9s create app -k directoryWithKustomize.yaml
    * Demo App
        * The demo consists of an app and a service + kustomize file.
        * `kubectl apply -k ...`

* Epic: Verbosity
    * For all commands ensure that the `-v` flag is respected and without it there's a clean output

* Feature: Create service instance from backup
    * Combines:
        * Create service instance
        * Restore backup

* Observability: Backup/Restore: The WaitForKubernetesResource function should indicate the current status cycling through all states with Status = true (scheduled, complete, ...)


* CHORE: Check if makeup.WaitForUser(demo.UnattendedMode) is used consistently for all new commands instance & backup

* FEATURE: Backups: Deleting Backups and Restore CRs
    * For completeness: Create command `a9s delete pg backup ...`
    * For completeness: Create command `a9s delete pg restore ...`

* Usability: Backup, Infrastructure Region: When executing a9s create demo a8s for the first time, the infrastructure-region should be queried as a user input instead of being a default-parameter. The probability is too high that the user choses a non-viable default option instead of providing a valid region.

* Create binaries in a release matrix, e.g. using Go Release Binaries with Gihub Action Matrix Strategy
    * https://github.com/marketplace/actions/go-release-binaries
    * https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix

* Create S3 bucket with configs
    * Alternatively: Install a local storage provider, e.g. minio.
        * Costly dependency: add the local storage provider to the backup agent.


# Questions

## Command Structure
* Question: Should the de   mo a8s-pg execute the entire demo or just install the operator? Other commands could be: 
    * Issue: What if a9s-pg is added to the a9s CLI?
        * How should it be resolved?
            * Option a)
                * `a9s pg instance create --isolation pod` > a8s PG
                * `a9s pg instance create --isolation vm` > a9s PG
            * Option b)
                * `a9s a8s-pg instance create ...`
                * `a9s a9s-pg instance create ...`
    * Issue: What if support for the a9s CrossBind services is added to the `a9s`-cli?
        * Then we not only need to differenciate a8s from a9s services but also local from remote service instances.
        * How should it be resolved?
            * Option a) Explicit commands/params/flags
                * Option a-1)
                    * `a9s create remote pg instance`
                    * `a9s create local pg instance`
                        * Allows each variant to have its own set of params/flags
                * Option a-2)
                    * `a9s create pg instance --local`
                    * `a9s create pg instance --remote`
                        * This variant may be harder to implement as local and remote PGs may have different attributes.

            * Option b) Implicit detection of the context
                * Not possible if both a local operator and a remote version of, let's say, a8s-pg is available
    * a8s-pg  
        * `a9s pg instance`
                * `a9s pg create instance --isolation pod`
            * `create`
        * `a9s pg service-binding`
            * `a9s pg binding` 
            * `a9s pg sb`

