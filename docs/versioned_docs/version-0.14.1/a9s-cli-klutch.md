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

Create a local Klutch Control Plane Cluster using `Kind`, including the `a8s` stack. Deploy an App Cluster and **bind** resources to the Control Plane Cluster.
This will allow you to use `a8s` resource instances such as `postgresql` on the App Cluster, which will run on the Control Plane Cluster.

## Prerequisites
- [General prerequisites](./a9s-cli-index.md#prerequisites) are met.
- Install [Helm](https://helm.sh/docs/intro/install/).
- Install `kubectl-bind` plugin version 1.3.0 or higher (see below).
- On **linux**, docker must be runnable without sudo. See the [docker documentation](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user) for further details.

### Installing the `kubectl-bind` plugin:

Download a binary for your platform with the following URL, make it executable and place it in a location in your `PATH`:

`https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/v1.3.0/$OS-$ARCH/kubectl-bind`

Replace `OS` and `ARCH` with values for your platform, e.g. `darwin-arm64` or `linux-amd64`. You can also use the following script to achieve this:

```bash
RELEASE="v1.3.0"
OS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o kubectl-bind https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/$RELEASE/$OS-$ARCH/kubectl-bind

sudo chmod 755 kubectl-bind
sudo mv kubectl-bind /usr/local/bin
```

### Running on Linux

To avoid issues with `Kind` on Linux, increase the `inotify` resource limits as described [here](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

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
|`--port`| The port to expose the Control Plane Cluster on. Defaults to `8080`. | `a9s klutch deploy --port 8080` | 

**Description**:

This command deploys a `Kind` cluster named `klutch-control-plane` and installs the required
components for Klutch. These components include:
- The `klutch-bind` backend and [Dex Idp](https://dexidp.io/) as a dummy OICD provider.
- Crossplane and the anynines configuration packages.
- The complete `a8s` stack including `Postgresql` operator, backup, restore and service binding capabilities.

In addition to the Control Plane Cluster, an App Cluster named `klutch-app` is deployed. This cluster can be used for the `a9s klutch bind` command to bind resources to the Control Plane Cluster.

The Control Plane Cluster exports the following resources for binding:

- `postgresqlinstance.anynines.com`
- `servicebinding.anynines.com`
- `backup.anynines.com`
- `restore.anynines.com`

**Important**: For technical reasons, the Control Plane Cluster is exposed on the local network using the local IP address. If your IP or network changes, the Control Plane Cluster may become unreachable and will have to be redeployed.

### 2. `bind`

**Usage**:
```
a9s klutch bind [options]
```

**Options**:
|Flag|Description|Example|
|----|-----------|-------|
|`-y`, `--yes`| Skip confirmation prompts | `a9s klutch bind --yes` |

**Description**:

This command will invoke `kubectl bind` in order to bind a resource exported by the Control Plane Cluster. This process will open a browser window for you where you can authenticate with the dummy dex OIDC provider using these credentials:

Email: `admin@example.com`

Password: `password`

After logging in, grant access, and then **choose the resource you would like to bind**. Once this is done, return to your terminal and wait for the process to finish.

After the `bind` command has succeeded, you can deploy instances of the chosen resource on your App Cluster, which will run in the Control Plane Cluster. The command will print an example manifest for the resource you bound that you can apply to the App Cluster with `kubectl`. You can do this easily by copying the printed yaml and using a heredoc, like so:

```bash
kubectl apply -f - <<EOF
<paste your manifests>
EOF
```

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

This command deletes the Control Plane and App clusters.
