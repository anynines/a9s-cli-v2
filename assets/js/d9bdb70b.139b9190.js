"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[3347],{9220:(e,s,n)=>{n.r(s),n.d(s,{assets:()=>l,contentTitle:()=>c,default:()=>h,frontMatter:()=>t,metadata:()=>i,toc:()=>o});const i=JSON.parse('{"id":"a9s-cli","title":"a9s CLI","description":"anynines provides a command line tool called a9s to facilitate application development, devops tasks and interact with selected anynines products.","source":"@site/versioned_docs/version-0.13.0/a9s-cli.md","sourceDirName":".","slug":"/a9s-cli","permalink":"/docs/0.13.0/a9s-cli","draft":false,"unlisted":false,"tags":[{"inline":true,"label":"a9s cli","permalink":"/docs/0.13.0/tags/a-9-s-cli"},{"inline":true,"label":"a9s hub","permalink":"/docs/0.13.0/tags/a-9-s-hub"},{"inline":true,"label":"a9s data services","permalink":"/docs/0.13.0/tags/a-9-s-data-services"},{"inline":true,"label":"a8s data services","permalink":"/docs/0.13.0/tags/a-8-s-data-services"},{"inline":true,"label":"a9s postgres","permalink":"/docs/0.13.0/tags/a-9-s-postgres"},{"inline":true,"label":"a8s postgres","permalink":"/docs/0.13.0/tags/a-8-s-postgres"},{"inline":true,"label":"data service","permalink":"/docs/0.13.0/tags/data-service"},{"inline":true,"label":"introduction","permalink":"/docs/0.13.0/tags/introduction"},{"inline":true,"label":"kubernetes","permalink":"/docs/0.13.0/tags/kubernetes"},{"inline":true,"label":"minikube","permalink":"/docs/0.13.0/tags/minikube"},{"inline":true,"label":"kind","permalink":"/docs/0.13.0/tags/kind"}],"version":"0.13.0","frontMatter":{"id":"a9s-cli","title":"a9s CLI","tags":["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind"],"keywords":["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind"]}}');var r=n(4848),a=n(8453);const t={id:"a9s-cli",title:"a9s CLI",tags:["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind"],keywords:["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind"]},c="a9s CLI",l={},o=[{value:"Use Cases",id:"use-cases",level:2},{value:"<code>a8s</code> Stack",id:"a8s-stack",level:3},{value:"Cold-Run",id:"cold-run",level:2},{value:"Setting Up a Working Directory",id:"setting-up-a-working-directory",level:3},{value:"Configuring the Backup Store",id:"configuring-the-backup-store",level:3},{value:"Skip Checking Prerequisites",id:"skip-checking-prerequisites",level:2},{value:"Number of Kubernetes Nodes",id:"number-of-kubernetes-nodes",level:2},{value:"Cluster Memory",id:"cluster-memory",level:2},{value:"Deployment Version",id:"deployment-version",level:2},{value:"Kubernetes Provider",id:"kubernetes-provider",level:2},{value:"Backup Infrastructure Region",id:"backup-infrastructure-region",level:2},{value:"Unattended Mode",id:"unattended-mode",level:2},{value:"Printing the Working Directory",id:"printing-the-working-directory",level:2},{value:"Creating a PostgreSQL Service Instance",id:"creating-a-postgresql-service-instance",level:2},{value:"Creating PostgreSQL Service Instance YAML Without Applying it",id:"creating-postgresql-service-instance-yaml-without-applying-it",level:3},{value:"Creating a Custom PostgreSQL Service Instance",id:"creating-a-custom-postgresql-service-instance",level:3},{value:"Deleting a PostgreSQL Service Instance",id:"deleting-a-postgresql-service-instance",level:2},{value:"Applying a SQL File to a PostgreSQL Service Instance",id:"applying-a-sql-file-to-a-postgresql-service-instance",level:2},{value:"Applying a SQL Statement to a PostgreSQL Service Instance",id:"applying-a-sql-statement-to-a-postgresql-service-instance",level:2},{value:"Creating a Backup of a PostgreSQL Service Instance",id:"creating-a-backup-of-a-postgresql-service-instance",level:2},{value:"Restoring a Backup of PostgreSQL Service Instance",id:"restoring-a-backup-of-postgresql-service-instance",level:2},{value:"Creating a PostgreSQL Service Binding",id:"creating-a-postgresql-service-binding",level:2}];function d(e){const s={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",header:"header",li:"li",ol:"ol",p:"p",pre:"pre",strong:"strong",ul:"ul",...(0,a.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(s.header,{children:(0,r.jsx)(s.h1,{id:"a9s-cli",children:"a9s CLI"})}),"\n",(0,r.jsxs)(s.p,{children:["anynines provides a command line tool called ",(0,r.jsx)(s.code,{children:"a9s"})," to facilitate application development, devops tasks and interact with selected anynines products."]}),"\n",(0,r.jsx)(s.h2,{id:"use-cases",children:"Use Cases"}),"\n",(0,r.jsxs)(s.p,{children:["The ",(0,r.jsx)(s.code,{children:"a9s"})," CLI can be used to install and use the following stacks:"]}),"\n",(0,r.jsxs)(s.h3,{id:"a8s-stack",children:[(0,r.jsx)(s.code,{children:"a8s"})," Stack"]}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsxs)(s.li,{children:["Install a local Kubernetes cluster (",(0,r.jsx)(s.code,{children:"minikube"})," or ",(0,r.jsx)(s.code,{children:"kind"}),")."]}),"\n",(0,r.jsxs)(s.li,{children:["Install the ",(0,r.jsx)(s.a,{href:"https://cert-manager.io/",children:"cert-manager"}),"."]}),"\n",(0,r.jsx)(s.li,{children:"Install a local Minio object store for storing Backups."}),"\n",(0,r.jsxs)(s.li,{children:["Install the a8s PostgreSQL Operator PostgreSQL supporting","\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsxs)(s.li,{children:["creating dedicated PostgreSQL clusters with","\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:"synchronous and asynchronous streaming replication."}),"\n",(0,r.jsx)(s.li,{children:"automatic failure detection and automatic failover."}),"\n"]}),"\n"]}),"\n",(0,r.jsx)(s.li,{children:"backup and restore capabilities storing backups in an S3 compatible object store such as AWS S3 or Minio."}),"\n",(0,r.jsx)(s.li,{children:"ability to easily create database users and Kubernetes Secrets by using the Service Bindings abstraction"}),"\n"]}),"\n"]}),"\n",(0,r.jsxs)(s.li,{children:["Easily apply ",(0,r.jsx)(s.code,{children:".sql"})," files and SQL commands to PostgreSQL clusters."]}),"\n"]}),"\n",(0,r.jsx)(s.h1,{id:"prerequisites",children:"Prerequisites"}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:"Using the backup/restore feature of a8s PostgreSQL requires an S3 compatible endpoint."}),"\n",(0,r.jsxs)(s.li,{children:["Install Go (if you want ",(0,r.jsx)(s.code,{children:"go env"})," to identify your OS and arch)."]}),"\n",(0,r.jsx)(s.li,{children:"Install Git."}),"\n",(0,r.jsx)(s.li,{children:"Install Docker."}),"\n",(0,r.jsx)(s.li,{children:"Install Kubectl."}),"\n",(0,r.jsx)(s.li,{children:"Install Kind and/or Minikube."}),"\n",(0,r.jsxs)(s.li,{children:["Install the ",(0,r.jsx)(s.a,{href:"https://cert-manager.io/docs/reference/cmctl/",children:"cert-manager CLI"}),"."]}),"\n"]}),"\n",(0,r.jsx)(s.h1,{id:"installing-the-cli",children:"Installing the CLI"}),"\n",(0,r.jsxs)(s.p,{children:["In order to install the ",(0,r.jsx)(s.code,{children:"a9s"})," CLI execute the following shell script:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"RELEASE=$(curl -L -s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/stable.txt); OS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o a9s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/releases/$RELEASE/a9s-$OS-$ARCH\n    \nsudo chmod 755 a9s\nsudo mv a9s /usr/local/bin\n"})}),"\n",(0,r.jsxs)(s.p,{children:["This will download the ",(0,r.jsx)(s.code,{children:"a9s"})," binary suitable for your architecture and move it to ",(0,r.jsx)(s.code,{children:"/usr/local/bin"}),".\nDepending on your system you have to adjust the ",(0,r.jsx)(s.code,{children:"PATH"})," variable or move the binary to a folder that's already in the ",(0,r.jsx)(s.code,{children:"PATH"}),"."]}),"\n",(0,r.jsx)(s.h1,{id:"using-the-cli",children:"Using the CLI"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s\n"})}),"\n",(0,r.jsx)(s.h1,{id:"creating-a-local-a8s-postgres-cluster",children:"Creating a Local a8s Postgres Cluster"}),"\n",(0,r.jsxs)(s.p,{children:["Create a local Kubernetes cluster using ",(0,r.jsx)(s.code,{children:"Minikube"})," or ",(0,r.jsx)(s.code,{children:"Kind"}),", ",(0,r.jsx)(s.strong,{children:"install a8s PostgreSQL"})," including its dependencies as well as a local ",(0,r.jsx)(s.a,{href:"https://min.io/",children:"Minio"})," object store."]}),"\n",(0,r.jsxs)(s.p,{children:["Get ready for ",(0,r.jsx)(s.strong,{children:"local development of applications with PostgreSQL"})," and/or ",(0,r.jsx)(s.strong,{children:"experimentation with a8s Postgres"})," by issuing the command:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s\n"})}),"\n",(0,r.jsx)(s.p,{children:"Recommended is 12 GB of free memory for the creation of three cluster nodes with each 4 GB. The number of nodes and memory size can be adjusted."}),"\n",(0,r.jsx)(s.h2,{id:"cold-run",children:"Cold-Run"}),"\n",(0,r.jsx)(s.p,{children:"When creating a cluster for the first time, a few setup steps will have to be taken which need to be performed only once:"}),"\n",(0,r.jsxs)(s.ol,{children:["\n",(0,r.jsxs)(s.li,{children:["Setting up a working directory for the use with the ",(0,r.jsx)(s.code,{children:"a9s"})," CLI. ",(0,r.jsx)(s.strong,{children:"This step asks for your confirmation of the proposed directory."})]}),"\n",(0,r.jsx)(s.li,{children:"Configuring the access credentials for the S3 compatible object store which is needed to use the backup/restore feature of a8s Postgres. This step is performed automatically."}),"\n",(0,r.jsxs)(s.li,{children:["Cloning deployment resources required by the ",(0,r.jsx)(s.code,{children:"a9s"})," CLI to create a cluster. This step is performed automatically."]}),"\n"]}),"\n",(0,r.jsx)(s.h3,{id:"setting-up-a-working-directory",children:"Setting Up a Working Directory"}),"\n",(0,r.jsxs)(s.p,{children:["The working directory is where are ",(0,r.jsx)(s.code,{children:"a9s"})," CLI related resources will go. This includes ",(0,r.jsx)(s.code,{children:"yaml"})," specifications being cloned from remote repositories, but also those generated by the ",(0,r.jsx)(s.code,{children:"a9s"})," CLI for your convenience."]}),"\n",(0,r.jsxs)(s.p,{children:["Once established, the working directory is stored in the ",(0,r.jsx)(s.code,{children:"~/.a9s"})," configuration file."]}),"\n",(0,r.jsxs)(s.p,{children:["The default working directory is ",(0,r.jsx)(s.code,{children:"~/a9s"}),"."]}),"\n",(0,r.jsx)(s.p,{children:"Alternatively, provide a custom working directory at the corresponding prompt."}),"\n",(0,r.jsx)(s.h3,{id:"configuring-the-backup-store",children:"Configuring the Backup Store"}),"\n",(0,r.jsx)(s.p,{children:"A non-prod Minio object store is installed in your local Kubernetes cluster and is automatically configured as the default backup store for a8s PostgreSQL backups."}),"\n",(0,r.jsxs)(s.p,{children:["If you want to use an alternative backup store, see ",(0,r.jsx)(s.code,{children:"a9s create cluster a8s --help"})," for the defaults of your particular CLI version and list of configuration options."]}),"\n",(0,r.jsx)(s.p,{children:"Most S3 compatible object stores, including AWS S3 itself of course, should work."}),"\n",(0,r.jsx)(s.h2,{id:"skip-checking-prerequisites",children:"Skip Checking Prerequisites"}),"\n",(0,r.jsx)(s.p,{children:"It is possible to skip the verification of prerequisites. This includes skipping the search for: required shell commands, a running Docker daemon and a running Kubernetes cluster."}),"\n",(0,r.jsxs)(s.p,{children:["In order to skip precheck use the ",(0,r.jsx)(s.code,{children:"--no-precheck"})," option:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s --no-precheck\n"})}),"\n",(0,r.jsx)(s.h2,{id:"number-of-kubernetes-nodes",children:"Number of Kubernetes Nodes"}),"\n",(0,r.jsx)(s.p,{children:"Specifying the number of Nodes in the Kubernetes cluster:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s --cluster-nr-of-nodes 1\n"})}),"\n",(0,r.jsx)(s.h2,{id:"cluster-memory",children:"Cluster Memory"}),"\n",(0,r.jsxs)(s.p,{children:["Specifying the memory of ",(0,r.jsx)(s.strong,{children:"each"})," Node of the Kubernetes cluster:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s --cluster-memory 4gb\n"})}),"\n",(0,r.jsx)(s.h2,{id:"deployment-version",children:"Deployment Version"}),"\n",(0,r.jsxs)(s.p,{children:["The deployment version refers to the version of manifests used for installing software. Deployment versions are managed by anynines in a Git repository. The deployment version option allows you to select a particular version of the deployment manifests identified by ",(0,r.jsx)(s.strong,{children:"Git tags"}),"."]}),"\n",(0,r.jsxs)(s.p,{children:["Select a particular release by providing the ",(0,r.jsx)(s.code,{children:"--deployment-version"})," parameter:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s --deployment-version v1.2.0\n"})}),"\n",(0,r.jsx)(s.p,{children:"Use:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s --deployment-version latest\n"})}),"\n",(0,r.jsx)(s.p,{children:"To get the latest, untagged version of the deployment manifests."}),"\n",(0,r.jsx)(s.h2,{id:"kubernetes-provider",children:"Kubernetes Provider"}),"\n",(0,r.jsxs)(s.p,{children:["When creating a Kubernetes cluster, the mechanism to manage the cluster can be selected by specifying the ",(0,r.jsx)(s.code,{children:"--provider"})," option:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s -p kind \na9s create cluster a8s -p minikube (default)\n"})}),"\n",(0,r.jsx)(s.p,{children:"Follow the instructions to learn about available sub commands."}),"\n",(0,r.jsx)(s.h2,{id:"backup-infrastructure-region",children:"Backup Infrastructure Region"}),"\n",(0,r.jsxs)(s.p,{children:["When using the backup and restore functionality, a backup infrastructure region must be specified by using the ",(0,r.jsx)(s.code,{children:"--backup-region"})," option:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s --backup-region us-east-1\n"})}),"\n",(0,r.jsxs)(s.p,{children:[(0,r.jsx)(s.strong,{children:"Note"}),": By default, an existing ",(0,r.jsx)(s.code,{children:"backup-config.yaml"})," will be used. Hence, if you intend to change\nyour backup config, remove the existing ",(0,r.jsx)(s.code,{children:"backup-config.yaml"}),", first:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"rm a8s-deployment/deploy/a8s/backup-config/backup-store-config.yaml\n"})}),"\n",(0,r.jsx)(s.h2,{id:"unattended-mode",children:"Unattended Mode"}),"\n",(0,r.jsxs)(s.p,{children:["It is possible to skip all yes-no questions by ",(0,r.jsx)(s.strong,{children:"enabling the unattended mode"})," by passing the ",(0,r.jsx)(s.code,{children:"-y"})," or ",(0,r.jsx)(s.code,{children:"--yes"})," flag:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create cluster a8s --yes\n"})}),"\n",(0,r.jsx)(s.h2,{id:"printing-the-working-directory",children:"Printing the Working Directory"}),"\n",(0,r.jsxs)(s.p,{children:["The working directory is stored in the ",(0,r.jsx)(s.code,{children:"~/.a8s"})," configuration file. The working directory contains all resources downloaded and generated by the ",(0,r.jsx)(s.code,{children:"a9s"})," CLI."]}),"\n",(0,r.jsx)(s.p,{children:"To print the working directory execute:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s cluster pwd\n"})}),"\n",(0,r.jsx)(s.h1,{id:"a8s-postgresql",children:"a8s PostgreSQL"}),"\n",(0,r.jsxs)(s.p,{children:["A selected subset of the a8s PostgreSQL features are available through the ",(0,r.jsx)(s.code,{children:"a9s"})," CLI."]}),"\n",(0,r.jsx)(s.h2,{id:"creating-a-postgresql-service-instance",children:"Creating a PostgreSQL Service Instance"}),"\n",(0,r.jsxs)(s.p,{children:["Creating a service instance with the name ",(0,r.jsx)(s.code,{children:"sample-pg-cluster"}),":"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create pg instance --name sample-pg-cluster\n"})}),"\n",(0,r.jsxs)(s.p,{children:["The generated YAML specification will be stored in the ",(0,r.jsx)(s.code,{children:"usermanifests"}),"."]}),"\n",(0,r.jsx)(s.h3,{id:"creating-postgresql-service-instance-yaml-without-applying-it",children:"Creating PostgreSQL Service Instance YAML Without Applying it"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create pg instance --name sample-pg-cluster --no-apply\n"})}),"\n",(0,r.jsxs)(s.p,{children:["The generated YAML specification will be stored in the ",(0,r.jsx)(s.code,{children:"usermanifests"})," but ",(0,r.jsx)(s.code,{children:"kubectl apply"})," won't be executed."]}),"\n",(0,r.jsx)(s.h3,{id:"creating-a-custom-postgresql-service-instance",children:"Creating a Custom PostgreSQL Service Instance"}),"\n",(0,r.jsx)(s.p,{children:"The command:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create pg instance --api-version v1beta3 --name sample-pg-cluster --namespace default --replicas 3 --requests-cpu 200m --limits-memory 200Mi --service-version 14 --volume-size 2Gi\n"})}),"\n",(0,r.jsxs)(s.p,{children:["Will generate a YAML spec called ",(0,r.jsx)(s.code,{children:"usermanifests/my-pg-instance.yaml"})," with the following content:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-yaml",children:"apiVersion: postgresql.anynines.com/v1beta3\nkind: Postgresql\nmetadata:\n  name: my-pg\nspec:\n  replicas: 3\n  resources:\n    limits:\n      memory: 200m\n    requests:\n      cpu: 200m\n  version: 14\n  volumeSize: 2Gi\n"})}),"\n",(0,r.jsx)(s.h2,{id:"deleting-a-postgresql-service-instance",children:"Deleting a PostgreSQL Service Instance"}),"\n",(0,r.jsxs)(s.p,{children:["Deleting a service instance with the name ",(0,r.jsx)(s.code,{children:"sample-pg-cluster"}),":"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s delete pg instance --name sample-pg-cluster\n"})}),"\n",(0,r.jsx)(s.p,{children:"Or by providing an explicit namespace:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s delete pg instance --name sample-pg-cluster -n default\n"})}),"\n",(0,r.jsxs)(s.p,{children:[(0,r.jsx)(s.strong,{children:"Note"}),": If the service instance doesn't exist, a warning is printed and the command exists with the\nreturn code ",(0,r.jsx)(s.code,{children:"0"})," as the desired state of the service instance being delete is reached."]}),"\n",(0,r.jsx)(s.h2,{id:"applying-a-sql-file-to-a-postgresql-service-instance",children:"Applying a SQL File to a PostgreSQL Service Instance"}),"\n",(0,r.jsxs)(s.p,{children:["Uploading a SQL file, executing it using ",(0,r.jsx)(s.code,{children:"psql"})," and deleting the file can be done with:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s pg apply --file /path/to/sql/file --service-instance sample-pg-cluster\n"})}),"\n",(0,r.jsx)(s.p,{children:"The file is uploaded to the current primary pod of the service instance."}),"\n",(0,r.jsxs)(s.p,{children:[(0,r.jsx)(s.strong,{children:"Note"}),": Ensure that, during the execution of the command, there is no change of the primary node for a given clustered service instance as otherwise the file upload may fail or target the wrong pod."]}),"\n",(0,r.jsxs)(s.p,{children:["Use ",(0,r.jsx)(s.code,{children:"--yes"})," to skip the confirmation prompt."]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s pg apply --file /path/to/sql/file --service-instance sample-pg-cluster --yes\n"})}),"\n",(0,r.jsxs)(s.p,{children:["Use ",(0,r.jsx)(s.code,{children:"--no-delete"})," to leave the file in the pod:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s pg apply --file /path/to/sql/file --service-instance sample-pg-cluster --no-delete\n"})}),"\n",(0,r.jsx)(s.h2,{id:"applying-a-sql-statement-to-a-postgresql-service-instance",children:"Applying a SQL Statement to a PostgreSQL Service Instance"}),"\n",(0,r.jsx)(s.p,{children:"Applying a SQL statement on the primary pod of a PostgreSQL service instance can be accomplished with:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:'a9s pg apply -i sample-pg-cluster --sql "select count(*) from posts" --yes\n'})}),"\n",(0,r.jsx)(s.h2,{id:"creating-a-backup-of-a-postgresql-service-instance",children:"Creating a Backup of a PostgreSQL Service Instance"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create pg backup --name sample-pg-cluster-backup-1 -i sample-pg-cluster\n"})}),"\n",(0,r.jsx)(s.h2,{id:"restoring-a-backup-of-postgresql-service-instance",children:"Restoring a Backup of PostgreSQL Service Instance"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create pg restore --name sample-pg-cluster-restore-1 -b sample-pg-cluster-backup-1 -i sample-pg-cluster\n"})}),"\n",(0,r.jsx)(s.h2,{id:"creating-a-postgresql-service-binding",children:"Creating a PostgreSQL Service Binding"}),"\n",(0,r.jsx)(s.p,{children:"A Service Binding is an entity facilitating the secure consumption of a service instance.\nBy creating a service instance, a Postgres user is created along with a corresponding Kubernetes Secret."}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s create pg servicebinding --name sb-clustered-1 -i sample-pg-cluster\n"})}),"\n",(0,r.jsxs)(s.p,{children:["Will therefore create a Kubernetes Secret named ",(0,r.jsx)(s.code,{children:"sb-clustered-1-service-binding"})," and provide the following\nkeys containing everything an application needs to connect to the PostgreSQL service instance:"]}),"\n",(0,r.jsxs)(s.ul,{children:["\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.code,{children:"database"})}),"\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.code,{children:"instance_service"})}),"\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.code,{children:"password"})}),"\n",(0,r.jsx)(s.li,{children:(0,r.jsx)(s.code,{children:"username"})}),"\n"]}),"\n",(0,r.jsx)(s.h1,{id:"cleaning-up",children:"Cleaning Up"}),"\n",(0,r.jsx)(s.p,{children:"In order to delete the cluster run:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"a9s delete cluster a8s\n"})}),"\n",(0,r.jsxs)(s.p,{children:[(0,r.jsx)(s.strong,{children:"Note"}),": This will not delete config files."]}),"\n",(0,r.jsx)(s.p,{children:"Config files are stored in the cluster working directory."}),"\n",(0,r.jsx)(s.p,{children:"They can be removed with:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{className:"language-bash",children:"rm -rf $( a9s cluster pwd )\n"})})]})}function h(e={}){const{wrapper:s}={...(0,a.R)(),...e.components};return s?(0,r.jsx)(s,{...e,children:(0,r.jsx)(d,{...e})}):d(e)}},8453:(e,s,n)=>{n.d(s,{R:()=>t,x:()=>c});var i=n(6540);const r={},a=i.createContext(r);function t(e){const s=i.useContext(a);return i.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function c(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(r):e.components||r:t(e.components),i.createElement(a.Provider,{value:s},e.children)}}}]);