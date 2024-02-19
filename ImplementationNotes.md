# Implementation Notes

Implementation notes have been written during the implementation. They are collected here as
they may contain information worthwile to persistest. However, be aware that these notes may
represent ideas that have not been implemented or that changes may have been applied. In other words:
**do not expect implementation notes to be in sync with the implementation**.

The implementation notes may document patterns on how certain features are developed by listing individual steps. This may help new developers to find a scaffold to start with when entering the project.

## v0.11.0

* Feature(s): `a9s version`: Prints the version of the a9s CLI
    * Optional: Prints version of installed components on the current kubernetes cluster, e.g. versions of the a8s operators ...
    * Implementation:     
        * Define place where to store the version
            * How and when is the version bumped?
            * How to make sure the version is used consistently 
                * In the code / build
                * Git tag
                * S3 bucket folder
                * CI testing

## v0.10.0

## v0.9.0

* Observability:

    * More robustness for `a9s pg apply`
        1. `a9s pg apply --file` should warn if a service-instance cannot be found
        1. `a9s pg apply` should demand mandatory params without defaults for `-f` and `-i`

    * Feature: Restore
        * The implementation plan is similar to creating the backup.
        * DONE: Create command `a9s create pg restore ...`
        * DONE: Generate a YAML manifest
        * DONE: Apply the YAML manifest
        * DONE: Test manually
        * DONE: Add tests to the e2e test suite
        * This completes the backup / restore cycle.


* `a9s delete pg servicebinding`
    * [DONE] Add command
        * [DONE] Create command variable
        * [DONE] Sett command params
    * [DONE] Implement kubectl command
    * [DONE] Test manually
    * [DONE] Add e2e test

* `a9s create pg servicebinding`
    * OR `service-binding`
    * OR `sb`
    * OR `b`
    * Implementation
        * [DONE] a9s cli
            * [DONE] Spec example: /Users/jfischer/Dropbox/workspace/a8s-demo-allesmeinspro/a8s-deployment/examples/service-binding.yaml
            * [DONE]Implement command
            * [DONE]Implement pg function 
                * [DONE] to generate yaml manifest
                * [DONE] to kubectl apply
            * [DONE] Wait for Secret to be created
                * [DONE]: Does `WaitForKubernetesResource` wait for a particular resource?
            * [DONE] Rename `binding` to `servicebinding` and `b` to `sb`

        * [DONE] manual test
        * [DONE] Changelog entry
        * [DONE] Readme entry
        * [DONE] e-2-test
        * [DONE] Add to demo script


* `a9s pg apply -i instancename --sql "select * from posts"`
    * Ensure: `--sql` and `--file` can't go together

  * `a9s pg apply`: Add `--database` option
    * `a9s pg apply -i instancename --database databasename --sql "select * from posts"`
    * `a9s pg apply -i instancename --database databasename -f my.sql`

  
* Usability: Namespace: `a9s create demo a8s` 
    * Don't use the `default` namespace, instead create a demo namespace, e.g. `a8s-demo`.
    * Provision a8s-pg into custom namespace
        * Change: Change the default namespace for `a9s create pg instance`

* Epic: Namespace param
    * Objective: Ensure commands consistently support `-n` and `--namespace` where applicable.
        * TODO: Check which commands are in scope
            * TODO: For each command in scope check whether `-n` and `--namespace` are supported, create feature if not.
            * Commands in scope are:
                * `a9s create demo a8s`
                    * Question: should the installation of the a8s system be configurable at this point in time?
                    * Answer: no! We can add this later.
                * `a9s create pg instance`                   
                    * DONE --namespace is present
                    * TODO add -n 
                    * TODO test creating a service instance in a non-default namespace
                * `a9s create pg backup`
                    * DONE --namespace is present
                    * DONE add -n 
                    * TODO test creating a service instance in a non-default namespace
                * `a9s create pg restore`
                    * DONE --namespace is present
                    * DONE add -n 
                    * TODO test creating a service instance in a non-default namespace
                * `a9s delete demo a8s`
                    * Support for namespaces will be added later.
                * `a9s delete pg instance`
                    * DONE --namespace is present
                    * TODO add -n 
                    * TODO test creating a service instance in a non-default namespace
                * `a9s pg apply`
                    * DONE --namespace is present
                    * DONE add -n 
                    * TODO test creating a service instance in a non-default namespace
        * Feature: Ensure that all commands accept and correctly apply a custom namespace for all postgres operations
            * Feature: `a9s delete pg instance`: add param `--namespace`
            * Manually test
                * Create a service instance in a non-default namespace
                * Create a backup
                * Restore a backup
                * Delete the service instance
            * Write an rspec test scenario for namespace commands
                * Alternatively, modify existing tests to use a namespace. Assumption: if it works in a custom namepspace, it'll surely work in the default namespace.


