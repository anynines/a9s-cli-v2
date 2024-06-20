# Implementation Notes

Implementation notes have been written during the implementation. They are collected here as
they may contain information worthwile to persistest. However, be aware that these notes may
represent ideas that have not been implemented or that changes may have been applied. In other words:
**do not expect implementation notes to be in sync with the implementation**.

The implementation notes may document patterns on how certain features are developed by listing individual steps. This may help new developers to find a scaffold to start with when entering the project.

## v0.13.0

* Remove `mc` as this command is not really necessary since the config happens as a Job which uses mc internally


* BUGFIX: Minio apply -k after the creation of its namespace requires waiting for the serviceaccount "default" in the minio-dev namespace to become ready
    * Otherwise the following error message is seen on some systems (e.g. the AWS CI linux server): `+Error from server (Forbidden): error when creating "/home/ubuntu/a9s/a8s-demo/minio": pods "minio" is forbidden: error looking up service account minio-dev/default: serviceaccount "default" not found`

* BUGFIX: WaitForA8sSystemToBecomeReady does not detect that a8s-backup-manager is in CrashLoopBackOff

* Epic: Minio as an alternative to AWS S3
    * Steps:
        * Manually deploy minio into a local cluster and record all command & steps
            * Install minio client
                * `brew install minio-mc`
            * Install minio operator
                * https://min.io/docs/minio/kubernetes/upstream/index.html
                    * Don't install the regular operator. Instead, install the non-prod minimal version of minio. We neither need nor want multi-tenancy.
                    * Remove the following lines:
                        * nodeSelector:
                            * kubernetes.io/hostname: kubealpha.local # Specify a node label associated to the Worker Node on which you want to deploy the pod.
                *   `kubectl apply -f minio-dev.yaml`
                * Start a kubectl proxy: `kubectl port-forward pod/minio 9000 9090 -n minio-dev`
                * Create an alias for the a8s-demo-minio target:                
                    `mc alias set a8s-demo-minio http://127.0.0.1:9000 minioadmin minioadmin`
                * Test the communication with the target: `mc admin info a8s-demo-minio`
                * Create minio user `a8s-user`:
                    `mc admin user add a8s-demo-minio a8s-user a8s-password`
                * Create a bucket `a8s-backups`: `mc mb a8s-demo-minio/a8s-backups`
                    * Assign bucket a policy
                        * localhost:9090 > minioadmin:minioadmin
                        * > Identity > users > a8s-user > Assign Policy > ReadWrite
                * When using the minikube stack, the `mc` command is required
            * Introduce stack-dependencies
        * UX:
            * Make minio the default storage option
                * When minio is selected, we don't need to ask for backup credentials, this is only necessary when S3 is selected.
                * `a9s create stack|cluster a8s --backup-provider=AWS`
                * `a9s create stack|cluster a8s --backup-provider=minio` (default)
    

* [**Abandonded**] Allow using a custom provided container image for the a8s-backup-manager for `create cluster a8s` and `create stack a8s`.
    `a9s create cluster a8s --backup-manager-image=myuser/myimage:mytag` 
    * Implement the param for
        * [DONE] `create stack`
        * [DONE] `create cluster`
    * Locate where the param needs to be applied
        * Try to use `kustomize` to overwrite the default image
            * spec
                * versions
                    * "v1beta3"
                        * schema
                            * openAPIV3Schema
                                * image                                    
            * line 639: `public.ecr.aws/w5n9a2g2/a9s-ds-for-k8s/dev/backup-manager:2616f22c4fe670541c3c78131ae018902f8471bf`
        * Using `kubectl apply -k deploy/a8s/manifests --dry-run=server -o yaml | bat -l yaml` the existing container image can be replaced using the `images` functionality of `kustomize` as described in (here)[https://github.com/kubernetes-sigs/kustomize/blob/master/examples/image.md]. Note that `--dry-run=client` won't show the substituted container image as the patching is done in the interaction with the live Kubernetes cluster.
    * Reason: This feature is not worth the effort. Applying a custom image is not frequently demanded and for local development and modified kustomization.yaml will cause the same effect.


## v0.12.0

* [ARCHITECTURE] Install a8s PG on an existing cluster
    * Decide which command/verb to use
        * [DONE] `a9s create stack`
            * Applies the a8s stack to the current k8s cluster
                * The following cluster/context/namespace is selected:
                    Do you want to apply the a8s stack to this cluster?
    * Make the context/namespaces configurable        
        * Streamline the UX for both `create cluster` and `create stack`        
                * Point out that `-c` can be used to specify a context / clustername
        * Update the readme
            * Add `create stack` documentation
    * Write a tutorial on how to apply a stack to an existing AWS cluster
        * Use `-c` option to point to the right context

* Make the default namespace/context not contain the word `demo`.

## v0.12.0

* [ARCHITECTURE] Install a8s PG on an existing cluster
    * Decide which command/verb to use
        * [DONE] `a9s create stack`
            * Applies the a8s stack to the current k8s cluster
                * The following cluster/context/namespace is selected:
                    Do you want to apply the a8s stack to this cluster?
    * Make the context/namespaces configurable        
        * Streamline the UX for both `create cluster` and `create stack`        
                * Point out that `-c` can be used to specify a context / clustername
        * Update the readme
            * Add `create stack` documentation
    * Write a tutorial on how to apply a stack to an existing AWS cluster
        * Use `-c` option to point to the right context

* Make the default namespace/context not contain the word `demo`.

## v0.11.1

* Create and publish `a9s` binaries for supported OSes
    * [DONE] Run test suite on linux
        * Perform cold test run
        * Run test suite
    * [DONE] Run test suite test on macos
        * Perform cold test run
        * Run test suite
* [DONE] Create CI/CD pipeline that
    * [DONE] Creates binaries for each supported OS
    * [SOMEWHATDONE] Runs tests for each supported OS
        * UPDATE: 
            * Can't do. To0 much effort.
            * Compromize: only linux is CI tested. MacOS is tested manually.
    * [SOMEWHATDONE] Publishes binaries for each supported OS
        * Upload binary
        * Publish docs / changelog


## v0.11.0

* Rename `a9s create demo a8s` to `a9s create local-dev-env` or something similar.
    * [DONE] Decide which word to use instead of `demo`
        * `a9s create stack` 
        * `a9s create environment` &&|| `a9s create env`
        * `a9s create cluster`
    * [DONE] Rename to `a9s create cluster`
        * Rename create cmd
        * Rename delete cmd
        * Adapt Readme
        * Adapt unit tests
        * Adapt e2e tests
    * Go through the entire `create cluster` use case and check of occurences of the word "demo"
    * Rename the "demo" package?
    * Write changelog
    * Adapt implementation notes


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
