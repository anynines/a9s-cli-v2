# Implementation Notes

Implementation notes have been written during the implementation. They are collected here as
they may contain information worthwile to persistest. However, be aware that these notes may
represent ideas that have not been implemented or that changes may have been applied. In other words:
**do not expect implementation notes to be in sync with the implementation**.

## v0.9.0

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