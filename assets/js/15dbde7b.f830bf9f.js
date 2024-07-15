"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[2534],{4137:function(e,t,n){n.d(t,{Zo:function(){return c},kt:function(){return m}});var a=n(7294);function r(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function i(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function s(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?i(Object(n),!0).forEach((function(t){r(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):i(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function l(e,t){if(null==e)return{};var n,a,r=function(e,t){if(null==e)return{};var n,a,r={},i=Object.keys(e);for(a=0;a<i.length;a++)n=i[a],t.indexOf(n)>=0||(r[n]=e[n]);return r}(e,t);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(a=0;a<i.length;a++)n=i[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(r[n]=e[n])}return r}var o=a.createContext({}),p=function(e){var t=a.useContext(o),n=t;return e&&(n="function"==typeof e?e(t):s(s({},t),e)),n},c=function(e){var t=p(e.components);return a.createElement(o.Provider,{value:t},e.children)},u="mdxType",d={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},k=a.forwardRef((function(e,t){var n=e.components,r=e.mdxType,i=e.originalType,o=e.parentName,c=l(e,["components","mdxType","originalType","parentName"]),u=p(n),k=r,m=u["".concat(o,".").concat(k)]||u[k]||d[k]||i;return n?a.createElement(m,s(s({ref:t},c),{},{components:n})):a.createElement(m,s({ref:t},c))}));function m(e,t){var n=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var i=n.length,s=new Array(i);s[0]=k;var l={};for(var o in t)hasOwnProperty.call(t,o)&&(l[o]=t[o]);l.originalType=e,l[u]="string"==typeof e?e:r,s[1]=l;for(var p=2;p<i;p++)s[p]=n[p];return a.createElement.apply(null,s)}return a.createElement.apply(null,n)}k.displayName="MDXCreateElement"},4022:function(e,t,n){n.r(t),n.d(t,{assets:function(){return o},contentTitle:function(){return s},default:function(){return d},frontMatter:function(){return i},metadata:function(){return l},toc:function(){return p}});var a=n(3117),r=(n(7294),n(4137));const i={id:"a9s-cli",title:"a9s CLI",tags:["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind"],keywords:["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind"]},s="a9s CLI",l={unversionedId:"a9s-cli",id:"a9s-cli",title:"a9s CLI",description:"anynines provides a command line tool called a9s to facilitate application development, devops tasks and interact with selected anynines products.",source:"@site/docs/a9s-cli.md",sourceDirName:".",slug:"/a9s-cli",permalink:"/docs/develop/a9s-cli",draft:!1,tags:[{label:"a9s cli",permalink:"/docs/develop/tags/a-9-s-cli"},{label:"a9s hub",permalink:"/docs/develop/tags/a-9-s-hub"},{label:"a9s data services",permalink:"/docs/develop/tags/a-9-s-data-services"},{label:"a8s data services",permalink:"/docs/develop/tags/a-8-s-data-services"},{label:"a9s postgres",permalink:"/docs/develop/tags/a-9-s-postgres"},{label:"a8s postgres",permalink:"/docs/develop/tags/a-8-s-postgres"},{label:"data service",permalink:"/docs/develop/tags/data-service"},{label:"introduction",permalink:"/docs/develop/tags/introduction"},{label:"kubernetes",permalink:"/docs/develop/tags/kubernetes"},{label:"minikube",permalink:"/docs/develop/tags/minikube"},{label:"kind",permalink:"/docs/develop/tags/kind"}],version:"current",frontMatter:{id:"a9s-cli",title:"a9s CLI",tags:["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind"],keywords:["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind"]},sidebar:"tutorialSidebar",next:{title:"Hands-On Tutorials",permalink:"/docs/develop/hands-on-tutorials/"}},o={},p=[{value:"Use Cases",id:"use-cases",level:2},{value:"<code>a8s</code> Stack",id:"a8s-stack",level:3},{value:"Cold-Run",id:"cold-run",level:2},{value:"Setting Up a Working Directory",id:"setting-up-a-working-directory",level:3},{value:"Configuring the Backup Store",id:"configuring-the-backup-store",level:3},{value:"Skip Checking Prerequisites",id:"skip-checking-prerequisites",level:2},{value:"Number of Kubernetes Nodes",id:"number-of-kubernetes-nodes",level:2},{value:"Cluster Memory",id:"cluster-memory",level:2},{value:"Deployment Version",id:"deployment-version",level:2},{value:"Kubernetes Provider",id:"kubernetes-provider",level:2},{value:"Backup Infrastructure Region",id:"backup-infrastructure-region",level:2},{value:"Unattended Mode",id:"unattended-mode",level:2},{value:"Printing the Working Directory",id:"printing-the-working-directory",level:2},{value:"Creating a PostgreSQL Service Instance",id:"creating-a-postgresql-service-instance",level:2},{value:"Creating PostgreSQL Service Instance YAML Without Applying it",id:"creating-postgresql-service-instance-yaml-without-applying-it",level:3},{value:"Creating a Custom PostgreSQL Service Instance",id:"creating-a-custom-postgresql-service-instance",level:3},{value:"Deleting a PostgreSQL Service Instance",id:"deleting-a-postgresql-service-instance",level:2},{value:"Applying a SQL File to a PostgreSQL Service Instance",id:"applying-a-sql-file-to-a-postgresql-service-instance",level:2},{value:"Applying a SQL Statement to a PostgreSQL Service Instance",id:"applying-a-sql-statement-to-a-postgresql-service-instance",level:2},{value:"Creating a Backup of a PostgreSQL Service Instance",id:"creating-a-backup-of-a-postgresql-service-instance",level:2},{value:"Restoring a Backup of PostgreSQL Service Instance",id:"restoring-a-backup-of-postgresql-service-instance",level:2},{value:"Creating a PostgreSQL Service Binding",id:"creating-a-postgresql-service-binding",level:2}],c={toc:p},u="wrapper";function d(e){let{components:t,...n}=e;return(0,r.kt)(u,(0,a.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"a9s-cli"},"a9s CLI"),(0,r.kt)("p",null,"anynines provides a command line tool called ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," to facilitate application development, devops tasks and interact with selected anynines products."),(0,r.kt)("h2",{id:"use-cases"},"Use Cases"),(0,r.kt)("p",null,"The ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI can be used to install and use the following stacks:"),(0,r.kt)("h3",{id:"a8s-stack"},(0,r.kt)("inlineCode",{parentName:"h3"},"a8s")," Stack"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},"Install a local Kubernetes cluster (",(0,r.kt)("inlineCode",{parentName:"li"},"minikube")," or ",(0,r.kt)("inlineCode",{parentName:"li"},"kind"),")."),(0,r.kt)("li",{parentName:"ul"},"Install the ",(0,r.kt)("a",{parentName:"li",href:"https://cert-manager.io/"},"cert-manager"),"."),(0,r.kt)("li",{parentName:"ul"},"Install a local Minio object store for storing Backups."),(0,r.kt)("li",{parentName:"ul"},"Install the a8s PostgreSQL Operator PostgreSQL supporting",(0,r.kt)("ul",{parentName:"li"},(0,r.kt)("li",{parentName:"ul"},"creating dedicated PostgreSQL clusters with ",(0,r.kt)("ul",{parentName:"li"},(0,r.kt)("li",{parentName:"ul"},"synchronous and asynchronous streaming replication."),(0,r.kt)("li",{parentName:"ul"},"automatic failure detection and automatic failover."))),(0,r.kt)("li",{parentName:"ul"},"backup and restore capabilities storing backups in an S3 compatible object store such as AWS S3 or Minio."),(0,r.kt)("li",{parentName:"ul"},"ability to easily create database users and Kubernetes Secrets by using the Service Bindings abstraction"))),(0,r.kt)("li",{parentName:"ul"},"Easily apply ",(0,r.kt)("inlineCode",{parentName:"li"},".sql")," files and SQL commands to PostgreSQL clusters.")),(0,r.kt)("h1",{id:"prerequisites"},"Prerequisites"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},"Using the backup/restore feature of a8s PostgreSQL requires an S3 compatible endpoint."),(0,r.kt)("li",{parentName:"ul"},"Install Go (if you want ",(0,r.kt)("inlineCode",{parentName:"li"},"go env")," to identify your OS and arch)."),(0,r.kt)("li",{parentName:"ul"},"Install Git."),(0,r.kt)("li",{parentName:"ul"},"Install Docker."),(0,r.kt)("li",{parentName:"ul"},"Install Kubectl."),(0,r.kt)("li",{parentName:"ul"},"Install Kind and/or Minikube."),(0,r.kt)("li",{parentName:"ul"},"Install the ",(0,r.kt)("a",{parentName:"li",href:"https://cert-manager.io/docs/reference/cmctl/"},"cert-manager CLI"),".")),(0,r.kt)("h1",{id:"installing-the-cli"},"Installing the CLI"),(0,r.kt)("p",null,"In order to install the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI execute the following shell script:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"RELEASE=$(curl -L -s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/stable.txt); OS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o a9s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/releases/$RELEASE/a9s-$OS-$ARCH\n    \nsudo chmod 755 a9s\nsudo mv a9s /usr/local/bin\n")),(0,r.kt)("p",null,"This will download the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," binary suitable for your architecture and move it to ",(0,r.kt)("inlineCode",{parentName:"p"},"/usr/local/bin"),".\nDepending on your system you have to adjust the ",(0,r.kt)("inlineCode",{parentName:"p"},"PATH")," variable or move the binary to a folder that's already in the ",(0,r.kt)("inlineCode",{parentName:"p"},"PATH"),"."),(0,r.kt)("h1",{id:"using-the-cli"},"Using the CLI"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s\n")),(0,r.kt)("h1",{id:"creating-a-local-a8s-postgres-cluster"},"Creating a Local a8s Postgres Cluster"),(0,r.kt)("p",null,"Create a local Kubernetes cluster using ",(0,r.kt)("inlineCode",{parentName:"p"},"Minikube")," or ",(0,r.kt)("inlineCode",{parentName:"p"},"Kind"),", ",(0,r.kt)("strong",{parentName:"p"},"install a8s PostgreSQL")," including its dependencies as well as a local ",(0,r.kt)("a",{parentName:"p",href:"https://min.io/"},"Minio")," object store. "),(0,r.kt)("p",null,"Get ready for ",(0,r.kt)("strong",{parentName:"p"},"local development of applications with PostgreSQL")," and/or ",(0,r.kt)("strong",{parentName:"p"},"experimentation with a8s Postgres")," by issuing the command:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s\n")),(0,r.kt)("p",null,"Recommended is 12 GB of free memory for the creation of three cluster nodes with each 4 GB. The number of nodes and memory size can be adjusted."),(0,r.kt)("h2",{id:"cold-run"},"Cold-Run"),(0,r.kt)("p",null,"When creating a cluster for the first time, a few setup steps will have to be taken which need to be performed only once:"),(0,r.kt)("ol",null,(0,r.kt)("li",{parentName:"ol"},"Setting up a working directory for the use with the ",(0,r.kt)("inlineCode",{parentName:"li"},"a9s")," CLI. ",(0,r.kt)("strong",{parentName:"li"},"This step asks for your confirmation of the proposed directory.")),(0,r.kt)("li",{parentName:"ol"},"Configuring the access credentials for the S3 compatible object store which is needed to use the backup/restore feature of a8s Postgres. This step is performed automatically."),(0,r.kt)("li",{parentName:"ol"},"Cloning deployment resources required by the ",(0,r.kt)("inlineCode",{parentName:"li"},"a9s")," CLI to create a cluster. This step is performed automatically.")),(0,r.kt)("h3",{id:"setting-up-a-working-directory"},"Setting Up a Working Directory"),(0,r.kt)("p",null,"The working directory is where are ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI related resources will go. This includes ",(0,r.kt)("inlineCode",{parentName:"p"},"yaml")," specifications being cloned from remote repositories, but also those generated by the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI for your convenience."),(0,r.kt)("p",null,"Once established, the working directory is stored in the ",(0,r.kt)("inlineCode",{parentName:"p"},"~/.a9s")," configuration file."),(0,r.kt)("p",null,"The default working directory is ",(0,r.kt)("inlineCode",{parentName:"p"},"~/a9s"),"."),(0,r.kt)("p",null,"Alternatively, provide a custom working directory at the corresponding prompt."),(0,r.kt)("h3",{id:"configuring-the-backup-store"},"Configuring the Backup Store"),(0,r.kt)("p",null,"A non-prod Minio object store is installed in your local Kubernetes cluster and is automatically configured as the default backup store for a8s PostgreSQL backups."),(0,r.kt)("p",null,"If you want to use an alternative backup store, see ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s create cluster a8s --help")," for the defaults of your particular CLI version and list of configuration options."),(0,r.kt)("p",null,"Most S3 compatible object stores, including AWS S3 itself of course, should work."),(0,r.kt)("h2",{id:"skip-checking-prerequisites"},"Skip Checking Prerequisites"),(0,r.kt)("p",null,"It is possible to skip the verification of prerequisites. This includes skipping the search for: required shell commands, a running Docker daemon and a running Kubernetes cluster."),(0,r.kt)("p",null,"In order to skip precheck use the ",(0,r.kt)("inlineCode",{parentName:"p"},"--no-precheck")," option:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --no-precheck\n")),(0,r.kt)("h2",{id:"number-of-kubernetes-nodes"},"Number of Kubernetes Nodes"),(0,r.kt)("p",null,"Specifying the number of Nodes in the Kubernetes cluster:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --cluster-nr-of-nodes 1\n")),(0,r.kt)("h2",{id:"cluster-memory"},"Cluster Memory"),(0,r.kt)("p",null,"Specifying the memory of ",(0,r.kt)("strong",{parentName:"p"},"each")," Node of the Kubernetes cluster:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --cluster-memory 4gb\n")),(0,r.kt)("h2",{id:"deployment-version"},"Deployment Version"),(0,r.kt)("p",null,"The deployment version refers to the version of manifests used for installing software. Deployment versions are managed by anynines in a Git repository. The deployment version option allows you to select a particular version of the deployment manifests identified by ",(0,r.kt)("strong",{parentName:"p"},"Git tags"),"."),(0,r.kt)("p",null,"Select a particular release by providing the ",(0,r.kt)("inlineCode",{parentName:"p"},"--deployment-version")," parameter:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --deployment-version v1.2.0\n")),(0,r.kt)("p",null,"Use:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --deployment-version latest\n")),(0,r.kt)("p",null,"To get the latest, untagged version of the deployment manifests."),(0,r.kt)("h2",{id:"kubernetes-provider"},"Kubernetes Provider"),(0,r.kt)("p",null,"When creating a Kubernetes cluster, the mechanism to manage the cluster can be selected by specifying the ",(0,r.kt)("inlineCode",{parentName:"p"},"--provider")," option:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s -p kind \na9s create cluster a8s -p minikube (default)\n")),(0,r.kt)("p",null,"Follow the instructions to learn about available sub commands."),(0,r.kt)("h2",{id:"backup-infrastructure-region"},"Backup Infrastructure Region"),(0,r.kt)("p",null,"When using the backup and restore functionality, a backup infrastructure region must be specified by using the ",(0,r.kt)("inlineCode",{parentName:"p"},"--backup-region")," option:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --backup-region us-east-1\n")),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"Note"),": By default, an existing ",(0,r.kt)("inlineCode",{parentName:"p"},"backup-config.yaml")," will be used. Hence, if you intend to change\nyour backup config, remove the existing ",(0,r.kt)("inlineCode",{parentName:"p"},"backup-config.yaml"),", first:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"rm a8s-deployment/deploy/a8s/backup-config/backup-store-config.yaml\n")),(0,r.kt)("h2",{id:"unattended-mode"},"Unattended Mode"),(0,r.kt)("p",null,"It is possible to skip all yes-no questions by ",(0,r.kt)("strong",{parentName:"p"},"enabling the unattended mode")," by passing the ",(0,r.kt)("inlineCode",{parentName:"p"},"-y")," or ",(0,r.kt)("inlineCode",{parentName:"p"},"--yes")," flag:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --yes\n")),(0,r.kt)("h2",{id:"printing-the-working-directory"},"Printing the Working Directory"),(0,r.kt)("p",null,"The working directory is stored in the ",(0,r.kt)("inlineCode",{parentName:"p"},"~/.a8s")," configuration file. The working directory contains all resources downloaded and generated by the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI."),(0,r.kt)("p",null,"To print the working directory execute:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s cluster pwd\n")),(0,r.kt)("h1",{id:"a8s-postgresql"},"a8s PostgreSQL"),(0,r.kt)("p",null,"A selected subset of the a8s PostgreSQL features are available through the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI."),(0,r.kt)("h2",{id:"creating-a-postgresql-service-instance"},"Creating a PostgreSQL Service Instance"),(0,r.kt)("p",null,"Creating a service instance with the name ",(0,r.kt)("inlineCode",{parentName:"p"},"sample-pg-cluster"),":"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg instance --name sample-pg-cluster\n")),(0,r.kt)("p",null,"The generated YAML specification will be stored in the ",(0,r.kt)("inlineCode",{parentName:"p"},"usermanifests"),"."),(0,r.kt)("h3",{id:"creating-postgresql-service-instance-yaml-without-applying-it"},"Creating PostgreSQL Service Instance YAML Without Applying it"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg instance --name sample-pg-cluster --no-apply\n")),(0,r.kt)("p",null,"The generated YAML specification will be stored in the ",(0,r.kt)("inlineCode",{parentName:"p"},"usermanifests")," but ",(0,r.kt)("inlineCode",{parentName:"p"},"kubectl apply")," won't be executed."),(0,r.kt)("h3",{id:"creating-a-custom-postgresql-service-instance"},"Creating a Custom PostgreSQL Service Instance"),(0,r.kt)("p",null,"The command:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg instance --api-version v1beta3 --name sample-pg-cluster --namespace default --replicas 3 --requests-cpu 200m --limits-memory 200Mi --service-version 14 --volume-size 2Gi\n")),(0,r.kt)("p",null,"Will generate a YAML spec called ",(0,r.kt)("inlineCode",{parentName:"p"},"usermanifests/my-pg-instance.yaml")," with the following content:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},"apiVersion: postgresql.anynines.com/v1beta3\nkind: Postgresql\nmetadata:\n  name: my-pg\nspec:\n  replicas: 3\n  resources:\n    limits:\n      memory: 200m\n    requests:\n      cpu: 200m\n  version: 14\n  volumeSize: 2Gi\n")),(0,r.kt)("h2",{id:"deleting-a-postgresql-service-instance"},"Deleting a PostgreSQL Service Instance"),(0,r.kt)("p",null,"Deleting a service instance with the name ",(0,r.kt)("inlineCode",{parentName:"p"},"sample-pg-cluster"),":"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s delete pg instance --name sample-pg-cluster\n")),(0,r.kt)("p",null,"Or by providing an explicit namespace:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s delete pg instance --name sample-pg-cluster -n default\n")),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"Note"),": If the service instance doesn't exist, a warning is printed and the command exists with the\nreturn code ",(0,r.kt)("inlineCode",{parentName:"p"},"0")," as the desired state of the service instance being delete is reached."),(0,r.kt)("h2",{id:"applying-a-sql-file-to-a-postgresql-service-instance"},"Applying a SQL File to a PostgreSQL Service Instance"),(0,r.kt)("p",null,"Uploading a SQL file, executing it using ",(0,r.kt)("inlineCode",{parentName:"p"},"psql")," and deleting the file can be done with:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s pg apply --file /path/to/sql/file --service-instance sample-pg-cluster\n")),(0,r.kt)("p",null,"The file is uploaded to the current primary pod of the service instance. "),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"Note"),": Ensure that, during the execution of the command, there is no change of the primary node for a given clustered service instance as otherwise the file upload may fail or target the wrong pod."),(0,r.kt)("p",null,"Use ",(0,r.kt)("inlineCode",{parentName:"p"},"--yes")," to skip the confirmation prompt."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s pg apply --file /path/to/sql/file --service-instance sample-pg-cluster --yes\n")),(0,r.kt)("p",null,"Use ",(0,r.kt)("inlineCode",{parentName:"p"},"--no-delete")," to leave the file in the pod:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s pg apply --file /path/to/sql/file --service-instance sample-pg-cluster --no-delete\n")),(0,r.kt)("h2",{id:"applying-a-sql-statement-to-a-postgresql-service-instance"},"Applying a SQL Statement to a PostgreSQL Service Instance"),(0,r.kt)("p",null,"Applying a SQL statement on the primary pod of a PostgreSQL service instance can be accomplished with:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'a9s pg apply -i sample-pg-cluster --sql "select count(*) from posts" --yes\n')),(0,r.kt)("h2",{id:"creating-a-backup-of-a-postgresql-service-instance"},"Creating a Backup of a PostgreSQL Service Instance"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg backup --name sample-pg-cluster-backup-1 -i sample-pg-cluster\n")),(0,r.kt)("h2",{id:"restoring-a-backup-of-postgresql-service-instance"},"Restoring a Backup of PostgreSQL Service Instance"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg restore --name sample-pg-cluster-restore-1 -b sample-pg-cluster-backup-1 -i sample-pg-cluster\n")),(0,r.kt)("h2",{id:"creating-a-postgresql-service-binding"},"Creating a PostgreSQL Service Binding"),(0,r.kt)("p",null,"A Service Binding is an entity facilitating the secure consumption of a service instance.\nBy creating a service instance, a Postgres user is created along with a corresponding Kubernetes Secret."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg servicebinding --name sb-clustered-1 -i sample-pg-cluster\n")),(0,r.kt)("p",null,"Will therefore create a Kubernetes Secret named ",(0,r.kt)("inlineCode",{parentName:"p"},"sb-clustered-1-service-binding")," and provide the following\nkeys containing everything an application needs to connect to the PostgreSQL service instance:"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("inlineCode",{parentName:"li"},"database")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("inlineCode",{parentName:"li"},"instance_service")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("inlineCode",{parentName:"li"},"password")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("inlineCode",{parentName:"li"},"username"))),(0,r.kt)("h1",{id:"cleaning-up"},"Cleaning Up"),(0,r.kt)("p",null,"In order to delete the cluster run:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s delete cluster a8s\n")),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"Note"),": This will not delete config files."),(0,r.kt)("p",null,"Config files are stored in the cluster working directory."),(0,r.kt)("p",null,"They can be removed with:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"rm -rf $( a9s cluster pwd )\n")))}d.isMDXComponent=!0}}]);