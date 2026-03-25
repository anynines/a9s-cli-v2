# Notes for the QA Tutorial

## Feedback on the release candidate

### First Impression: No Verbose Flag Set

- `a9s create cluster klutch control-plane -p aws` is not idempotent - it
  timed out while attempting to pull the helm charts and when re-running it
  tried to allocate 3 additional EIPs even though it already allocated 3 EIPs in
  the first execution
  - fix: had to move the function call for `ensureElasticIPQuota` from
  `ensureNetworking` to `ensureNATs`, so that the IP quota is only checked in
  the case where new IPs need to be allocated
- the CLI only checks for 1 Elastic IP per NAT Gateway but ends ob provisioning
  2 per Gateway - either we have to adapt the check in the code or look into
  what causes the additional IP to be provisioned
- being asked to confirm or abort without a proposed/planned action feels
  bizarre, since I am not told what is about to happen - either give short
  description of the action being authorized or don't wait for confirmation

  - ```bash
    ╭───────────────────────────────────────────────────────────────────────╮
    │                                                                       │
    │    Applying Klutch Control Plane to the current Kubernetes cluster    │
    │                                                                       │
    ╰───────────────────────────────────────────────────────────────────────╯
    Let's install the Klutch control plane into your current Kubernetes cluster...

    Press <ENTER> key to continue or <CTRL>+C to abort.
    ```

    ...yes, that is why I've called the CLI

  - ```bash
    ✅ The Crossplane components appear to be ready.

    Press <ENTER> key to continue or <CTRL>+C to abort.
    ```

    I have no idea what I'm consenting to here

- Sometimes when I press enter, **then** the manifests are printed to the terminal

- Sometimes pressing enter leads to the manifests being printed and then me
being asked to confirm the deployment, which is *better* but not *good*

- Sometimes I get shown manifests and asked to confirm, then when I do confirm
  it applies the shown manifests, tells me a command is about
  to be executed **and** immediately executes that command. This is inconsistent and makes me wonder why the command is printed in the first place

  <details>
    <summary><strong>Example</strong></summary>

    ```bash
     Deploying the Kubernetes Crossplane provider config...

     Applying the following manifests:
    apiVersion: kubernetes.crossplane.io/v1alpha1
    kind: ProviderConfig
    metadata:
      name: kubernetes-provider
    spec:
      credentials:
        source: InjectedIdentity

     Press <ENTER> key to continue or <CTRL>+C to abort.


     🏃‍♂️ ...
     ✅ Kubernetes Crossplane provider config applied.

     Deploying the Klutch Crossplane configuration package...

    任╭──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮任
    任│                                                                                                                                                                                          │任
    任│                                                                     The following command will be executed for you:                                                                      │任
    任│                                                                                                                                                                                          │任
    任│  /usr/local/bin/kubectl apply -f https://raw.githubusercontent.com/anynines/klutchio/refs/tags/v1.3.0/crossplane-api/deploy/config-pkg-anynines.yaml --context arn:aws:eks:eu-central-   │任
    任│                                                                       1:378836732719:cluster/klutch-control-plane                                                                        │任
    任│                                                                                                                                                                                          │任
    任╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯任
    任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁任何尼宁
     ✅ Klutch Crossplane configuration package applied.

     Waiting for the Klutch Crossplane configuration package to become ready...

     ✅ The Klutch Crossplane configuration package appears to be ready.

     Press <ENTER> key to continue or <CTRL>+C to abort.
    ```

  </details>

- It would be nice to have estimates for how long certain tasks approximately
  take, e.g.

  ```bash
  ℹ️  No Cognito settings provided. Provisioning Cognito (region: eu-central-1, prefix: hub)...
  ℹ️  Estimated time: 20-25min
  ```

- Is Cognito ever going to be done provisioning? AWS console tells me it was
  created 10 minutes ago, what are we waiting on? More information would be nice
  - also, if I didn't have access to the aws console I wouldn't even know the
    provisioning request was accepted
  - if this information is hidden behind the verbose flag then a concise version
    of it should be displayed by default
  - 21 minutes later, AWS dashboard shows no changes - now the CLI succeeds
    - this took ~35 minutes, a warning would have definitely been appropriate
- it is odd how even without the verbose flag after creating the Cognito User
  Pool the CLI starts showing full manifests and commands in this giant frame
