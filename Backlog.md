# Backlog
* Next release: backup/restore
    * Feature: Restore
        * Create command `a9s create pg restore ...`
        * This completes the backup / restore cycle.

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

* Question: Should the demo a8s-pg execute the entire demo or just install the operator? Other commands could be: 
    * a8s-pg 
        * `create`
            * It's more idiomatic in Kubernetes for the verb to be the first command: `kubectl get pods` vs `kubectl pod get`.
        * `a9s pg instance`
            * `create`
                * `a9s pg instance create --isolation pod` > a8s PG
                * `a9s pg create instance --isolation pod`
                * `a9s pg instance create --isolation vm` > a9s PG
        * `a9s pg service-binding`
            * `a9s pg binding` 
            * `a9s pg sb`
        * `a9s pg backup`
        * `a9s pg restore`
    * a8s-pg-instance 
    * a8s-pg-app
    * Alternatively, the entire demo could be driven by the "assistent" asking the user questions, interactively.


* Don't use the `default` namespace, instead create a demo namespace, e.g. `a8s-demo`.
    * Provision a8s-pg into namespace

* Create binaries in a release matrix, e.g. using Go Release Binaries with Gihub Action Matrix Strategy
    * https://github.com/marketplace/actions/go-release-binaries
    * https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix

* Create S3 bucket with configs
    * Alternatively: Install a local storage provider, e.g. minio.
        * Costly dependency: add the local storage provider to the backup agent.
