# Backlog

## Next

* BUG: There can be an error message when trying to create a service instance instantly after installing the a8s system.
    * Error message: `+Error from server (InternalError): error when creating "/Users/jfischer/a9s/usermanifests/a8s-pg-instance-clustered.yaml": Internal error occurred: failed calling webhook "mpostgresql.kb.io": failed to call webhook: Post "https://postgresql-webhook-service.a8s-system.svc:443/mutate-postgresql-anynines-com-v1beta3-postgresql?timeout=10s": dial tcp 10.104.231.248:443: connect: connection refused`
    * Reproduction: Remove `sleep(10)` from `a9s_cli_spec.rb` and run e2e-tests

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
            
* [Question] Remove ?
    * Should there be a remove option?
        * Yes
    * Removes the a8s stack from the current k8s cluster


* [POSTPONED] Run test suite on windows
        * Perform cold test run
        * Run test suite

* CHORE: Harmonize variable declaration for params. Use package config. Maybe use viper.

* BUGFIX: `a9s create demo a8s` fails if minikube has never started before.

```sh
    minikube profile list -o json
{"error":{"Op":"open","Path":"/home/ubuntu/.minikube/profiles","Err":2}}
```

* Feature: In order to script the usage of a9s I want to pass all information to a9s create cluster a8s to avoid any user dialog including the creation of a work direction as well as bucket credentials.

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
        * `eksctl utils write-kubeconfig --cluster='my-eks-cluster' --region=us-west-2`
        * Activate EBS for dynamic volume provisioning
            * eksctl utils associate-iam-oidc-provider --region=us-west-2 --cluster=my-eks-cluster --approve
            * eksctl create iamserviceaccount \
                --name ebs-csi-controller-sa \
                --namespace kube-system \
                --cluster my-eks-cluster \
                --role-name AmazonEKS_EBS_CSI_DriverRole \
                --role-only \
                --attach-policy-arn arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy \
                --approve \
                --region=us-west-2
            * Then add the add-on "aws-ebs-csi-driver"
                * eksctl create addon --name aws-ebs-csi-driver --cluster my-eks-cluster
                --region=us-west-2
                --service-account-role-arn arn:aws:iam::${ACCOUNT_ID}:role/AmazonEKS_EBS_CSI_DriverRole --force

* `a9s delete cluster -p aws`
    `eksctl delete cluster --name my-eks-cluster --region us-west-2`

    * Tutorial: 
        * https://medium.com/@prateek.malhotra004/step-by-step-guide-creating-an-amazon-eks-cluster-with-aws-cli-edab2c7eac41        
            * https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
            * `eksctl utils write-kubeconfig --cluster=<name> [--kubeconfig=<path>] [--set-kubeconfig-context=<bool>]`

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

* Create binaries in a release matrix, e.g. using Go Release Binaries with Gihub Action Matrix Strategy
    * https://github.com/marketplace/actions/go-release-binaries
    * https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix

* Create S3 bucket with configs
    * Alternatively: Install a local storage provider, e.g. minio.
        * Costly dependency: add the local storage provider to the backup agent.
* CHORE: Switch to https://docs.github.com/de/get-started/using-github/github-flow over git flow