* Observability: As a user I expect the a9s cli to provide a descriptive error message when executing a command involving a service instance that does not exist.
    * Branch: `service-instance-existence-1``
    * Affected commands:
        * [DONE] Create backup
        * [DONE] Restore backup
            * [DONE] As a user I expect the restore command to fail when the given service instance does not exist
            * [DONE] As a user I expect the restore command to fail when the given backup does not exist
        * [DONE]: Delete service instance    
    * Create e-2-e tests for attempting to 
        * delete a non-existing service instance
        * create a backup for a non existing service instance
        * create a restore for a non existing service instance
        * create a restore for a non existing backup




* Release v0.9.0: backup/restore v1
    * Feature: `a9s pg apply --file my.sql` 
        * Load data into service instance
        * Implement a command that loads a well known dataset into a service instance
            * Store data somewhere where it will also be checked out, e.g. in a8s-demo
            * Create a command `a9s pg apply --file load_demo_data.sql` 
                * This way the command can also be used for other purposes
                    * It does say "import" as the sql file could also be about deleting data.
                        `a9s pg apply --file delete_demo_data.sql`
                * This should be executing the following statements
                    * `kubectl cp demo_data.sql default/clustered-0:/home/postgres -c postgres`
                    * `kubectl exec -n default clustered-0 -c postgres -- psql -U postgres -d a9s_apps_default_db -f demo_data.sql`
                    * TODO: Modify the exec command so that the file is deleted within the pod after it has been imported.

        * Implementation notes:
            * Implementation Outline
                1. [Done] Determine the master Pod for a given service instance name/namespace
                    * **Important**: For clustered instances, before copying the file, it must be determined which Pod is the master-Pod as the role assignment may change over time.
                    * The master pod is the pod with the following label: `a8s.a9s/replication-role=master`
                    * `kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'`
                    * [Done] Implement in `pg/a8s_pg.go`
                        * [Done] BUG: Creating a service instance named `solo` should not print output containing the name `clustered-0`.
                        * [Done] `FindPrimaryPodOfServiceInstance`
                            * [Done] In `k8s/kubectl.go` implement `FindFirstPodByLabel`
                                * Implement a more generic version `Kubectl` being a variadic function just like `Command` is.
                    
                1. [Done] Upload file to pod
                    * The container to copy the file to is called `postgres`
                    * The file should be uploaded to the pod's `tmp` folder
                    * For `kubectl cp` to work, the `tar` command must be present in the target pod.
                    * Implement copy in `kubectl.go`
                    * In `k8s/kubectl.go` implement
                        * `KubectlUploadFileToPod`
                        * `KubectlUploadFileToTmp`
                        * `KubectlDeleteTmpFile`
                        * `KubectlDeleteFile`
                        * `KubectlExec`
                        * `KubectlCp` 
                1. [DONE] Apply file by executing `psql`
                    * Implement apply in `a8s_pg.go`
                1. [DONE] Delete file
                    * Implement copy in `kubernetes_workload.go`
                1. [DONE] Test manually
                1. [Next] Add tests to the e2e test suite
                1. `a9s pg apply --file` should warn if a service-instance cannot be found
                1. `a9s pg apply` should demand mandatory params without defaults for `-f` and `-i`
    * Feature: Restore
        * The implementation plan is similar to creating the backup.
        * DONE: Create command `a9s create pg restore ...`
        * DONE: Generate a YAML manifest
        * DONE: Apply the YAML manifest
        * DONE: Test manually
        * DONE: Add tests to the e2e test suite
        * This completes the backup / restore cycle.

# v0.8.0 and Older

* Chore: Write a testSuite to run end-to-end tests on a local machine using the `a9s`-cli applying all major usecases for both the kind and minikube providers.

* Sub command to delete all demo resources.
    * Remove everything (incl. config files)
        * e.g. `a9s demo delete --all`
