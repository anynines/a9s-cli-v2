# Report for Ticket KLT-869 - a9s CLI Manual Release Process

<!-- a report on the functional correctness of the CLI as evaluated using the steps
from the tutorial. If additional steps were necessary to successfully execute
the tutorial, the report must document them as precisely as possible.

a report on the user experience from the perspective of a fresh user. The report
must contain but is not limited to sections on the following topics:

Noisiness of the CLI’s output (Does it provide the right amount of output?)

Applicability of the CLI’s output (Does its output provide the right
information?)

Interface-Design of the CLI (How does it feel to use it?)

Correctness of the CLI (Does it do the right things in the right way?) -->.

- [Functional Correctness](#functional-correctness)
  - [The command `a9s create cluster klutch` only works for users logged into JF's personal AWS account](#the-command-a9s-create-cluster-klutch-only-works-for-users-logged-into-jfs-personal-aws-account)
  - [The command `a9s create cluster klutch` did not set the Control Plane cluster up correctly](#the-command-a9s-create-cluster-klutch-did-not-set-the-control-plane-cluster-up-correctly)
  - [The command `a9s create cluster klutch` is not idempotent](#the-command-a9s-create-cluster-klutch-is-not-idempotent)
  - [The command `a9s create cluster klutch` does not correctly check, whether the Quota for Elastic IPs is already filled](#the-command-a9s-create-cluster-klutch-does-not-correctly-check-whether-the-quota-for-elastic-ips-is-already-filled)
  - [The "Ready" condition on a claim does not reflect the condition of the managed resources](#the-ready-condition-on-a-claim-does-not-reflect-the-condition-of-the-managed-resources)
  - [The command `a9s delete klutch pg backup` left orphaned resources behind](#the-command-a9s-delete-klutch-pg-backup-left-orphaned-resources-behind)
  - [The command `a9s create cluster klutch` deploys an old version of the Configuration Package for the a8s DS Framework](#the-command-a9s-create-cluster-klutch-deploys-an-old-version-of-the-configuration-package-for-the-a8s-ds-framework)
  - [Remaining Resources after Clean Up](#remaining-resources-after-clean-up)
- [User Experience](#user-experience)
  - [Noisiness](#noisiness)
  - [Applicability](#applicability)
  - [Interface-Design](#interface-design)
  - [Correctness](#correctness)

## Functional Correctness

### The command `a9s create cluster klutch` only works for users logged into JF's personal AWS account

#### Details

The command uses a Helm chart and an operator image, that are hosted in that
account's private ECR.

#### Consequence

Anyone who wants to use the CLI has to either be logged into that specific AWS
account or at least have access to the source code for the Helm chart and the
tenant operator to build these images themselves.

#### Short-Term Fix

Build interim images and uploaded them to AWS dev account for the a8s team.

#### Long-Term Fix

Build official images and upload them to a public repo in the ECR of the a8s
team's AWS production account.

### The command `a9s create cluster klutch` did not set the Control Plane cluster up correctly

#### Details

EKS add-on for provisioning AWS EBS Volumes was missing.

#### Consequence

a8s pg instances did not become ready.

#### Short-Term Fix

Run eksctl command to add add-on to cluster.

#### Long-Term Fix

Add step to CLI for adding add-on during CP cluster creation.

### The command `a9s create cluster klutch` is not idempotent

#### Details

The command always checks, if the difference between the Elastic IPs in use and
the Elastic IP Quota is large enough to provision new IPs, regardless of whether
new IPs are actually needed.

#### Consequence

If the necessary NATs already exist and the quota is already filled, then the
command will fail with the message that not enough IPs are available.

#### Short-Term Fix

Move check for the new IPs further back in the CLI code.

#### Long-Term Fix

Verify, that the new place for the CLI check is correct.

### The command `a9s create cluster klutch` does not correctly check, whether the Quota for Elastic IPs is already filled

#### Details

When checking, whether enough Elastic IPs are available to deploy the cluster,
the command checks for 3 IPs (one per NAT Gateway), but when deployed each NAT
will provision 2 Elastic IPs, leading to a total of 6 IPs in use.

#### Consequence

Users might run into issue, as their IP quotas are filled faster than they
expected.

#### Short-Term Fix

Increase IP quota for dev AWS account to be high enough to accommodate for 12
EIPs.

#### Long-Term Fix

Either look into what makes the NATs provision 2 IPs per Gateway and change it
to only 1 IP or update the logic to check for 6 IPs instead of 3.

### The "Ready" condition on a claim does not reflect the condition of the managed resources

#### Details

As soon as the managed resources can be provisioned, the claim gets marked as
"Ready", regardless of if the managed resources themselves become "Ready" or
"Available". Updating to the newest version of the Configuration Package for the
a8s Framework did not solve this issue.

#### Consequences

- When using the "Ready" condition to wait for a provisioned instance,
servicebinding, backup or restore to be usable (e.g. via `kubectl wait`) a user
might get the impression, that their provisioned resource would be usable even
when it is not.
- When deleting a claim, Crossplane does not wait for the managed resources to
be deleted, potentially leading to orphaned resources stuck in deletion, which
are invisible to the Workload cluster.

#### Short-Term Fix

None, accepted the behaviour.

#### Long-Term Fix

Update a8s CompositeResourceDefinitions to dynamically update the "Ready"
condition based on the state of the managed resources.

### The command `a9s delete klutch pg backup` left orphaned resources behind

#### Details

When executing the command, the Claim and the Composition were deleted, but the
managed resource was unable to be removed. The cause is a bug in the
`a8s-backup-manager`, caused by a hardcoded S3 Endpoint URL. Because of this the
controller could not delete the backup file and therefore could not clear the
backup MR for deletion.

#### Consequence

Users who want to store their backups in MinIO or other Object Stores apart from
AWS S3 will have to Manually clean out orphaned backups.

#### Short-Term Fix

Opened [PR against a8s-backup-manager
repo](https://github.com/anynines/a8s-backup-manager/pull/99), build interim
image
`378836732719.dkr.ecr.eu-central-1.amazonaws.com/a8s-backup-manager:fix-minio-deletion-0`,
Manually paste interim image into deployment manifests in local copy of
`a8s-deployment` repo.

#### Long-Term Fix

Verify that the code in the PR is up to our standards, merge it, build a new
image with the fix, swap out the image in the `a8s-deployment` repo, cut a new
release.

### The command `a9s create cluster klutch` deploys an old version of the Configuration Package for the a8s DS Framework

#### Details

The version deployed is `v1.3.0`, which was released in January last year. In
December, we released version `v1.4.0`, which contains some changes to the way
Compositions are generated.

#### Consequence

A user who sticks to a8s PG will not notice a difference, as the changes to the
a8s PG Compositions are all under the hood and done in a way to preserve
identical behaviour. A user who wants to use the `provider-anynines` for
provisioning instances from an a9s DS Service Broker will however lose out on
some features added to the compositions in the meantime (e.g. custom
parameters for some data services, support for a9s KeyValue etc.).

#### Short-Term Fix

Accept situation as-is

#### Long-Term Fix

Update `configPackageManifestUrl` from `v1.3.0` to `v1.4.0` and extend `create
cluster` command to also deploy the `patch-and-transform` Crossplane function

### Remaining Resources after Clean Up

The Hosted Zone, the Cognito User Pools hub-klutch and tenant-\<tenant-id\>-klutch
and the secret klutch/\<tenant-id\>/oidc-client are still left after running `a9s
delete cluster klutch control-plane`.

## User Experience

### Noisiness

#### Printing YAML manifests leads to noise

##### Problem

It feels overwhelming to get one or sometimes even multiple yaml manifests
thrown at you, even without setting the `--verbose` flag.

##### Suggested Solution

Instead of printing the whole yaml at once before asking for confirmation, give
the user a short description of what the CLI will try and achieve by applying
resource manifests with the **option** of showing the manifests for inspection
before allowing their application.

This could be implemented by telling the user to press a specific key (e.g.
Space) to display the manifests, `\<Enter\>` to apply them and `\<CTRL\>+C` to
abort (like it already does).
That way users still have the option of inspecting manifests before they are
applied but users who are only interested in the high-level progress and process
of the CLI can easily follow along without having to scroll past blocks of
manifests.

I'd also suggest, that if the user does press the
key to display the manifests they are not pasted into the terminal as-is but
displayed via a pager utility like `less`.

A possible but not as necessary polish option would be to implement a flag for
automatically printing the manifests or to automatically print the manifests
when the `--verbose` flag has been specified.
Whether these manifests are then better displayed via a pager utility or pasted
as-is is still an open question to me.

#### The formatting of the shell commands creates noise

##### Problem

Trying to follow along with the high-level progress of the CLI is hard when the
large frames around the commands take up so much space in the terminal window.

It is also annoying to copy and past multi-line-commands from the CLI output to
a different terminal, since the formatting inserts an additional linefeed that
has to be deleted, the side-borders of the frame have to be deleted and the
spaces the CLI adds for center-aligning the text have to be deleted.

Also, and this is very much my own taste, I must say I don't find the formatting
professional-looking at all, it gives me the vibe of a hobby-project, which is
probably not something we want to convey with a product of ours at all.

##### Suggested Solution

We do want the formatting to be somewhat eye-catching, since (especially for
steps that might take multiple minutes to resolve) users might run the CLI in a
terminal at the edge of their screen and multitask while the CLI does its thing.

But I think printing the command in reverse formatting or in color or with an
empty line above and below it will suffice for that while keeping the command
friendly to copy and not having it clog up the log in the terminal as much.

```bash
REVERSE_FORMATTING="\e[7m"
RESET_FORMATTING="\e[0m"
echo -e "The following command will be executed for you:

${REVERSE_FORMATTING}eksctl create cluster --n my-awesome-cluster${RESET_FORMATTING}

Press <Enter> to continue or <CTRL>+C to abort."
```

Since this eye-catching effect is only necessary until the command is approved by
the user, it is also worth considering to display the eye-catching version in a
pager utility and to just print `executing \<command\>...` in the regular
terminal, to keep the high-level overview even less affected

### Applicability

### Interface-Design

### Correctness
