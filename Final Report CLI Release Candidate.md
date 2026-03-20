# Report for Ticket KLT-869 - a9s CLI Manual Release Process

<!-- a report on the functional correctness of the CLI as evaluated using the steps
from the tutorial. If additional steps were necessary to successfully execute
the tutorial, the report must document them as precisely as possible.

a report on the user experience from the perspective of a fresh user. The report
must contain but is not limited to sections on the following topics:

Noisiness of the CLI’s output (Does it provide the right amount of output?)

Interface-Design of the CLI (How does it feel to use it?) -->.

- [Functional Correctness](#functional-correctness)
  - [Problems with CLI logic](#problems-with-cli-logic)
    - [The command `a9s create cluster klutch` only works for users logged into JF's personal AWS account](#the-command-a9s-create-cluster-klutch-only-works-for-users-logged-into-jfs-personal-aws-account)
    - [The command `a9s create cluster klutch` did not set the Control Plane cluster up correctly](#the-command-a9s-create-cluster-klutch-did-not-set-the-control-plane-cluster-up-correctly)
    - [The command `a9s create cluster klutch` is not idempotent](#the-command-a9s-create-cluster-klutch-is-not-idempotent)
    - [The command `a9s create cluster klutch` does not check the Quota for enough Elastic IPs](#the-command-a9s-create-cluster-klutch-does-not-check-the-quota-for-enough-elastic-ips)
    - [The command `a9s create cluster klutch` deploys an old Dataservices Configuration Package](#the-command-a9s-create-cluster-klutch-deploys-an-old-dataservices-configuration-package)
    - [The command `a9s delete cluster klutch` leaves resources behind](#the-command-a9s-delete-cluster-klutch-leaves-resources-behind)
  - [Problems with components deployed by the CLI](#problems-with-components-deployed-by-the-cli)
    - [The conditions on a claim does not reflect the condition of the managed resources](#the-conditions-on-a-claim-does-not-reflect-the-condition-of-the-managed-resources)
    - [The command `a9s delete klutch pg backup` left orphaned resources behind](#the-command-a9s-delete-klutch-pg-backup-left-orphaned-resources-behind)
- [User Experience](#user-experience)
  - [Noisiness](#noisiness)
    - [always printing YAML manifests](#always-printing-yaml-manifests)
    - [the formatting of shell commands](#the-formatting-of-shell-commands)
    - [silence during Cognito provisioning](#silence-during-cognito-provisioning)
  - [Applicability](#applicability)
    - [sometimes not enough information is provided before asking for user confirmation to proceed](#sometimes-not-enough-information-is-provided-before-asking-for-user-confirmation-to-proceed)
    - [asking users to authorize `kubectl apply -f -` without showing the manifests](#asking-users-to-authorize-kubectl-apply--f---without-showing-the-manifests)
    - [N2H: estimated duration of step](#n2h-estimated-duration-of-step)
  - [Interface-Design](#interface-design)
    - [command path feels overly nested](#command-path-feels-overly-nested)
    - [the strict separation between local and remote path feels unjustified](#the-strict-separation-between-local-and-remote-path-feels-unjustified)
    - [having the service name in the command path for referencing resources feels unjustified](#having-the-service-name-in-the-command-path-for-referencing-resources-feels-unjustified)
  - [Correctness](#correctness)

## Functional Correctness

### Problems with CLI logic

#### The command `a9s create cluster klutch` only works for users logged into JF's personal AWS account

Details | The command references a Helm chart and an operator image that are hosted in a private ECR.
--- | ---
Consequence | Anyone who wants to use the CLI has to either be logged into that specific AWS account or at least have access to the source code for the Helm chart and the tenant operator to build these images themselves.
Short-Term Fix | Build interim images and uploaded them to AWS dev account for the a8s team.
Suggested Long-Term Fix | Build official images and upload them to a public repo in the ECR of the a8s team's AWS production account.

#### The command `a9s create cluster klutch` did not set the Control Plane cluster up correctly

Details | EKS add-on for provisioning AWS EBS Volumes was missing.
--- | ---
Consequence | a8s pg instances did not become ready because they were stuck waiting for their persistent volumes to be provisioned
Short-Term Fix | Manually run eksctl command to create required IAM service account and install required add-on to cluster.
Suggested Long-Term Fix | Add steps to CLI for adding IAM service account and installing add-on during CP cluster creation.

#### The command `a9s create cluster klutch` is not idempotent

Details | The command always checks, if the difference between the Elastic IPs in use and the Elastic IP Quota is large enough to provision new IPs, regardless of whether new IPs are actually needed.
--- | ---
Consequence | If the necessary NATs already exist and the quota is already filled, then the command will fail with the message that not enough IPs are available.
Short-Term Fix | Move check for the new IPs further back in the CLI code on my feature branch.
Suggested Long-Term Fix | Verify, that the new place for the CLI check is correct, then merge the change into the main branch.

#### The command `a9s create cluster klutch` does not check the Quota for enough Elastic IPs

Details | When checking, whether enough Elastic IPs are available to deploy the cluster, the command checks for 3 IPs (one per NAT Gateway), but when deployed each NAT will use 2 Elastic IPs, leading to a total of 6 IPs in use.
--- | ---
Consequence | Users might run into issue, as their IP quotas are filled faster than they expected.
Short-Term Fix | Increase IP quota for dev AWS account to be high enough to accommodate for 12 EIPs.
Suggested Long-Term Fix | Either look into what makes the NATs provision 2 IPs per Gateway and change it to only 1 IP per NAT or update the logic to check for 6 IPs instead of 3.

#### The command `a9s create cluster klutch` deploys an old Dataservices Configuration Package

Details | The version deployed is `v1.3.0`, which was released in January last year. In December, we released version `v1.4.0`, which contains some changes to the way Compositions are generated.
--- | ---
Consequence | A user who sticks to a8s PG will not notice a difference, as the changes to the a8s PG Compositions are all under the hood and don't change the behaviour of the affected resources. A user who wants to use the `provider-anynines` for provisioning instances from an a9s DS Service Broker will however lose out on some features added to the compositions in the meantime (e.g. custom parameters for some data services, support for a9s KeyValue etc.).
Short-Term Fix |Update `configPackageManifestUrl` from `v1.3.0` to `v1.4.0` on feature branch and manually deploy the `patch-and-transform` Crossplane function
Suggested Long-Term Fix | Merge change to `configPackageManifestUrl` into main branch and extend `create cluster` command to also deploy the `patch-and-transform` Crossplane function

#### The command `a9s delete cluster klutch` leaves resources behind

Details | The Hosted Zone, the Cognito User Pools hub-klutch and tenant-\<tenant-id\>-klutch and the secret klutch/\<tenant-id\>/oidc-client are still left after running `a9s delete cluster klutch control-plane`.
--- | ---
Consequences | The hosted zone will cost $0.50 per month, the secret will cost $0.40 per month while the Cognito User Pools will not incur further costs unless a machine user request a new access token and even then they are fractions of a cent per token. Because these costs are very small and the Quotas of these services (500/Account for Hosted Zones, 1,000/Region for User Pools, 500,000/Region for Secrets) are high enough that it would take a lot of orphaned resources to clog them up, the impact of these orphans is not that high.
Short-Term Fix | keep the resources around while testing of the release candidate is ongoing, delete them all together after it's done
Suggested Long-Term Fix | decide, which of these resources are okay to be left behind and which of them need to be cleaned up. Adapt CLI logic to clean up the ones that are not fit to be left behind, document the ones left behind and/or output this information when running the `a9s delete cluster klutch` command.

### Problems with components deployed by the CLI

#### The conditions on a claim does not reflect the condition of the managed resources

Details | As soon as the managed resources can be provisioned, the claim gets marked as "Ready", regardless of if the managed resources themselves become "Ready" or "Available". Updating to the newest version of the Configuration Package for the a8s Framework did not solve this issue.
--- | ---
Consequences | When using the "Ready" condition to wait for a provisioned instance, servicebinding, backup or restore to be usable (e.g. via `kubectl wait`) a user might get the impression, that their provisioned resource would be usable even when it is not. - When deleting a claim, Crossplane does not wait for the managed resources to be deleted, potentially leading to orphaned resources stuck in deletion, which are invisible to the Workload cluster.
Short-Term Fix | Accepted the behaviour and worked around it by monitoring the `.status.managed` subresource as well as manually checking after deleting claims, that the managed resources are cleaned up as well.
Suggested Long-Term Fix | Update a8s CompositeResourceDefinitions to dynamically update the "Ready" condition based on the state of the managed resources as is already done for the a9s CompositeResourceDefinitions.

#### The command `a9s delete klutch pg backup` left orphaned resources behind

Details | When executing the command, the Claim and the Composition were deleted, but the managed resource was unable to be removed. The cause is a bug in the `a8s-backup-manager`, caused by a hardcoded S3 Endpoint URL. Because of this the controller could not delete the backup file and therefore could not clear the backup MR for deletion.
--- | ---
Consequence | Users who want to store their backups in MinIO or other Object Stores apart from AWS S3 will have to Manually clean out orphaned backups.
Short-Term Fix | Opened [PR against a8s-backup-manager repo](https://github.com/anynines/a8s-backup-manager/pull/99), build interim image `378836732719.dkr.ecr.eu-central-1.amazonaws.com/a8s-backup-manager:fix-minio-deletion-0`, Manually paste interim image into deployment manifests in local copy of `a8s-deployment` repo.
Suggested Long-Term Fix | Verify that the code in the PR is up to our standards, merge it, build a new image with the fix, swap out the image in the `a8s-deployment` repo, cut a new release.

## User Experience

### Noisiness

#### always printing YAML manifests

##### Problem

It feels overwhelming and disorienting to get one or sometimes even multiple yaml manifests
thrown at you, even without setting the `--verbose` flag.

##### Suggested Solution: On-demand paginated yaml

- give user the option to ask for manifests before confirming
- only print manifest(s) when user asks for them
- don't print them in the regular terminal but show them paginated
- possible but less necessary polish:
  - implement flag for automatically printing manifests

    OR

  - automatically print manifests when `--verbose` flag is set

##### Reasoning

- Users who want to inspect the manifests can still do that before they are
applied
- Users who are only interested in the high-level progress and process
of the CLI can easily follow along without having to scroll past blocks of
manifests

#### the formatting of shell commands

##### Problem

- following along with the high-level progress of the CLI is hard because the
frames around the commands take up so much terminal space
- copy-pasting multi-line-commands from the CLI output for terminal use is
  annoying because of elements inserted by the formatting (empty line,
  side-borders of the frame, spaces for centring) which have to be deleted
- Disclaimer: strictly my own opinion/taste
  - the formatting gives me the vibe of a hobby-project instead of looking
    professional, which is probably not something we want a product of ours to
    feel like

##### Suggested Solution: change formatting to take up less space and be copy-paste-friendly

- print the command in reverse formatting or in colour
- only leave one line above and below it empty
- N2H: show the eye-catching, formatted version in a pager utility, print
  unformatted message like `executing <command>...` into regular terminal

```bash
REVERSE_FORMATTING="\e[7m"
RESET_FORMATTING="\e[0m"
echo -e "The following command will be executed for you:

${REVERSE_FORMATTING}eksctl create cluster --n my-awesome-cluster${RESET_FORMATTING}

Press <Enter> to continue or <CTRL>+C to abort."
```

##### Reasoning

- copy-pasting a multi-line command is no longer obstructed by the formatting
- following the high-level process via the terminal is now easier
- users who run the CLI in a terminal at the edge of their screen and multitask
while the CLI does its thing will still have their attention caught by the formatting
- using a pager utility will allow for more eye-catching formatting if necessary
  while keeping the general output easy to follow

#### silence during Cognito provisioning

##### Problem

This is a process that took 35 minutes, the AWS Dashboard showed the Cognito
User Pool as provisioned after 10 minutes and I had no idea, what was happening
during the other 25 minutes, as the CLI did not give any updates during that
time.

##### Suggested Solution

Give updates from time to time in the terminal to make it clear **that** things
are happening and **what** is happening to make the experience less frustrating
for the user.

### Applicability

<!-- Applicability of the CLI’s output: Does its output provide the right
information? -->

#### sometimes not enough information is provided before asking for user confirmation to proceed

##### Problem

- being asked to confirm or abort without a proposed/planned action feels
  bizarre when a user is not told what is about to happen
- sometimes the manifests are provided **after** the user has already consented
  to the action, which feels like adding insult to injury
- sometimes the user is asked for approval, then shown the manifests and then
  asked for approval again, which feels very clunky
- sometimes after the user approves continuing, is shown a manifest and approves
  continuing again, the CLI applies multiple manifests or applies a manifest and
 executes an additional step (like provisioning cloud resources) which feels very clunky

##### Examples

- I have no idea what I'm consenting to here

  ```bash
  ✅ The Crossplane components appear to be ready.

  Press <ENTER> key to continue or <CTRL>+C to abort.
  ```

- ...yes, that is why I've called the `a9s create klutch control-plane` command?
  This should have been more specific.

  ```bash
  ╭───────────────────────────────────────────────────────────────────────╮
  │                                                                       │
  │    Applying Klutch Control Plane to the current Kubernetes cluster    │
  │                                                                       │
  ╰───────────────────────────────────────────────────────────────────────╯
  Let's install the Klutch control plane into your current Kubernetes cluster...

  Press <ENTER> key to continue or <CTRL>+C to abort.
  ```

##### Suggested Solution

- make sure that any step that waits for authorization is preceded by a
  description of what the CLI is about to do
- make sure that the description encompasses or implies all actions that follow the users approval

#### asking users to authorize `kubectl apply -f -` without showing the manifests

##### Problem

- when the user is not shown the manifests before applying and has no option to
  get the displayed they can't know what is about to happen
- asking them to approve of the command feels pointless

##### Suggested Solution

- either don't ask users to approve of running `kubectl apply -f -` or give them
  the option to display the manifests beforehand

#### N2H: estimated duration of step

##### Problem

A fresh user might not know how long it usually takes to provision certain
resources on AWS, e.g. a Cognito user pool. Not getting an estimate means they
can't commit to doing other work while these resources are getting provisioned.

##### Suggested Solution

- estimate usual durations for all resources that need to be (de)provisioned on AWS
- if an operation is usually a matter of seconds, then no output is needed
- if an operation is usually a matter of minutes, then the estimation is printed
  to the terminal to let the user know and potentially use that time for
  something else
- example:

  ```bash
  ℹ️  No Cognito settings provided. Provisioning Cognito (region: eu-central-1, prefix: hub)...
  ℹ️  Estimated time: 30-35min
  ```

### Interface-Design

#### command path feels overly nested

##### Problem

- holding a mental model of the command structure is cumbersome because of its depth
- when calling the CLI myself instead of copy-pasting commands from the tutorial
  I constantly had to go back and check whether I forgot a piece of the command
  structure and whether all the pieces are called in the correct order

##### Solution: limit command structure depth

- two levels seems to be the standard
  - `aws ec2 describe-vpcs`
  - `eksctl create cluster`
  - `vcluster delete cluster`
  - `kubectl get pods`
  - `tkn pipeline start`
- examples (`--target [remote|local]` could also be something else like `--infrastructure [aws|kind]`)
- operation before operand
  - `a9s create control-plane --target remote`/`a9s create control-plane --target local`
  - `a9s create-control-plane remote`/`a9s create-control-plane local`
  - `a9s create pg-instance --target remote`
  - `a9s create pg-instance remote`
- operation after operand
  - `a9s remote create-control-plane`/`a9s local create-control-plane`
  - `a9s remote create-pg-instance`
  - `a9s control-plane create  --target remote`/`a9s control-plane create --target local`
  - `a9s pg-instance create --target remote`

#### the strict separation between local and remote path feels unjustified

- it feels strange that there is such a separation between deploying a pg
  instance for a local and for a remote cluster
- understandable for creating and deleting the clusters themselves, but
  as a user I'd expect the CLI to be able to deploy an instance by just being
  told `a9s create pg` and to infer everything else it needs to know from context
- moving from a strict separation in the second level of the command structure
  to a flag to be supplied solely for operations on clusters would help with the
  problem of command depth

#### having the service name in the command path for referencing resources feels unjustified

- requiring `pg` to be added before `backup`, `restore` and `servicebinding` in
  `a9s [create|delete] pg [backup|restore|servicebinding]` feels unjustified
- it's understandable that the name of the service is required when creating the
  service instance itself, but as a user I'd expect the CLI to be able to infer
  the service kind from the instance reference passed via the
  `--service-instance`/`-i` parameter
- removing the service name from the command path would help with the problem of
  command path depth
- removing the service name and the `klutch` for remote environments from the
  command path would solve the problem of command path depth, as then all
  possible commands would have at most 2 levels and stick to the schema `a9s
  <operation> <operand>`

### Correctness

<!-- Correctness of the CLI: Does it do the right things in the right way? -->

I'm not well-versed enough in AWS best practices in order to know whether the
things which work are done in the right way.

For the things which don't work, either because the CLI misses a necessary
operation or because it executes operations incorrectly, see the section
[Problems with CLI logic](#problems-with-cli-logic).