- maybe an alternative to printing the manifests all into the console
  unconditionally would be to have the following structure for approvals:

  ```bash
  ℹ️  Planned action: [concise description of what the CLI will attempt to do]`

  Press <SPACEBAR> [or any other key] key to show [manifests|command], <ENTER> key to continue or <CTRL>+C to abort.
  ```

  Then if the user chooses to inspect the manifests they are not just printed
  into their command line but shown in paginated viewer like `less`.
- Also, I'd suggest we dial back the formatting around the commands (frame,
  empty lines and centered alignment), as the current formatting makes it very
  hard to copy-paste commands for troubleshooting
- Time needed: 1h 50m (just for creating the cluster)
- after raising the max IP quota the `a9s create cluster klutch workload`
  command worked
  without a hitch, although the specific set of commands detailed in the
  tutorial needed to be adapted (more on that in the [Suggested Tweaks Section](#suggested-tweaks))
- `a9s create klutch pg instance` worked
  - I'd still suggest reducing the formatting of the commands
  - also, one of the commands that is printed is `/usr/local/bin/kubectl apply
    -f -` which is obviously missing some piece in order to be useful
    information
  - claim became ready immediately regardless of the state of the MR which was
    stuck in pending
  - MR was stuck in pending because setup for Control Plane Cluster does not install the add-on necessary for
    provisioning persistent storage

    => PG instance never becomes ready, yet the Claim was `ready`

    - fix

      ```bash
      eksctl create iamserviceaccount \
            --name ebs-csi-controller-sa \
            --namespace kube-system \
            --cluster klutch-control-plane \
            --role-name AmazonEKS_EBS_CSI_DriverRole \
            --role-only \
            --attach-policy-arn arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy \
            --approve
      eksctl create addon --cluster klutch-control-plane --name aws-ebs-csi-driver
      --version latest --force
      ```

- `a9s create klutch pg [servicebinding|backup|restore]` => same feedback as for `a9s create klutch pg instance`
- `a9s delete klutch pg servicebinding` deleted the claim, but the secret was
  left behind when the instance could not get ready due to the missing eks add-on
  - turns out this was due to the instance never getting ready, which means the
    other claims (service binding, backup and restore) should not have become
    ready either, arguably the CLI should not have applied them in the first
    place (or been able to
    apply them, if you want to shift the responsibility from the CLI to a Cloud
    Component)
  - further investigation confirmed that the claims **and** the compositions
    were immediately ready regardless of the state of the managed resources
- the "nestedness" of the CLI does not feel great to use
  - I understand that this is due to the desire to separate the "local flow" and
    the "remote flow", but it feels cumbersome to use when not copy-pasting
    commands outright
  - a max of two subcommands should be what we strive for in my opinion, either `a9s
    <verb/command> <subresource>` or `a9s <subresource> <verb/command>`, I prefer
    the latter option to be honest
    - e.g.

      ```bash
      a9s remote create-control-plane
      a9s local create-control-plane
      # or
      a9s create control-plane --target remote
      a9s create control-plane --target local
      a9s create pg-instance --target remote # maybe a default target can be put into a config file in the User's home directory
      a9s create pg-service-binding --target remote # I wonder if this even needs to be called in an instance-specific way or if it would be feasible for the CLI to auto-determine the service type based on the instance ref provided
      ```

- Backup MR could not be deleted on the Control Plane Cluster, yet Backup Claim
  got deleted
  - backup-manager logs say: `failed to extract stats for backup metadata file: Access Denied.`
  - manual inspection of the minio bucket showed a backup file was created, so
    the credentials for putting it there seemed to work - I don't know why the
    ones for deletion don't
  - the admin credentials also don't work
  - fix is in [open PR against a8s-backup-manager repo](https://github.com/anynines/a8s-backup-manager/pull/99)
  - interim image is at `378836732719.dkr.ecr.eu-central-1.amazonaws.com/a8s-backup-manager:fix-minio-deletion-0`

## Feedback on the tutorial

- it is not ideal to have the output in `CREATE_OUTPUT="$(a9s create cluster
  klutch workload -p aws --tenant-name "${TENANT}" --eks-nodes 1 --yes)"`
  hidden
- `eksctl`, `jq` and `rg` are all required but not part of the preflight checks
- `rg` ([ripgrep](https://github.com/burntsushi/ripgrep)) is required for the
  process but not mentioned in the requirements

### Suggested tweaks

- Change...

  - ```bash
    WORKLOAD_CLUSTER="$(echo "${CREATE_OUTPUT}" | sed -n 's/.*Generated workload cluster name: \(klutch-workload-cluster-[a-z0-9]\+\).*/\1/p' | head -n1)"
    ```

    With

    ```bash
    WORKLOAD_CLUSTER="$(echo "${CREATE_OUTPUT}" | grep -o 'klutch-workload-cluster-[a-z0-9]\+'|uniq)"
    ```

  - ```bash
    WORKLOAD_CLUSTER="$(echo "${CREATE_OUTPUT}" | sed -n 's/.*Generated workload cluster name: \(klutch-workload-cluster-[a-z0-9]\+\).*/\1/p' | head -n1)"
    ```

    to

    ```bash
    WORKLOAD_CLUSTER="$(echo "${CREATE_OUTPUT}" | grep -o 'klutch-workload-cluster-[a-z0-9]\+'|uniq)"
    ```

    Advantages:
    - the original `sed` expression is not very portable (didn't work on my machine)
    - BSD `sed` (which macOS uses) differs more from GNU `sed` (which most Linux
      distros use) than BSD `grep` from GNU `grep`, making `sed` less portable
      than `grep` in general
    - substituting `head -n1` for `uniq` allows testers to catch if there are
      multiple different names in the output (which would hint towards undesired
      behavior)

    Alternative suggestion (less preferable to me as I deem it to be the more
    hacky way to extract the name): instead of `grep` replace the `sed` call with

    ```bash
    WORKLOAD_CLUSTER="$(echo "${CREATE_OUTPUT}" | sed -E 's/^.*Generated workload cluster name: (klutch-workload-cluster-[a-z0-9]+).*$/\1/p' | uniq)" # or head -n1
    ```

  - ```bash
    kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" -o yaml
    ```

    to

    ```bash
    kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" -o jsonpath='{.status.managed}'
    ```

- ```bash
  kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.database}' | base64 -d; echo
  kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.instance_service}' | base64 -d; echo
  kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.username}' | base64 -d; echo
  kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.password}' | base64 -d; echo
  ```

  to

  ```bash
  echo "database name: $(kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.database}' | base64 -d)"
  echo "instance service: $(kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.instance_service}' | base64 -d)"
  echo "username: $(kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.username}' | base64 -d)"
  echo "password: $(kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.password}' | base64 -d)"
  ```

- ```bash
  kubectl get backups.anynines.com "${BU}" -n "${NS}" -o yaml
  ```

  to

  ```bash
  kubectl get backups.anynines.com "${BU}" -n "${NS}" -o json | jq '.status.managed'
  ```
