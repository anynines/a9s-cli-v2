# Klutch Demo

This package adds functionality for the `klutch` CLI command. It allows demoing the creation of a management cluster with `a8s-framework` and a consumer cluster, binding to the management cluster from the consumer cluster, as well as deleting the clusters.

There are three subcommands:
- `deploy`
- `bind`
- `delete`

The following assumptions for using `klutch` commands are made:
- `kind`, `kubectl`, `helm`, `git` are present in the PATH.
- The following external resources are reachable:
    - Files on https://raw.githubusercontent.com/
    - https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/v1.3.0/crds.yaml
    - `public.ecr.aws/w5n9a2g2/anynines/` image repositories
    - `dexidp/dex` image
    - `curlimages/curl` image
    - https://raw.githubusercontent.com/anynines/klutchio/main/crossplane-api/deploy/config-pkg-anynines.yaml
    - https://charts.crossplane.io

## `deploy` command

`a9s klutch deploy --port 8080 --yes`

This command uses `kind` to deploy the management cluster with all required components including `a8s-framework` and a consumer cluster.

It has the following command line flags:
- `yes` : skips "Wait" prompts, this is inherited from the root command.
- `port`: the cluster's ingress will listen on this port. It defaults to `8080`.

This command writes a files to the user's configured workspace which contains the IP and port of the management cluster.
This allows subsequent commands such as `bind` to correctly connect to the management cluster.

## `bind` command

`a9s klutch bind`

It makes following assumptions:
- `kubectl` and the `kubectl-bind` plugin (`v1.3.0`) are present in the PATH.

This command automates the interactive binding process where possible. The `kubectl bind` command is called, opening a browser tab/window where the authorization can be performed. Once this is completed, the automation resumes and finishes the binding process.

It has the following command line flags:
- `yes` : skips "Wait" prompts, this is inherited from the root command.

## `delete` command

`a9s klutch delete`

This command deletes the management and consumer kind clusters.
