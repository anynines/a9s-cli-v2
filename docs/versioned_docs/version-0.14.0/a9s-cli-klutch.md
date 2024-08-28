---
id: a9s-cli-klutch
title: a9s CLI Klutch
tags:
  - a9s cli
  - a9s hub  
  - a9s data services
  - a8s data services
  - a9s postgres
  - a8s postgres
  - data service
  - introduction
  - kubernetes
  - minikube
  - kind
  - klutch
keywords:
  - a9s cli
  - a9s hub
  - a9s platform
  - a9s data services
  - a8s data services
  - a9s postgres
  - a8s postgres
  - data service
  - introduction
  - postgresql  
  - kubernetes
  - minikube
  - kind
  - klutch
---

# klutch Stack

Create a local Klutch central management cluster using `Kind`, including the `a8s` stack. Deploy a consumer cluster and **bind** resources to the management cluster.
This will allow you to use `a8s` resource instances such as `postgresql` on the consumer cluster, which will run on the management cluster.

## Commands 

### 1. `deploy`

**Usage**:
```bash
a9s klutch deploy [options]
```

**Options**:
|Flag|Description|Example|
|----|-----------|-------|
|`-y`, `--yes`| Skip confirmation prompts | `a9s klutch deploy --yes` |
|`--port`| The port to expose the management cluster on. Defaults to `8080`. | `a9s klutch deploy --port 8080` | 

**Description**:

This command deploys a `Kind` cluster named `klutch-management` and installs the required
components for Klutch. These components include:
- The `klutch-bind` backend and [Dex Idp](https://dexidp.io/) as a dummy OICD provider.
- Crossplane and the anynines configuration packages.
- The complete `a8s` stack including `Postgresql` operator, backup, restore and service binding capabilities.

In addition to the management cluster, a **consumer** cluster named `klutch-consumer` is deployed. This cluster can be used for the `a9s klutch bind` command to bind resources to the management cluster.

The management cluster exports the following resources for binding:

- `postgresqlinstance.anynines.com`
- `servicebinding.anynines.com`
- `backup.anynines.com`
- `restore.anynines.com`

**Important**: For technical reasons, the management cluster is exposed on the local network using the local IP address. If your IP or network changes, the management cluster may become unreachable and will have to be redeployed.

### 2. `bind`

**Usage**:
```
a9s klutch bind [options]
```

**Options**:
|Flag|Description|Example|
|----|-----------|-------|
|`-y`, `--yes`| Skip confirmation prompts | `a9s klutch bind --yes` |

**Prerequisites**:
- `kubectl`
- `kube-bind` plugin (see below)

#### Installing the `kubectl-bind` plugin:

Download a binary for your platform with the following URL, make it executable and place it in a location in your `PATH`:

`https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/v1.3.0/$OS-$ARCH/kubectl-bind`

Replace `OS` and `ARCH` with values for your platform, e.g. `darwin-arm64` or `linux-amd64`. You can also use the following script to achieve this:

```bash
RELEASE="v1.3.0"
OS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o kubectl-bind https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/$RELEASE/$OS-$ARCH/kubectl-bind

sudo chmod 755 kubectl-bind
sudo mv kubectl-bind /usr/local/bin
```

**Description**:

This command will invoke `kubectl bind` in order to bind a resource exported by the management cluster. This process will open a browser window for you where you can authenticate with the dummy dex OIDC provider using these credentials:

Email: `admin@example.com`

Password: `password`

After logging in, grant access, and then **choose the resource you would like to bind**. Once this is done, return to your terminal and wait for the process to finish.

After the `bind` command has succeeded, you can deploy instances of the chosen resource on your consumer cluster, which will run in the management cluster. The command will print an example manifest for the resource you bound that you can apply to the consumer cluster with `kubectl`.

### 3. `delete`

**Usage**:

```bash
a9s klutch delete [options]
```

**Options**:

|Flag|Description|Example|
|----|-----------|-------|
|`-y`, `--yes`| Skip confirmation prompts | `a9s klutch delete --yes` |

**Description**:

This command deletes the management and consumer clusters.
