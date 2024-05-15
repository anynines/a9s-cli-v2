# Backlog

## Next

* [ARCHITECTURE] Install a8s PG on an existing cluster
    * Decide which command/verb to use
        * [DONE] `a9s create stack`
            * Applies the a8s stack to the current k8s cluster
                * The following cluster/context/namespace is selected:
                    Do you want to apply the a8s stack to this cluster?
    * Make the context/namespaces configurable
        
        * Streamline the UX for both `create cluster` and `create stack`
            * Per default the context `a8s-demo` with the namespace `a8s-demo` is mandatorily required. 
            * Let the user select a different namespace / context name
            * Make the default namespace/context not contain the word `demo`.
                * Point out that `-c` can be used to specify a context / clustername
        * Update the readme
            * Add `create stack` documentation
    * Write a tutorial on how to apply a stack to an existing AWS cluster
            
            
* [Question] Remove ?
    * Should there be a remove option?
    * Removes the a8s stack from the current k8s cluster

## Unassigned

* [POSTPONED] Run test suite on windows
        * Perform cold test run
        * Run test suite

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

* Epic: Add AWS-EKS as a provider to create cluster-stack
    * `a9s create cluster a8s -p aws`
    * Rerequisites
        * `aws` command and `aws configure` been run
            * Darwin: `brew install awscli`
        * `eksctl` command
            * Darwin: `brew install eksctl`
    * Steps
        * `aws configure`
        * `eksctl create cluster --name my-eks-cluster --region us-west-2 --node-type t2.medium --nodes 3`
    * Tutorial: 
        * https://medium.com/@prateek.malhotra004/step-by-step-guide-creating-an-amazon-eks-cluster-with-aws-cli-edab2c7eac41


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

