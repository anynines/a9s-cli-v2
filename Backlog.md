# Backlog

## Next

* Update https://docs.a9s-cli.anynines.com/ to v0.12.0


* Suggest a more meaningful working directory.
    * Using the current directory as the default directory is often not a good choice. We want the default values to be meaningful.
    * Decide about default directory
        * How about $HOME/.a9s/ ?
            * Why hidden?

* New default config file location.
    * It is confusing for a user to use the `a9s` CLI but then have an `.a8s` config file.
    * Decide
        * $HOME/.a9s/cli.yml  
        * $HOME/.a9s

* [Optional] Apply chmod 600 to the config file

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
        * Enable a8s-backup-manager
            * [DONE] Implement support for minio by allowing to set custom s3 endpoint and enable pathStyle
            * [Waiting] Release updated container image version of a8s-backup-manager
            * Update a8s-deployment to use new version of a8s-backup-manager
            * Update a9s-CLI to use new version of a8s-deployment && a8s-backup-manager
        * Implement CLI functionality    
            * When using the minikube stack, the `mc` command is required
                * Introduce stack-dependencies
            * UX:
                * Make minio the default storage option
                    * When minio is selected, we don't need to ask for backup credentials, this is only necessary when S3 is selected.
                    * `a9s create stack|cluster a8s --backup-provider=AWS`
                    * `a9s create stack|cluster a8s --backup-provider=minio` (default)
                * 
        * Update Backlog and Implementation log
        * Update Changelog
        * Update Readme
    * Update a9s CLI Tutorial at https://docs.a9s-cli.anynines.com/
* Epic: Custom essential container images to fascilitate development
    * For each essential container image allow providing a custom container image
        * a8s-backup-manager
        * ...


* [Question] Remove ?
    * Should there be a remove option?
        * Yes
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


    * Tutorial: 
        * https://medium.com/@prateek.malhotra004/step-by-step-guide-creating-an-amazon-eks-cluster-with-aws-cli-edab2c7eac41        
            * https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
            * `eksctl utils write-kubeconfig --cluster=<name> [--kubeconfig=<path>] [--set-kubeconfig-context=<bool>]`



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

