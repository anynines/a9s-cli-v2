"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[9948],{7312:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>l,contentTitle:()=>c,default:()=>h,frontMatter:()=>r,metadata:()=>t,toc:()=>o});const t=JSON.parse('{"id":"hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli","title":"Deploying a Demo App using a8s PostgreSQL","description":"What you will accomplish","source":"@site/versioned_docs/version-0.14.1/hands-on-tutorials/a9s-cli-a8s-postgresql.md","sourceDirName":"hands-on-tutorials","slug":"/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli","permalink":"/docs/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli","draft":false,"unlisted":false,"tags":[{"inline":true,"label":"a9s hub","permalink":"/docs/tags/a-9-s-hub"},{"inline":true,"label":"a9s cli","permalink":"/docs/tags/a-9-s-cli"},{"inline":true,"label":"a8s data services","permalink":"/docs/tags/a-8-s-data-services"},{"inline":true,"label":"a8s postgres","permalink":"/docs/tags/a-8-s-postgres"},{"inline":true,"label":"data service","permalink":"/docs/tags/data-service"},{"inline":true,"label":"tutorial","permalink":"/docs/tags/tutorial"},{"inline":true,"label":"kubernetes","permalink":"/docs/tags/kubernetes"},{"inline":true,"label":"minikube","permalink":"/docs/tags/minikube"},{"inline":true,"label":"kind","permalink":"/docs/tags/kind"}],"version":"0.14.1","frontMatter":{"id":"hands-on-tutorial-a8s-pg-a9s-cli","title":"Deploying a Demo App using a8s PostgreSQL","tags":["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind"],"keywords":["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind","postgresql","web app"]}}');var a=s(4848),i=s(8453);const r={id:"hands-on-tutorial-a8s-pg-a9s-cli",title:"Deploying a Demo App using a8s PostgreSQL",tags:["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind"],keywords:["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind","postgresql","web app"]},c="Overview",l={},o=[{value:"What you will accomplish",id:"what-you-will-accomplish",level:2},{value:"What you will learn",id:"what-you-will-learn",level:2},{value:"Prerequisites",id:"prerequisites",level:2},{value:"Step 1: Creating a Kubernetes Cluster with a8s PostgreSQL",id:"step-1-creating-a-kubernetes-cluster-with-a8s-postgresql",level:2},{value:"Step 1.1: Initial Configuration on the First a9s create cluster Execution",id:"step-11-initial-configuration-on-the-first-a9s-create-cluster-execution",level:3},{value:"What&#39;s Happening During the Installation",id:"whats-happening-during-the-installation",level:3},{value:"Cert-Manager",id:"cert-manager",level:4},{value:"a8s PostgreSQL",id:"a8s-postgresql",level:4},{value:"Step 2: Creating a PostgreSQL Cluster",id:"step-2-creating-a-postgresql-cluster",level:2},{value:"Inspecting the Service Instance",id:"inspecting-the-service-instance",level:3},{value:"Step 3: Creating a Service Binding",id:"step-3-creating-a-service-binding",level:2},{value:"Step 4: Deploying a Demo Application",id:"step-4-deploying-a-demo-application",level:2},{value:"Step 5: Interacting with PostgreSQL",id:"step-5-interacting-with-postgresql",level:2},{value:"Applying a Local SQL File",id:"applying-a-local-sql-file",level:3},{value:"Applying an SQL String",id:"applying-an-sql-string",level:3},{value:"Step 6: Creating and Restoring a Backup",id:"step-6-creating-and-restoring-a-backup",level:2},{value:"Creating a Backup",id:"creating-a-backup",level:3},{value:"Restoring a Backup",id:"restoring-a-backup",level:3},{value:"Congratulations",id:"congratulations",level:2},{value:"What to do next?",id:"what-to-do-next",level:2},{value:"Links",id:"links",level:2}];function d(e){const n={a:"a",code:"code",em:"em",h1:"h1",h2:"h2",h3:"h3",h4:"h4",header:"header",li:"li",ol:"ol",p:"p",pre:"pre",strong:"strong",ul:"ul",...(0,i.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsx)(n.header,{children:(0,a.jsx)(n.h1,{id:"overview",children:"Overview"})}),"\n",(0,a.jsx)(n.h2,{id:"what-you-will-accomplish",children:"What you will accomplish"}),"\n",(0,a.jsxs)(n.p,{children:["In this tutorial you will learn how to ",(0,a.jsx)(n.strong,{children:"create a local Kubernetes cluster"}),", fully equipped ",(0,a.jsx)(n.strong,{children:"with a PostgreSQL"})," operator, ready for you to deploy a PostgreSQL database instance for ",(0,a.jsx)(n.strong,{children:"developing your application"}),"."]}),"\n",(0,a.jsx)(n.h2,{id:"what-you-will-learn",children:"What you will learn"}),"\n",(0,a.jsxs)(n.ul,{children:["\n",(0,a.jsxs)(n.li,{children:["Install the ",(0,a.jsx)(n.a,{href:"https://github.com/anynines/a9s-cli-v2",children:"a9s CLI"})]}),"\n",(0,a.jsx)(n.li,{children:"Create a local Kubernetes cluster"}),"\n",(0,a.jsxs)(n.li,{children:["Install ",(0,a.jsx)(n.a,{href:"https://cert-manager.io/docs/",children:"cert-manager"})]}),"\n",(0,a.jsx)(n.li,{children:"Install a8s PostgreSQL"}),"\n",(0,a.jsx)(n.li,{children:"Create a PostgreSQL database instance"}),"\n",(0,a.jsx)(n.li,{children:"Create a PostgreSQL user"}),"\n",(0,a.jsx)(n.li,{children:"Connect to the PostgreSQL database"}),"\n",(0,a.jsx)(n.li,{children:"Deploy a demo application"}),"\n",(0,a.jsx)(n.li,{children:"Connect the application to the PostgreSQL database"}),"\n",(0,a.jsx)(n.li,{children:"Create a backup"}),"\n",(0,a.jsx)(n.li,{children:"Restore a backup"}),"\n"]}),"\n",(0,a.jsx)(n.h2,{id:"prerequisites",children:"Prerequisites"}),"\n",(0,a.jsxs)(n.ul,{children:["\n",(0,a.jsxs)(n.li,{children:["MacOS / Linux","\n",(0,a.jsxs)(n.ul,{children:["\n",(0,a.jsx)(n.li,{children:"Other platforms, including Windows, may work but are currently untested."}),"\n"]}),"\n"]}),"\n",(0,a.jsx)(n.li,{children:(0,a.jsx)(n.a,{href:"https://www.docker.com/",children:"Docker"})}),"\n",(0,a.jsxs)(n.li,{children:[(0,a.jsx)(n.a,{href:"https://minikube.sigs.k8s.io/docs/start/",children:"Minikube"})," or ",(0,a.jsx)(n.a,{href:"https://kind.sigs.k8s.io/",children:"Kind"})]}),"\n",(0,a.jsx)(n.li,{children:(0,a.jsx)(n.a,{href:"https://github.com/anynines/a9s-cli-v2",children:"a9s CLI"})}),"\n",(0,a.jsx)(n.li,{children:(0,a.jsx)(n.a,{href:"https://kubernetes.io/docs/reference/kubectl/",children:"Kubectl"})}),"\n",(0,a.jsx)(n.li,{children:"Optional for backup/restore: AWS S3 Bucket with credentials"}),"\n"]}),"\n",(0,a.jsx)(n.h1,{id:"implementation",children:"Implementation"}),"\n",(0,a.jsxs)(n.p,{children:["In this tutorial you will be using the ",(0,a.jsx)(n.code,{children:"a9s"})," CLI to facilitate the creation of both a local Kubernetes cluster and a PostgreSQL database instance."]}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.code,{children:"a9s"})," CLI will guide you through the process while providing you with transparency and ability to set your own pace. Transparency means that you will see the exact commands to be executed. By default, the commands are executed only after you have confirmed the execution by pressing the ",(0,a.jsx)(n.code,{children:"ENTER"})," key. This allows you to have a closer look at the command and/or the YAML specifications to understand what the current step in the tutorial is about. If all you care about is the result, the ",(0,a.jsx)(n.code,{children:"--yes"})," option will answer all yes-no questions with yes. See ",(0,a.jsx)(n.a,{href:"https://github.com/anynines/a9s-cli-v2",children:"[1]"})," for documentation and source code of the ",(0,a.jsx)(n.code,{children:"a9s"})," CLI."]}),"\n",(0,a.jsx)(n.h2,{id:"step-1-creating-a-kubernetes-cluster-with-a8s-postgresql",children:"Step 1: Creating a Kubernetes Cluster with a8s PostgreSQL"}),"\n",(0,a.jsx)(n.p,{children:"In this section you will create a Kubernetes cluster with a8s PostgreSQL and all its dependencies:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s create cluster a8s\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Per default, ",(0,a.jsx)(n.code,{children:"minikube"})," will be used. In case you prefer ",(0,a.jsx)(n.code,{children:"kind"})," you can use the ",(0,a.jsx)(n.code,{children:"--provider"})," option:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s create cluster a8s --provider kind\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The remainder of the tutorial works equally for both ",(0,a.jsx)(n.code,{children:"minikube"})," and ",(0,a.jsx)(n.code,{children:"kind"}),"."]}),"\n",(0,a.jsx)(n.h3,{id:"step-11-initial-configuration-on-the-first-a9s-create-cluster-execution",children:"Step 1.1: Initial Configuration on the First a9s create cluster Execution"}),"\n",(0,a.jsx)(n.p,{children:"When creating a cluster for the first time, a few setup steps will have to be taken which need to be performed only once:"}),"\n",(0,a.jsxs)(n.ol,{children:["\n",(0,a.jsxs)(n.li,{children:["Setting up a working directory for the use with the ",(0,a.jsx)(n.code,{children:"a9s"})," CLI. ",(0,a.jsx)(n.strong,{children:"This step asks for your confirmation of the proposed directory."})]}),"\n",(0,a.jsx)(n.li,{children:"Configuring the access credentials for the S3 compatible object store which is needed to use the backup/restore feature of a8s Postgres. This step is performed automatically."}),"\n",(0,a.jsxs)(n.li,{children:["Cloning deployment resources required by the ",(0,a.jsx)(n.code,{children:"a9s"})," CLI to create a cluster. This step is performed automatically."]}),"\n"]}),"\n",(0,a.jsx)(n.h3,{id:"whats-happening-during-the-installation",children:"What's Happening During the Installation"}),"\n",(0,a.jsx)(n.p,{children:"After the initial configuration, the Kubernetes cluster is being created."}),"\n",(0,a.jsx)(n.h4,{id:"cert-manager",children:"Cert-Manager"}),"\n",(0,a.jsxs)(n.p,{children:["Once the Kubernetes cluster is ready, the ",(0,a.jsx)(n.code,{children:"a9s"})," CLI proceeds with the installation of the ",(0,a.jsx)(n.a,{href:"https://cert-manager.io/docs/",children:"cert-manager"}),". The cert-manager is a Kubernetes extension handling TLS certificates. Among others, in a8s PostgreSQL TSL certificates are used for securing the communication between Kubernetes and the operator."]}),"\n",(0,a.jsx)(n.h4,{id:"a8s-postgresql",children:"a8s PostgreSQL"}),"\n",(0,a.jsxs)(n.p,{children:["With the cert-manager being ready, the ",(0,a.jsx)(n.code,{children:"a9s"})," CLI continues and installs the a8s PostgreSQL components. Namely, this is"]}),"\n",(0,a.jsxs)(n.ul,{children:["\n",(0,a.jsx)(n.li,{children:"The PostgreSQL operator"}),"\n",(0,a.jsx)(n.li,{children:"The Service Binding controller"}),"\n",(0,a.jsx)(n.li,{children:"The Backup Manager"}),"\n"]}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.strong,{children:"PostgreSQL Operator"})," is responsible for creating and managing ",(0,a.jsx)(n.em,{children:"Service Instances"}),", that is dedicated PostgreSQL servers represented by a single or a cluster of Pods."]}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.strong,{children:"Service Binding Controller"}),", as the name suggests, is responsible for creating so-called ",(0,a.jsx)(n.em,{children:"Service Bindings"}),". A Service Binding represents ",(0,a.jsx)(n.strong,{children:"a unique set of credentials"})," connecting a database client, such as an application and a Service Instance, in this case a PostgreSQL instance. In the case of a8s PostgreSQL, a Service Binding contains a ",(0,a.jsx)(n.strong,{children:"username/password"})," combination as well as other information necessary to establish a connection such as the ",(0,a.jsx)(n.strong,{children:"hostname"}),"."]}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.strong,{children:"Backup Manager"})," is responsible for managing backup and restore requests and dispatching them to the ",(0,a.jsx)(n.em,{children:"Backup Agents"})," located alongside Postgres Service Instances. It is the Backup Agent of a Service Instance that actually triggers the execution, encryption, compression and streaming of backup and restore operations."]}),"\n",(0,a.jsxs)(n.p,{children:["After ",(0,a.jsx)(n.em,{children:"waiting for a8s Postgres Control Plane to become ready"})," the message ",(0,a.jsx)(n.code,{children:"\ud83c\udf89 The a8s Postgres Control Plane appears to be ready. All expected pods are running."})," indicates that ",(0,a.jsx)(n.strong,{children:"the installation of a8s PostgreSQL was successful"}),"."]}),"\n",(0,a.jsx)(n.h2,{id:"step-2-creating-a-postgresql-cluster",children:"Step 2: Creating a PostgreSQL Cluster"}),"\n",(0,a.jsxs)(n.p,{children:["In order to keep all tutorial resources in one place, create a Kubernetes ",(0,a.jsx)(n.code,{children:"tutorial"})," namespace:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl create namespace tutorial\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Now that the a8s PostgreSQL Operator and the ",(0,a.jsx)(n.code,{children:"tutorial"})," namespace is ready, it's time to create a database."]}),"\n",(0,a.jsxs)(n.p,{children:["Using the ",(0,a.jsx)(n.code,{children:"a9s"})," CLI the process is as simple as:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s create pg instance --name clustered-instance --replicas 3 -n tutorial\n"})}),"\n",(0,a.jsxs)(n.p,{children:["This creates a clustered PostgreSQL instance named ",(0,a.jsx)(n.code,{children:"clustered-instance"})," represented as a StatefulSet with ",(0,a.jsx)(n.code,{children:"3"})," Pods. Each Pod runs a PostgreSQL process."]}),"\n",(0,a.jsxs)(n.p,{children:[(0,a.jsx)(n.strong,{children:"Note"}),": The ",(0,a.jsx)(n.code,{children:"a9s CLI"})," does not shield you the YAML specs is generated. Quite the opposite, it is intended to provide you with meaningful templates to start with. ",(0,a.jsxs)(n.strong,{children:["You can find all YAML specs generated by the ",(0,a.jsx)(n.code,{children:"a9s CLI"})," in the ",(0,a.jsx)(n.code,{children:"usermanifests"})," folder in your a9s working directory"]}),":"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"ls $(a9s cluster pwd)/usermanifests\n"})}),"\n",(0,a.jsx)(n.h3,{id:"inspecting-the-service-instance",children:"Inspecting the Service Instance"}),"\n",(0,a.jsx)(n.p,{children:"It's worth inspecting the PostgreSQL Service Instance to see what the a8s PostgreSQL Operator has created:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get postgresqls -n tutorial\n"})}),"\n",(0,a.jsx)(n.p,{children:"Output:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"NAME                 AGE\nclustered-instance   131m\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.code,{children:"postgresql"})," object named ",(0,a.jsx)(n.code,{children:"clustered-instance"}),", as the name suggests, represents your PostgreSQL instance. It is implemented by a set of Kubernetes Services and a StatefulSet."]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get statefulsets -n tutorial\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The operator has created a Kubernetes StatefulSet with the name ",(0,a.jsx)(n.code,{children:"clustered-instance"}),":"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"NAME                 READY   AGE\nclustered-instance   3/3     89m\n"})}),"\n",(0,a.jsx)(n.p,{children:"And the StatefulSet, in turn, manages three Pods, namely:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get pods -n tutorial\n"})}),"\n",(0,a.jsx)(n.p,{children:"The following Pods:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"NAME                   READY   STATUS    RESTARTS   AGE\nclustered-instance-0   3/3     Running   0          70m\nclustered-instance-1   3/3     Running   0          68m\nclustered-instance-2   3/3     Running   0          66m\n"})}),"\n",(0,a.jsx)(n.p,{children:"Have a closer look at one of them:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl describe pod clustered-instance-0 -n tutorial\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Especially, look at the ",(0,a.jsx)(n.code,{children:"Labels"})," section in the output:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"Name:             clustered-instance-0\nNamespace:        tutorial\nPriority:         0\nService Account:  clustered-instance\nNode:             a8s-demo-m02/192.168.58.3\nStart Time:       Tue, 12 Mar 2024 08:15:39 +0100\nLabels:           a8s.a9s/dsi-group=postgresql.anynines.com\n                a8s.a9s/dsi-kind=Postgresql\n                a8s.a9s/dsi-name=clustered-instance\n                a8s.a9s/replication-role=master\n                apps.kubernetes.io/pod-index=0\n                controller-revision-hash=clustered-instance-749699f5b9\n                statefulset.kubernetes.io/pod-name=clustered-instance-0\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The label ",(0,a.jsx)(n.code,{children:"a8s.a9s/replication-role=master"})," indicates that the Pod ",(0,a.jsx)(n.code,{children:"clustered-instance-0"})," is the ",(0,a.jsx)(n.strong,{children:"primary"})," PostgreSQL server for the asynchronous streaming replication within the cluster. Don't worry if you are not familiar with this terminology. Just bare in mind that ",(0,a.jsx)(n.strong,{children:"all data altering SQL statements always need to go to the primary Pod"}),". There's a mechanism in place that will help with this."]}),"\n",(0,a.jsx)(n.p,{children:"By executing:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get services -n tutorial\n"})}),"\n",(0,a.jsxs)(n.p,{children:["You will see a ",(0,a.jsx)(n.code,{children:"clustered-instance-master"})," Kubernetes service:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE\nclustered-instance-config   ClusterIP   None           <none>        <none>              74m\nclustered-instance-master   ClusterIP   10.105.7.211   <none>        5432/TCP,8008/TCP   75m\n"})}),"\n",(0,a.jsxs)(n.p,{children:[(0,a.jsxs)(n.strong,{children:["The ",(0,a.jsx)(n.code,{children:"clustered-instance-master"})," service provides a reference to the primary PostgreSQL server within the clustered Service Instance"]}),". As the cluster comes with failure-detection and automatic failover capabilities, the primary role may be assigned to another Pod in the cluster during leading election. However, the ",(0,a.jsx)(n.code,{children:"clustered-instance-master"})," service will be updated so that any application connecting through the ",(0,a.jsx)(n.code,{children:"clustered-instance-master"})," service automatically connects to the ",(0,a.jsx)(n.strong,{children:"current"})," primary."]}),"\n",(0,a.jsxs)(n.p,{children:[(0,a.jsx)(n.strong,{children:"Congratulations \ud83c\udf89"}),", you've managed to create yourself a highly available PostgreSQL cluster using asynchronous streaming replication."]}),"\n",(0,a.jsx)(n.h2,{id:"step-3-creating-a-service-binding",children:"Step 3: Creating a Service Binding"}),"\n",(0,a.jsxs)(n.p,{children:["In order to prepare the deployment of an application, the database need to be configured to ",(0,a.jsx)(n.strong,{children:"grant the application access to the PostgreSQL service instance"}),". Granting an application running in Kubernetes access to a PostgreSQL database involves the following steps:"]}),"\n",(0,a.jsxs)(n.ol,{children:["\n",(0,a.jsxs)(n.li,{children:["\n",(0,a.jsx)(n.p,{children:"Create a unique set of access credentials including a database role as well as a corresponding password."}),"\n"]}),"\n",(0,a.jsxs)(n.li,{children:["\n",(0,a.jsx)(n.p,{children:"Creating a Kubernetes Secret containing the credentials."}),"\n"]}),"\n"]}),"\n",(0,a.jsx)(n.p,{children:"The credential set should be unique to the application and the data service instance. So if a second application, such as a worker process, needs access, a separate credential set and Kubernetes Secret is to be created."}),"\n",(0,a.jsxs)(n.p,{children:["With a8s PostgreSQL the process of creating access credentials on-demand is referred to as creating ",(0,a.jsx)(n.em,{children:"Service Bindings"}),". In other words, ",(0,a.jsx)(n.strong,{children:"a Service Binding in a8s PostgreSQL is a database role, password which is then stored in a Kubernetes Secret"})," to be used by exactly one application."]}),"\n",(0,a.jsxs)(n.p,{children:["Think about the implication of managing Service Bindings using the Kubernetes API. Instead of writing custom scripts connecting to the database, the creation of a database user is as simple as creating a Kubernetes object. Therefore, ",(0,a.jsx)(n.strong,{children:"Service Bindings facilitate deployments to multiple Kubernetes environments describing application systems entirely using Kubernetes objects"}),"."]}),"\n",(0,a.jsx)(n.p,{children:"Creating a Service Binding is easy:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s create pg servicebinding --name sb-sample -n tutorial -i clustered-instance\n"})}),"\n",(0,a.jsx)(n.p,{children:"Have a look at the resources that have been generated:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get servicebindings -n tutorial\n"})}),"\n",(0,a.jsx)(n.p,{children:"Output:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"NAME        AGE\nsb-sample   6s\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.code,{children:"servicebinding"})," object named ",(0,a.jsx)(n.code,{children:"sb-sample"})," is owned by the a8s PostgreSQL Operator or, more precisely, the ServiceBindingController. As part of the Service Binding, a Kubernetes Secret has been created:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get secrets -n tutorial\n"})}),"\n",(0,a.jsx)(n.p,{children:"Output:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"NAME                                      TYPE     DATA   AGE\npostgres.credentials.clustered-instance   Opaque   2      9m16s\nsb-sample-service-binding                 Opaque   4      25s\nstandby.credentials.clustered-instance    Opaque   2      9m16s\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Investigate the Secret ",(0,a.jsx)(n.code,{children:"sb-sample-service-binding"}),":"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get secret sb-sample-service-binding -n tutorial -o yaml\n"})}),"\n",(0,a.jsx)(n.p,{children:"Output:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-yaml",children:'apiVersion: v1\ndata:\n  database: YTlzX2FwcHNfZGVmYXVsdF9kYg==\n  instance_service: Y2x1c3RlcmVkLWluc3RhbmNlLW1hc3Rlci50dXRvcmlhbA==\n  password: bk1wNGI2WHdMeXUwYVkzWmF4ekExS1VURTNzM2xham4=\n  username: YThzLXNiLWN4cDZCMFRUQg==\nimmutable: true\nkind: Secret\nmetadata:\n  creationTimestamp: "2024-03-12T14:50:33Z"\n  finalizers:\n  - a8s.anynines.com/servicebinding.controller\n  labels:\n    service-binding: "true"\n  name: sb-sample-service-binding\n  namespace: tutorial\n  ownerReferences:\n  - apiVersion: servicebindings.anynines.com/v1beta3\n    blockOwnerDeletion: true\n    controller: true\n    kind: ServiceBinding\n    name: sb-sample\n    uid: e4636254-433a-4e82-a46b-e79fd7f25f58\n  resourceVersion: "2648"\n  uid: ebee4e29-4796-4e9a-8114-ec4d546644a9\ntype: Opaque\n'})}),"\n",(0,a.jsxs)(n.p,{children:["Note that the values in the ",(0,a.jsx)(n.code,{children:"data"})," hash aren't readable right away as they are base64 encoded. Values can be decoded using the ",(0,a.jsx)(n.code,{children:"base64"})," command, for example:"]}),"\n",(0,a.jsx)(n.p,{children:(0,a.jsx)(n.code,{children:"database:"})}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'echo "YTlzX2FwcHNfZGVmYXVsdF9kYg==" | base64 --decode\na9s_apps_default_db\n'})}),"\n",(0,a.jsx)(n.p,{children:(0,a.jsx)(n.code,{children:"instance_service:"})}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'echo "Y2x1c3RlcmVkLWluc3RhbmNlLW1hc3Rlci50dXRvcmlhbA==" | base64 --decode\nclustered-instance-master.tutorial\n'})}),"\n",(0,a.jsxs)(n.p,{children:["Given a Service name, the generic naming pattern in Kubernetes to derive its DNS entry is: ",(0,a.jsx)(n.code,{children:"{service-name}.{namespace}.svc.{cluster-domain:cluster.local}"}),"."]}),"\n",(0,a.jsxs)(n.p,{children:["Assuming that your Kubernetes' cluster domain is the default ",(0,a.jsx)(n.code,{children:"cluster.local"}),", this means that the primary (formerly master) node of your PostgreSQL cluster is reachable via the DNS entry: ",(0,a.jsx)(n.strong,{children:(0,a.jsx)(n.code,{children:"clustered-instance-master.tutorial.svc.cluster.local"})}),"."]}),"\n",(0,a.jsx)(n.p,{children:(0,a.jsx)(n.code,{children:"username:"})}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'echo "YThzLXNiLWN4cDZCMFRUQg==" | base64 --decode\na8s-sb-cxp6B0TTB\n'})}),"\n",(0,a.jsx)(n.p,{children:(0,a.jsx)(n.code,{children:"password:"})}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'echo "bk1wNGI2WHdMeXUwYVkzWmF4ekExS1VURTNzM2xham4=" | base64 --decode\nnMp4b6XwLyu0aY3ZaxzA1KUTE3s3lajn\n'})}),"\n",(0,a.jsxs)(n.p,{children:["As you can see, the secret ",(0,a.jsx)(n.code,{children:"sb-sample-service-binding"})," contains all relevant information required by an application to connect to your PostgreSQL instance."]}),"\n",(0,a.jsx)(n.h2,{id:"step-4-deploying-a-demo-application",children:"Step 4: Deploying a Demo Application"}),"\n",(0,a.jsx)(n.p,{children:"With the PostgreSQL database at hand, an exemplary application can be deployed."}),"\n",(0,a.jsx)(n.p,{children:"The demo app has already been checked out for you. Hence, installing it just a single command away:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl apply -k $(a9s cluster pwd)/a8s-demo/demo-postgresql-app -n tutorial\n"})}),"\n",(0,a.jsx)(n.p,{children:"Output:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"service/demo-app created\ndeployment.apps/demo-app created\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The demo app consists of a Kubernetes Service and a Deployment both named ",(0,a.jsx)(n.code,{children:"demo-app"}),"."]}),"\n",(0,a.jsx)(n.p,{children:"You can verify that the app is running by executing:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl get pods -n tutorial -l app=demo-app\n"})}),"\n",(0,a.jsx)(n.p,{children:"Output:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"NAME                        READY   STATUS    RESTARTS   AGE\ndemo-app-65f6dd4445-glgc4   1/1     Running   0          81s\n"})}),"\n",(0,a.jsxs)(n.p,{children:["In order to access the app locally, create a port forward mapping the container port ",(0,a.jsx)(n.code,{children:"3000"})," your local machine's port ",(0,a.jsx)(n.code,{children:"8080"}),":"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"kubectl port-forward service/demo-app -n tutorial 8080:3000\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Then navigate your browser to: ",(0,a.jsx)(n.a,{href:"http://localhost:8080",children:"http://localhost:8080"})]}),"\n",(0,a.jsx)(n.h2,{id:"step-5-interacting-with-postgresql",children:"Step 5: Interacting with PostgreSQL"}),"\n",(0,a.jsxs)(n.p,{children:["Once you've created a PostgreSQL Service Instance, you can use the ",(0,a.jsx)(n.code,{children:"a9s CLI"})," to interact with it."]}),"\n",(0,a.jsx)(n.h3,{id:"applying-a-local-sql-file",children:"Applying a Local SQL File"}),"\n",(0,a.jsx)(n.p,{children:"Although not the preferred way to load seed data into a production database, during development it might be handy to execute a SQL file to a PostgreSQL instance. This allows executing one or multiple SQL statements conveniently."}),"\n",(0,a.jsx)(n.p,{children:"Download an exemplary SQL file:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"curl https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/demo_data.sql -o demo_data.sql\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Executing an SQL file is as simple as using the ",(0,a.jsx)(n.code,{children:"--file"})," option:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.code,{children:"a9s CLI"})," will determine the replication leader, upload, execute and delete the SQL file."]}),"\n",(0,a.jsxs)(n.p,{children:["The ",(0,a.jsx)(n.code,{children:"--no-delete"})," option can be used during debugging of erroneous SQL statements\nas the SQL file remains in the PostgreSQL Leader's Pod."]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial --no-delete\n"})}),"\n",(0,a.jsx)(n.p,{children:"With the SQL file still available in the Pod, statements can be quickly altered and re-tested."}),"\n",(0,a.jsx)(n.h3,{id:"applying-an-sql-string",children:"Applying an SQL String"}),"\n",(0,a.jsxs)(n.p,{children:["It is also possible to execute a SQL string containing one or several SQL statements by using the ",(0,a.jsx)(n.code,{children:"--sql"})," option:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"\n'})}),"\n",(0,a.jsx)(n.p,{children:"The output of the command will be printed on the screen, for example:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{children:"Output from the Pod:\n    \ncount \n-------\n    10 \n(1 row)\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Again, the ",(0,a.jsx)(n.code,{children:"pg apply"})," commands are not meant to interact with production databases but may become handy during debugging and local development."]}),"\n",(0,a.jsxs)(n.p,{children:["Be aware that these commands are executed by the privileged ",(0,a.jsx)(n.code,{children:"postgres"})," user. Schemas (tables) created by the ",(0,a.jsx)(n.code,{children:"postgres"})," user may not be accessible by roles (users) created in conjunction with Service Bindings. You will then have to grant access privileges to the Service Binding role."]}),"\n",(0,a.jsx)(n.h2,{id:"step-6-creating-and-restoring-a-backup",children:"Step 6: Creating and Restoring a Backup"}),"\n",(0,a.jsx)(n.p,{children:"Assuming you have configured the backup store and provided access credentials to an AWS S3 compatible object store, try creating and restoring a backup for your application."}),"\n",(0,a.jsx)(n.h3,{id:"creating-a-backup",children:"Creating a Backup"}),"\n",(0,a.jsx)(n.p,{children:"Creating a backup can be achieved with a single command:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s create pg backup --name clustered-backup-1 -i clustered-instance -n tutorial\n"})}),"\n",(0,a.jsx)(n.p,{children:"With a closer look at the output you will notice that a backup is also specified by a YAML specification and thus is done in a declarative way. You express that you want a backup to be created:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-YAML",children:"apiVersion: backups.anynines.com/v1beta3\nkind: Backup\nmetadata:\n    name: clustered-backup-1\n    namespace: tutorial\nspec:\n    serviceInstance:\n    apiGroup: postgresql.anynines.com\n    kind: Postgresql\n    name: clustered-instance\n"})}),"\n",(0,a.jsxs)(n.p,{children:["The a8s Backup Manager is the responsible for making the backup happen. It does that by locating the Service Instance ",(0,a.jsx)(n.code,{children:"clustered-instance"})," which also runs the ",(0,a.jsx)(n.code,{children:"a8s Backup Agent"}),". This agent is then executing the PostgreSQL backup command and, depending on its configuration, compressing, encrypting and streaming the backup to the backup object store (S3)."]}),"\n",(0,a.jsx)(n.h3,{id:"restoring-a-backup",children:"Restoring a Backup"}),"\n",(0,a.jsxs)(n.p,{children:["In order to experience the value of a backup, simulate a data loss by issueing the following ",(0,a.jsx)(n.code,{children:"DELETE"})," statement:"]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'a9s pg apply -i clustered-instance -n tutorial --sql "DELETE FROM posts"\n'})}),"\n",(0,a.jsx)(n.p,{children:"Verify the destructive effect on your data by counting the number of posts:"}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"\n'})}),"\n",(0,a.jsx)(n.p,{children:"And/or reloading the demo-app."}),"\n",(0,a.jsx)(n.p,{children:"Once you've confirmed that all blog posts are gone, it's time to recover the data from the backup."}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:"a9s create pg restore --name clustered-restore-1 -b clustered-backup-1 -i clustered-instance -n tutorial\n"})}),"\n",(0,a.jsxs)(n.p,{children:["Again, apply the ",(0,a.jsx)(n.code,{children:"COUNT"})," or reload the website to see that the restore has brought back all blog posts."]}),"\n",(0,a.jsx)(n.pre,{children:(0,a.jsx)(n.code,{className:"language-bash",children:'a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"\n'})}),"\n",(0,a.jsx)(n.p,{children:"Some engineers say that a convenient backup/restore functionality at your disposal improves the quality of sleep by 37% \ud83d\ude09."}),"\n",(0,a.jsx)(n.h2,{id:"congratulations",children:"Congratulations"}),"\n",(0,a.jsx)(n.p,{children:"With just a few commands, you have created a local Kubernetes cluster, installed the a8s PostgreSQL Control Plane including all its dependencies. Furthermore, you have provisioned an PostgreSQL cluster consisting of three Pods providing you with an asynchronous streaming cluster supporting automatic failure detection, lead-election and failover. Deploying the demo application you've also experienced the convenience of Service Bindings and their automatic creation of Kubernetes Secrets. The backup and restore experiment then illustrated how effortless handling a production database can be."}),"\n",(0,a.jsx)(n.p,{children:"Did you every think that running a production database as an application developer with full self-service could be so easy?"}),"\n",(0,a.jsx)(n.h2,{id:"what-to-do-next",children:"What to do next?"}),"\n",(0,a.jsxs)(n.p,{children:["Wait, there's more to it! This hands-on tutorial merely scratched the surface. Did you see that the ",(0,a.jsx)(n.code,{children:"a9s CLI"})," has created many YAML manifests stored in the ",(0,a.jsx)(n.code,{children:"usermanifests"})," folder of your working directory? This is a good place to start tweaking your manifests and start your own experiments."]}),"\n",(0,a.jsx)(n.p,{children:"If you want to learn more about a8s PostgreSQL feel free to have a look at the documentation at TODO."}),"\n",(0,a.jsxs)(n.p,{children:["For more about the ",(0,a.jsx)(n.code,{children:"a9s CLI"})," have a look at ",(0,a.jsx)(n.a,{href:"https://github.com/anynines/a9s-cli-v2",children:"https://github.com/anynines/a9s-cli-v2"}),"."]}),"\n",(0,a.jsx)(n.h2,{id:"links",children:"Links"}),"\n",(0,a.jsxs)(n.ol,{children:["\n",(0,a.jsxs)(n.li,{children:["a9s CLI documentation and source, ",(0,a.jsx)(n.a,{href:"https://github.com/anynines/a9s-cli-v2",children:"https://github.com/anynines/a9s-cli-v2"})]}),"\n",(0,a.jsxs)(n.li,{children:["PostgreSQL documentation, Log-Shipping Standby Servers, ",(0,a.jsx)(n.a,{href:"https://www.postgresql.org/docs/current/warm-standby.html",children:"https://www.postgresql.org/docs/current/warm-standby.html"})]}),"\n"]})]})}function h(e={}){const{wrapper:n}={...(0,i.R)(),...e.components};return n?(0,a.jsx)(n,{...e,children:(0,a.jsx)(d,{...e})}):d(e)}},8453:(e,n,s)=>{s.d(n,{R:()=>r,x:()=>c});var t=s(6540);const a={},i=t.createContext(a);function r(e){const n=t.useContext(i);return t.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function c(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(a):e.components||a:r(e.components),t.createElement(i.Provider,{value:n},e.children)}}}]);