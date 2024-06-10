# Backlog

## Next

* Refactor `EstablishBackupStoreCredentials`
  * Set accesskey and secretaccesskey automatically.  
  * Make generating the access and secret key the default but allow passing it as an argument for AWS and others
    * Requires updating the Readme and Tutorial    

* [**In Progress**] Add minio deployment
    * [DONE] Apply manifests
    * [DONE] Implement WaitForMinioToBecomeReady
            * Read credentials from Secret
                * Can't use the existing secret for two reasons
                    1. Can't use the secret from `a8s-system` in another (`minio-dev`) namespace
                    2. The `a8s-system` secret is in file format: keys are filenames and values are content, a volume mount would be required. While this is possible, it is not what is desired for the config script.
    * Implement a K8s Job to configure minio    
        * Add configmap for `minio_config.sh` script
        * Create job defninition `minio-config-job.yaml`
        * Add configmap and scripts to a8s-demo repo
        * Add Job to a9s-cli creation of minio
        

* [DONE] Add params for Endpoint and Pathstyle
    * [DONE] Add params 
    

* Update https://docs.a9s-cli.anynines.com/ to v0.12.0
* Update https://docs.a9s-cli.anynines.com/ to v0.13.0

* In versioned docs the install command under "installing the CLI" refers so the lastest version. This is misleading as the downloaded version should match the version of the doc.


* [**In Progress**] Epic: Minio as an alternative to AWS S3
    * Enable a8s-backup-manager
        * [DONE] Implement support for minio by allowing to set custom s3 endpoint and enable pathStyle
        * [Waiting] Release updated container image version of a8s-backup-manager
        * Update a8s-deployment to use new version of a8s-backup-manager
        * Update a9s-CLI to use new version of a8s-deployment && a8s-backup-manager
    * Implement CLI functionality    
        
* [Epic] Make Release v0.13.0
    * Update Backlog and Implementation log
    * Update Changelog
    * Update Readme
    * Update a9s CLI Tutorial at https://docs.a9s-cli.anynines.com/

* [DONE] Generate configs before creating a cluster, this saves time if something is wrong with configs as the cluster creation is likely to also have problem then.
    * Move creating a cluster out of `CheckPrerequisites` as it is an odd-place to create a cluster.
        * Split env and config checks from 


* [DONE] Suggest a more meaningful working directory.
    * Using the current directory as the default directory is often not a good choice. We want the default values to be meaningful.
    * Decide about default directory
        * How about $HOME/.a9s/ ?
            * Why hidden?

* [DONE] New default config file location.
    * It is confusing for a user to use the `a9s` CLI but then have an `.a8s` config file.    
    * $HOME/.a9s


## Unassigned

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


