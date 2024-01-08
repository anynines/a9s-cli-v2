# Backlog
* Next release: backup/restore    
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
                    
                1. **[Next]** Upload file to pod
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
                1. Apply file by executing `psql`
                    * Implement apply in `a8s_pg.go`
                1. Delete file
                    * Implement copy in `kubernetes_workload.go`
            



    * Feature: Restore
        * The implementation plan is similar to creating the backup.
        * DONE: Create command `a9s create pg restore ...`
        * DONE: Generate a YAML manifest
        * DONE: Apply the YAML manifest
        * DONE: Test manually
        * DONE: Add tests to the e2e test suite
        * This completes the backup / restore cycle.
    * Backup: A failed backup should be indicated to the user.
    * Restore: A failed restore should be indicated to the user.

* `a9s pg apply --file` should warn if a service-instance cannot be found

* Backup/Restore: The WaitForKubernetesResource function should indicate the current status cycling through all states with Status = true (scheduled, complete, ...)
* Backup: The create backup command should verify whether the given service instance exists.
* Restore: The create restore command should verify whether both the given backup and service instance exists.

Feature: Ensure that all commands accept and correctly apply a custom namespace for all postgres operations
    * Manually test
        * Create a service instance in a non-default namespace
        * Create a backup
        * Restore a backup
        * Delete the service instance
    * Write an rspec test scenario for the above

* CHORE: Use PrintVerbose to make output much more clean.

* BUG: Backups for non existing service instances shouldnt return success messages.
    * The event was `map[lastTransitionTime:2023-12-29T09:18:43Z message:Backup Completed reason:Complete status:True type:Complete]`
    * The bug may exist in the a8s backup manager

* CHORE: Check if makeup.WaitForUser(demo.UnattendedMode) is used consistently for all new commands instance & backup

* Deleting Backups and Restore CRs
    * For completeness: Create command `a9s delete pg backup ...`
    * For completeness: Create command `a9s delete pg restore ...`

* Chore: Write a testSuite to run end-to-end tests on a local machine using the `a9s`-cli applying all major usecases for both the kind and minikube providers.

* Sub command to delete all demo resources.
    * Remove everything (incl. config files)
        * e.g. `a9s demo delete --all`

* When executing a9s create demo a8s for the first time, the infrastructure-region should be queried as a user input instead of being a default-parameter. The probability is too high that the user choses a non-viable default option instead of providing a valid region.


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

* Don't use the `default` namespace, instead create a demo namespace, e.g. `a8s-demo`.
    * Provision a8s-pg into namespace

* Create binaries in a release matrix, e.g. using Go Release Binaries with Gihub Action Matrix Strategy
    * https://github.com/marketplace/actions/go-release-binaries
    * https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix

* Create S3 bucket with configs
    * Alternatively: Install a local storage provider, e.g. minio.
        * Costly dependency: add the local storage provider to the backup agent.
