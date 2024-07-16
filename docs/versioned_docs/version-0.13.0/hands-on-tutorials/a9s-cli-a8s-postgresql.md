---
id: hands-on-tutorial-a8s-pg-a9s-cli
title: "Deploying a Demo App using a8s PostgreSQL"
tags:
  - a9s hub  
  - a9s cli
  - a8s data services
  - a8s postgres
  - data service
  - tutorial
  - kubernetes
  - minikube
  - kind
keywords:
  - a9s hub  
  - a9s cli
  - a8s data services
  - a8s postgres
  - data service
  - tutorial
  - kubernetes
  - minikube
  - kind
  - postgresql
  - web app
---

# Overview



## What you will accomplish

In this tutorial you will learn how to **create a local Kubernetes cluster**, fully equipped **with a PostgreSQL** operator, ready for you to deploy a PostgreSQL database instance for **developing your application**.

## What you will learn

* Install the [a9s CLI](https://github.com/anynines/a9s-cli-v2)
* Create a local Kubernetes cluster
* Install [cert-manager](https://cert-manager.io/docs/)
* Install a8s PostgreSQL
* Create a PostgreSQL database instance
* Create a PostgreSQL user
* Connect to the PostgreSQL database
* Deploy a demo application
* Connect the application to the PostgreSQL database
* Create a backup
* Restore a backup

## Prerequisites

* MacOS / Linux 
    * Other platforms, including Windows, may work but are currently untested.
* [Docker](https://www.docker.com/)
* [Minikube](https://minikube.sigs.k8s.io/docs/start/) or [Kind](https://kind.sigs.k8s.io/)
* [a9s CLI](https://github.com/anynines/a9s-cli-v2)
* [Kubectl](https://kubernetes.io/docs/reference/kubectl/)
* Optional for backup/restore: AWS S3 Bucket with credentials

# Implementation

In this tutorial you will be using the `a9s` CLI to facilitate the creation of both a local Kubernetes cluster and a PostgreSQL database instance. 

The `a9s` CLI will guide you through the process while providing you with transparency and ability to set your own pace. Transparency means that you will see the exact commands to be executed. By default, the commands are executed only after you have confirmed the execution by pressing the `<ENTER>` key. This allows you to have a closer look at the command and/or the YAML specifications to understand what the current step in the tutorial is about. If all you care about is the result, the `--yes` option will answer all yes-no questions with yes. See [[1]](https://github.com/anynines/a9s-cli-v2) for documentation and source code of the `a9s` CLI.

## Step 1: Creating a Kubernetes Cluster with a8s PostgreSQL

In this section you will create a Kubernetes cluster with a8s PostgreSQL and all its dependencies:

    a9s create cluster a8s

Per default, `minikube` will be used. In case you prefer `kind` you can use the `--provider` option:
    
    a9s create cluster a8s --provider kind

The remainder of the tutorial works equally for both `minikube` and `kind`.

### Step 1.1: Initial Configuration on the First a9s create cluster Execution

When creating a cluster for the first time, a few setup steps will have to be taken which need to be performed only once:

1. Setting up a working directory for the use with the `a9s` CLI. **This step asks for your confirmation of the proposed directory.**
2. Configuring the access credentials for the S3 compatible object store which is needed to use the backup/restore feature of a8s Postgres. This step is performed automatically.
3. Cloning deployment resources required by the `a9s` CLI to create a cluster. This step is performed automatically.

### What's Happening During the Installation

After the initial configuration, the Kubernetes cluster is being created.

#### Cert-Manager

Once the Kubernetes cluster is ready, the `a9s` CLI proceeds with the installation of the [cert-manager](https://cert-manager.io/docs/). The cert-manager is a Kubernetes extension handling TLS certificates. Among others, in a8s PostgreSQL TSL certificates are used for securing the communication between Kubernetes and the operator.

#### a8s PostgreSQL

With the cert-manager being ready, the `a9s` CLI continues and installs the a8s PostgreSQL components. Namely, this is 

* The PostgreSQL operator
* The Service Binding controller
* The Backup Manager

The **PostgreSQL Operator** is responsible for creating and managing *Service Instances*, that is dedicated PostgreSQL servers represented by a single or a cluster of Pods.

The **Service Binding Controller**, as the name suggests, is responsible for creating so-called *Service Bindings*. A Service Binding represents **a unique set of credentials** connecting a database client, such as an application and a Service Instance, in this case a PostgreSQL instance. In the case of a8s PostgreSQL, a Service Binding contains a **username/password** combination as well as other information necessary to establish a connection such as the **hostname**.

The **Backup Manager** is responsible for managing backup and restore requests and dispatching them to the *Backup Agents* located alongside Postgres Service Instances. It is the Backup Agent of a Service Instance that actually triggers the execution, encryption, compression and streaming of backup and restore operations.

After *waiting for a8s Postgres Control Plane to become ready* the message `🎉 The a8s Postgres Control Plane appears to be ready. All expected pods are running.` indicates that **the installation of a8s PostgreSQL was successful**.

## Step 2: Creating a PostgreSQL Cluster

In order to keep all tutorial resources in one place, create a Kubernetes `tutorial` namespace:

    kubectl create namespace tutorial

Now that the a8s PostgreSQL Operator and the `tutorial` namespace is ready, it's time to create a database.

Using the `a9s` CLI the process is as simple as:

    a9s create pg instance --name clustered-instance --replicas 3 -n tutorial

This creates a clustered PostgreSQL instance named `clustered-instance` represented as a StatefulSet with `3` Pods. Each Pod runs a PostgreSQL process.

**Note**: The `a9s CLI` does not shield you the YAML specs is generated. Quite the opposite, it is intended to provide you with meaningful templates to start with. **You can find all YAML specs generated by the `a9s CLI` in the `usermanifests` folder in your a9s working directory**:

    ls $(a9s cluster pwd)/usermanifests

### Inspecting the Service Instance

It's worth inspecting the PostgreSQL Service Instance to see what the a8s PostgreSQL Operator has created:

    kubectl get postgresqls -n tutorial

Output:

    NAME                 AGE
    clustered-instance   131m

The `postgresql` object named `clustered-instance`, as the name suggests, represents your PostgreSQL instance. It is implemented by a set of Kubernetes Services and a StatefulSet.

    kubectl get statefulsets -n tutorial

The operator has created a Kubernetes StatefulSet with the name `clustered-instance`:

    NAME                 READY   AGE
    clustered-instance   3/3     89m

And the StatefulSet, in turn, manages three Pods, namely:

    kubectl get pods -n tutorial

The following Pods:

    NAME                   READY   STATUS    RESTARTS   AGE
    clustered-instance-0   3/3     Running   0          70m
    clustered-instance-1   3/3     Running   0          68m
    clustered-instance-2   3/3     Running   0          66m

Have a closer look at one of them:

    kubectl describe pod clustered-instance-0 -n tutorial

Especially, look at the `Labels` section in the output:


    Name:             clustered-instance-0
    Namespace:        tutorial
    Priority:         0
    Service Account:  clustered-instance
    Node:             a8s-demo-m02/192.168.58.3
    Start Time:       Tue, 12 Mar 2024 08:15:39 +0100
    Labels:           a8s.a9s/dsi-group=postgresql.anynines.com
                    a8s.a9s/dsi-kind=Postgresql
                    a8s.a9s/dsi-name=clustered-instance
                    a8s.a9s/replication-role=master
                    apps.kubernetes.io/pod-index=0
                    controller-revision-hash=clustered-instance-749699f5b9
                    statefulset.kubernetes.io/pod-name=clustered-instance-0

The label `a8s.a9s/replication-role=master` indicates that the Pod `clustered-instance-0` is the **primary** PostgreSQL server for the asynchronous streaming replication within the cluster. Don't worry if you are not familiar with this terminology. Just bare in mind that **all data altering SQL statements always need to go to the primary Pod**. There's a mechanism in place that will help with this.

By executing:

    kubectl get services -n tutorial

You will see a `clustered-instance-master` Kubernetes service:

    NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
    clustered-instance-config   ClusterIP   None           \<none\>        \<none\>              74m
    clustered-instance-master   ClusterIP   10.105.7.211   \<none\>        5432/TCP,8008/TCP   75m

**The `clustered-instance-master` service provides a reference to the primary PostgreSQL server within the clustered Service Instance**. As the cluster comes with failure-detection and automatic failover capabilities, the primary role may be assigned to another Pod in the cluster during leading election. However, the `clustered-instance-master` service will be updated so that any application connecting through the `clustered-instance-master` service automatically connects to the **current** primary.

**Congratulations 🎉**, you've managed to create yourself a highly available PostgreSQL cluster using asynchronous streaming replication.

## Step 3: Creating a Service Binding

In order to prepare the deployment of an application, the database need to be configured to **grant the application access to the PostgreSQL service instance**. Granting an application running in Kubernetes access to a PostgreSQL database involves the following steps:

1. Create a unique set of access credentials including a database role as well as a corresponding password. 

2. Creating a Kubernetes Secret containing the credentials.


The credential set should be unique to the application and the data service instance. So if a second application, such as a worker process, needs access, a separate credential set and Kubernetes Secret is to be created.

With a8s PostgreSQL the process of creating access credentials on-demand is referred to as creating *Service Bindings*. In other words, **a Service Binding in a8s PostgreSQL is a database role, password which is then stored in a Kubernetes Secret** to be used by exactly one application.

Think about the implication of managing Service Bindings using the Kubernetes API. Instead of writing custom scripts connecting to the database, the creation of a database user is as simple as creating a Kubernetes object. Therefore, **Service Bindings facilitate deployments to multiple Kubernetes environments describing application systems entirely using Kubernetes objects**.

Creating a Service Binding is easy:

    a9s create pg servicebinding --name sb-sample -n tutorial -i clustered-instance

Have a look at the resources that have been generated:

    kubectl get servicebindings -n tutorial

Output:

    NAME        AGE
    sb-sample   6s

The `servicebinding` object named `sb-sample` is owned by the a8s PostgreSQL Operator or, more precisely, the ServiceBindingController. As part of the Service Binding, a Kubernetes Secret has been created:

    kubectl get secrets -n tutorial

Output:

    NAME                                      TYPE     DATA   AGE
    postgres.credentials.clustered-instance   Opaque   2      9m16s
    sb-sample-service-binding                 Opaque   4      25s
    standby.credentials.clustered-instance    Opaque   2      9m16s

Investigate the Secret `sb-sample-service-binding`:

    kubectl get secret sb-sample-service-binding -n tutorial -o yaml

Output:

```yaml
apiVersion: v1
data:
  database: YTlzX2FwcHNfZGVmYXVsdF9kYg==
  instance_service: Y2x1c3RlcmVkLWluc3RhbmNlLW1hc3Rlci50dXRvcmlhbA==
  password: bk1wNGI2WHdMeXUwYVkzWmF4ekExS1VURTNzM2xham4=
  username: YThzLXNiLWN4cDZCMFRUQg==
immutable: true
kind: Secret
metadata:
  creationTimestamp: "2024-03-12T14:50:33Z"
  finalizers:
  - a8s.anynines.com/servicebinding.controller
  labels:
    service-binding: "true"
  name: sb-sample-service-binding
  namespace: tutorial
  ownerReferences:
  - apiVersion: servicebindings.anynines.com/v1beta3
    blockOwnerDeletion: true
    controller: true
    kind: ServiceBinding
    name: sb-sample
    uid: e4636254-433a-4e82-a46b-e79fd7f25f58
  resourceVersion: "2648"
  uid: ebee4e29-4796-4e9a-8114-ec4d546644a9
type: Opaque
```

Note that the values in the `data` hash aren't readable right away as they are base64 encoded. Values can be decoded using the `base64` command, for example:

`database:`

    echo "YTlzX2FwcHNfZGVmYXVsdF9kYg==" | base64 --decode
    a9s_apps_default_db

`instance_service:`

    echo "Y2x1c3RlcmVkLWluc3RhbmNlLW1hc3Rlci50dXRvcmlhbA==" | base64 --decode
    clustered-instance-master.tutorial


Given a Service name, the generic naming pattern in Kubernetes to derive its DNS entry is: `{service-name}.{namespace}.svc.{cluster-domain:cluster.local}`.

Assuming that your Kubernetes' cluster domain is the default `cluster.local`, this means that the primary (formerly master) node of your PostgreSQL cluster is reachable via the DNS entry: **`clustered-instance-master.tutorial.svc.cluster.local`**. 

`username:`

    echo "YThzLXNiLWN4cDZCMFRUQg==" | base64 --decode
    a8s-sb-cxp6B0TTB

`password:`

    echo "bk1wNGI2WHdMeXUwYVkzWmF4ekExS1VURTNzM2xham4=" | base64 --decode
    nMp4b6XwLyu0aY3ZaxzA1KUTE3s3lajn

As you can see, the secret `sb-sample-service-binding` contains all relevant information required by an application to connect to your PostgreSQL instance.

## Step 4: Deploying a Demo Application

With the PostgreSQL database at hand, an exemplary application can be deployed. 

The demo app has already been checked out for you. Hence, installing it just a single command away:

    kubectl apply -k $(a9s cluster pwd)/a8s-demo/demo-app -n tutorial

Output:

    service/demo-app created
    deployment.apps/demo-app created

The demo app consists of a Kubernetes Service and a Deployment both named `demo-app`.

You can verify that the app is running by executing:

    kubectl get pods -n tutorial -l app=demo-app

Output:

    NAME                        READY   STATUS    RESTARTS   AGE
    demo-app-65f6dd4445-glgc4   1/1     Running   0          81s

In order to access the app locally, create a port forward mapping the container port `3000` your local machine's port `8080`:

    kubectl port-forward service/demo-app -n tutorial 8080:3000

Then navigate your browser to: [http://localhost:8080](http://localhost:8080)

## Step 5: Interacting with PostgreSQL

Once you've created a PostgreSQL Service Instance, you can use the `a9s CLI` to interact with it.

### Applying a Local SQL File

Although not the preferred way to load seed data into a production database, during development it might be handy to execute a SQL file to a PostgreSQL instance. This allows executing one or multiple SQL statements conveniently. 

Download an exemplary SQL file:

    wget https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/demo_data.sql

Download an exemplary SQL file:

    wget https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/demo_data.sql

Executing an SQL file is as simple as using the `--file` option:

    a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial

The `a9s CLI` will determine the replication leader, upload, execute and delete the SQL file. 

The `--no-delete` option can be used during debugging of erroneous SQL statements
as the SQL file remains in the PostgreSQL Leader's Pod.

    a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial --no-delete

With the SQL file still available in the Pod, statements can be quickly altered and re-tested.

### Applying an SQL String

It is also possible to execute a SQL string containing one or several SQL statements by using the `--sql` option:

    a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"

The output of the command will be printed on the screen, for example:

    Output from the Pod:
        
    count 
    -------
        10 
    (1 row)

Again, the `pg apply` commands are not meant to interact with production databases but may become handy during debugging and local development.

Be aware that these commands are executed by the privileged `postgres` user. Schemas (tables) created by the `postgres` user may not be accessible by roles (users) created in conjunction with Service Bindings. You will then have to grant access privileges to the Service Binding role.

## Step 6: Creating and Restoring a Backup

Assuming you have configured the backup store and provided access credentials to an AWS S3 compatible object store, try creating and restoring a backup for your application.

### Creating a Backup

Creating a backup can be achieved with a single command:

    a9s create pg backup --name clustered-backup-1 -i clustered-instance -n tutorial

With a closer look at the output you will notice that a backup is also specified by a YAML specification and thus is done in a declarative way. You express that you want a backup to be created:

```YAML
apiVersion: backups.anynines.com/v1beta3
kind: Backup
metadata:
    name: clustered-backup-1
    namespace: tutorial
spec:
    serviceInstance:
    apiGroup: postgresql.anynines.com
    kind: Postgresql
    name: clustered-instance
```

The a8s Backup Manager is the responsible for making the backup happen. It does that by locating the Service Instance `clustered-instance` which also runs the `a8s Backup Agent`. This agent is then executing the PostgreSQL backup command and, depending on its configuration, compressing, encrypting and streaming the backup to the backup object store (S3).

### Restoring a Backup

In order to experience the value of a backup, simulate a data loss by issueing the following `DELETE` statement:

    a9s pg apply -i clustered-instance -n tutorial --sql "DELETE FROM posts"

Verify the destructive effect on your data by counting the number of posts:

    a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"

And/or reloading the demo-app.

Once you've confirmed that all blog posts are gone, it's time to recover the data from the backup.

    a9s create pg restore --name clustered-restore-1 -b clustered-backup-1 -i clustered-instance -n tutorial

Again, apply the `COUNT` or reload the website to see that the restore has brought back all blog posts.

    a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"

Some engineers say that a convenient backup/restore functionality at your disposal improves the quality of sleep by 37% 😉.

## Congratulations

With just a few commands, you have created a local Kubernetes cluster, installed the a8s PostgreSQL Control Plane including all its dependencies. Furthermore, you have provisioned an PostgreSQL cluster consisting of three Pods providing you with an asynchronous streaming cluster supporting automatic failure detection, lead-election and failover. Deploying the demo application you've also experienced the convenience of Service Bindings and their automatic creation of Kubernetes Secrets. The backup and restore experiment then illustrated how effortless handling a production database can be.

Did you every think that running a production database as an application developer with full self-service could be so easy?

## What to do next?

Wait, there's more to it! This hands-on tutorial merely scratched the surface. Did you see that the `a9s CLI` has created many YAML manifests stored in the `usermanifests` folder of your working directory? This is a good place to start tweaking your manifests and start your own experiments.

If you want to learn more about a8s PostgreSQL feel free to have a look at the documentation at TODO.

For more about the `a9s CLI` have a look at https://github.com/anynines/a9s-cli-v2.

## Links

1. a9s CLI documentation and source, https://github.com/anynines/a9s-cli-v2 
2. PostgreSQL documentation, Log-Shipping Standby Servers, https://www.postgresql.org/docs/current/warm-standby.html