"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[134],{4137:function(e,t,a){a.d(t,{Zo:function(){return c},kt:function(){return h}});var n=a(7294);function r(e,t,a){return t in e?Object.defineProperty(e,t,{value:a,enumerable:!0,configurable:!0,writable:!0}):e[t]=a,e}function i(e,t){var a=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),a.push.apply(a,n)}return a}function s(e){for(var t=1;t<arguments.length;t++){var a=null!=arguments[t]?arguments[t]:{};t%2?i(Object(a),!0).forEach((function(t){r(e,t,a[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(a)):i(Object(a)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(a,t))}))}return e}function l(e,t){if(null==e)return{};var a,n,r=function(e,t){if(null==e)return{};var a,n,r={},i=Object.keys(e);for(n=0;n<i.length;n++)a=i[n],t.indexOf(a)>=0||(r[a]=e[a]);return r}(e,t);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(n=0;n<i.length;n++)a=i[n],t.indexOf(a)>=0||Object.prototype.propertyIsEnumerable.call(e,a)&&(r[a]=e[a])}return r}var o=n.createContext({}),p=function(e){var t=n.useContext(o),a=t;return e&&(a="function"==typeof e?e(t):s(s({},t),e)),a},c=function(e){var t=p(e.components);return n.createElement(o.Provider,{value:t},e.children)},u="mdxType",d={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},m=n.forwardRef((function(e,t){var a=e.components,r=e.mdxType,i=e.originalType,o=e.parentName,c=l(e,["components","mdxType","originalType","parentName"]),u=p(a),m=r,h=u["".concat(o,".").concat(m)]||u[m]||d[m]||i;return a?n.createElement(h,s(s({ref:t},c),{},{components:a})):n.createElement(h,s({ref:t},c))}));function h(e,t){var a=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var i=a.length,s=new Array(i);s[0]=m;var l={};for(var o in t)hasOwnProperty.call(t,o)&&(l[o]=t[o]);l.originalType=e,l[u]="string"==typeof e?e:r,s[1]=l;for(var p=2;p<i;p++)s[p]=a[p];return n.createElement.apply(null,s)}return n.createElement.apply(null,a)}m.displayName="MDXCreateElement"},5861:function(e,t,a){a.r(t),a.d(t,{assets:function(){return o},contentTitle:function(){return s},default:function(){return d},frontMatter:function(){return i},metadata:function(){return l},toc:function(){return p}});var n=a(3117),r=(a(7294),a(4137));const i={id:"hands-on-tutorial-a8s-pg-a9s-cli",title:"Deploying a Demo App using a8s PostgreSQL",tags:["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind"],keywords:["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind","postgresql","web app"]},s="Overview",l={unversionedId:"hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli",id:"version-0.11.1/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli",title:"Deploying a Demo App using a8s PostgreSQL",description:"What you will accomplish",source:"@site/versioned_docs/version-0.11.1/hands-on-tutorials/a9s-cli-a8s-postgresql.md",sourceDirName:"hands-on-tutorials",slug:"/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli",permalink:"/docs/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli",draft:!1,tags:[{label:"a9s hub",permalink:"/docs/tags/a-9-s-hub"},{label:"a9s cli",permalink:"/docs/tags/a-9-s-cli"},{label:"a8s data services",permalink:"/docs/tags/a-8-s-data-services"},{label:"a8s postgres",permalink:"/docs/tags/a-8-s-postgres"},{label:"data service",permalink:"/docs/tags/data-service"},{label:"tutorial",permalink:"/docs/tags/tutorial"},{label:"kubernetes",permalink:"/docs/tags/kubernetes"},{label:"minikube",permalink:"/docs/tags/minikube"},{label:"kind",permalink:"/docs/tags/kind"}],version:"0.11.1",frontMatter:{id:"hands-on-tutorial-a8s-pg-a9s-cli",title:"Deploying a Demo App using a8s PostgreSQL",tags:["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind"],keywords:["a9s hub","a9s cli","a8s data services","a8s postgres","data service","tutorial","kubernetes","minikube","kind","postgresql","web app"]},sidebar:"tutorialSidebar",previous:{title:"Hands-On Tutorials",permalink:"/docs/hands-on-tutorials/"}},o={},p=[{value:"What you will accomplish",id:"what-you-will-accomplish",level:2},{value:"What you will learn",id:"what-you-will-learn",level:2},{value:"Prerequisites",id:"prerequisites",level:2},{value:"Step 1: Creating a Kubernetes Cluster with a8s PostgreSQL",id:"step-1-creating-a-kubernetes-cluster-with-a8s-postgresql",level:2},{value:"Step 1.1: Initial Configuration on the First a9s create cluster Execution",id:"step-11-initial-configuration-on-the-first-a9s-create-cluster-execution",level:3},{value:"What&#39;s Happening During the Installation",id:"whats-happening-during-the-installation",level:3},{value:"Cert-Manager",id:"cert-manager",level:4},{value:"a8s PostgreSQL",id:"a8s-postgresql",level:4},{value:"Step 2: Creating a PostgreSQL Cluster",id:"step-2-creating-a-postgresql-cluster",level:2},{value:"Inspecting the Service Instance",id:"inspecting-the-service-instance",level:3},{value:"Step 3: Creating a Service Binding",id:"step-3-creating-a-service-binding",level:2},{value:"Step 4: Deploying a Demo Application",id:"step-4-deploying-a-demo-application",level:2},{value:"Step 5: Interacting with PostgreSQL",id:"step-5-interacting-with-postgresql",level:2},{value:"Applying a Local SQL File",id:"applying-a-local-sql-file",level:3},{value:"Applying an SQL String",id:"applying-an-sql-string",level:3},{value:"(Optional) Step 6: Creating and Restoring a Backup",id:"optional-step-6-creating-and-restoring-a-backup",level:2},{value:"Creating a Backup",id:"creating-a-backup",level:3},{value:"Restoring a Backup",id:"restoring-a-backup",level:3},{value:"Congratulations",id:"congratulations",level:2},{value:"What to do next?",id:"what-to-do-next",level:2},{value:"Links",id:"links",level:2}],c={toc:p},u="wrapper";function d(e){let{components:t,...a}=e;return(0,r.kt)(u,(0,n.Z)({},c,a,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"overview"},"Overview"),(0,r.kt)("h2",{id:"what-you-will-accomplish"},"What you will accomplish"),(0,r.kt)("p",null,"In this tutorial you will learn how to ",(0,r.kt)("strong",{parentName:"p"},"create a local Kubernetes cluster"),", fully equipped ",(0,r.kt)("strong",{parentName:"p"},"with a PostgreSQL")," operator, ready for you to deploy a PostgreSQL database instance for ",(0,r.kt)("strong",{parentName:"p"},"developing your application"),"."),(0,r.kt)("h2",{id:"what-you-will-learn"},"What you will learn"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},"Install the ",(0,r.kt)("a",{parentName:"li",href:"https://github.com/anynines/a9s-cli-v2"},"a9s CLI")),(0,r.kt)("li",{parentName:"ul"},"Create a local Kubernetes cluster"),(0,r.kt)("li",{parentName:"ul"},"Install ",(0,r.kt)("a",{parentName:"li",href:"https://cert-manager.io/docs/"},"cert-manager")),(0,r.kt)("li",{parentName:"ul"},"Install a8s PostgreSQL"),(0,r.kt)("li",{parentName:"ul"},"Create a PostgreSQL database instance"),(0,r.kt)("li",{parentName:"ul"},"Create a PostgreSQL user"),(0,r.kt)("li",{parentName:"ul"},"Connect to the PostgreSQL database"),(0,r.kt)("li",{parentName:"ul"},"Deploy a demo application"),(0,r.kt)("li",{parentName:"ul"},"Connect the application to the PostgreSQL database"),(0,r.kt)("li",{parentName:"ul"},"Create a backup"),(0,r.kt)("li",{parentName:"ul"},"Restore a backup")),(0,r.kt)("h2",{id:"prerequisites"},"Prerequisites"),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},"MacOS / Linux"),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"https://www.docker.com/"},"Docker")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"https://minikube.sigs.k8s.io/docs/start/"},"Minikube")," or ",(0,r.kt)("a",{parentName:"li",href:"https://kind.sigs.k8s.io/"},"Kind")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"https://github.com/anynines/a9s-cli-v2"},"a9s CLI")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("a",{parentName:"li",href:"https://kubernetes.io/docs/reference/kubectl/"},"Kubectl")),(0,r.kt)("li",{parentName:"ul"},"Optional for backup/restore: AWS S3 Bucket with credentials")),(0,r.kt)("h1",{id:"implementation"},"Implementation"),(0,r.kt)("p",null,"In this tutorial you will be using the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI to facilitate the creation of both a local Kubernetes cluster and a PostgreSQL database instance. "),(0,r.kt)("p",null,"The ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI will guide you through the process while providing you with transparency and ability to set your own pace. Transparency means that you will see the exact commands to be executed. By default, the commands are executed only after you have confirmed the execution by pressing the ",(0,r.kt)("inlineCode",{parentName:"p"},"<ENTER>")," key. This allows you to have a closer look at the command and/or the YAML specifications to understand what the current step in the tutorial is about. If all you care about is the result, the ",(0,r.kt)("inlineCode",{parentName:"p"},"--yes")," option will answer all yes-no questions with yes. See ",(0,r.kt)("a",{parentName:"p",href:"https://github.com/anynines/a9s-cli-v2"},"[1]")," for documentation and source code of the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI."),(0,r.kt)("h2",{id:"step-1-creating-a-kubernetes-cluster-with-a8s-postgresql"},"Step 1: Creating a Kubernetes Cluster with a8s PostgreSQL"),(0,r.kt)("p",null,"In this section you will create a Kubernetes cluster with a8s PostgreSQL and all its dependencies:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s\n")),(0,r.kt)("p",null,"Per default, ",(0,r.kt)("inlineCode",{parentName:"p"},"minikube")," will be used. In case you prefer ",(0,r.kt)("inlineCode",{parentName:"p"},"kind")," you can use the ",(0,r.kt)("inlineCode",{parentName:"p"},"--provider")," option:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create cluster a8s --provider kind\n")),(0,r.kt)("p",null,"The remainder of the tutorial works equally for both ",(0,r.kt)("inlineCode",{parentName:"p"},"minikube")," and ",(0,r.kt)("inlineCode",{parentName:"p"},"kind"),"."),(0,r.kt)("h3",{id:"step-11-initial-configuration-on-the-first-a9s-create-cluster-execution"},"Step 1.1: Initial Configuration on the First a9s create cluster Execution"),(0,r.kt)("p",null,"When creating a cluster for the first time, a few setup steps will have to be taken which need to be performed only once:"),(0,r.kt)("ol",null,(0,r.kt)("li",{parentName:"ol"},"Setting up a working directory for the use with the ",(0,r.kt)("inlineCode",{parentName:"li"},"a9s")," CLI."),(0,r.kt)("li",{parentName:"ol"},"Configuring the access credentials for the S3 compatible object store which is needed if you intend to use the backup/restore feature of a8s Postgres."),(0,r.kt)("li",{parentName:"ol"},"Cloning deployment resources required by the ",(0,r.kt)("inlineCode",{parentName:"li"},"a9s")," CLI to create a cluster.")),(0,r.kt)("p",null,"In order to ",(0,r.kt)("strong",{parentName:"p"},"optionally skip the backup store configuration, enter random credentials for the backup store's ",(0,r.kt)("inlineCode",{parentName:"strong"},"ACCESSKEYID")," and ",(0,r.kt)("inlineCode",{parentName:"strong"},"SECRETKEY")),", for example, in case you don't have an AWS S3 compatible object store at hand and/or don't want to use the backup/restore functionality of a8s PostgreSQL."),(0,r.kt)("h3",{id:"whats-happening-during-the-installation"},"What's Happening During the Installation"),(0,r.kt)("p",null,"After the initial configuration, the Kubernetes cluster is being created."),(0,r.kt)("h4",{id:"cert-manager"},"Cert-Manager"),(0,r.kt)("p",null,"Once the Kubernetes cluster is ready, the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI proceeds with the installation of the ",(0,r.kt)("a",{parentName:"p",href:"https://cert-manager.io/docs/"},"cert-manager"),". The cert-manager is a Kubernetes extension handling TLS certificates. Among others, in a8s PostgreSQL TSL certificates are used for securing the communication between Kubernetes and the operator."),(0,r.kt)("h4",{id:"a8s-postgresql"},"a8s PostgreSQL"),(0,r.kt)("p",null,"With the cert-manager being ready, the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI continues and installs the a8s PostgreSQL components. Namely, this is "),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},"The PostgreSQL operator"),(0,r.kt)("li",{parentName:"ul"},"The Service Binding controller"),(0,r.kt)("li",{parentName:"ul"},"The Backup Manager")),(0,r.kt)("p",null,"The ",(0,r.kt)("strong",{parentName:"p"},"PostgreSQL Operator")," is responsible for creating and managing ",(0,r.kt)("em",{parentName:"p"},"Service Instances"),", that is dedicated PostgreSQL servers represented by a single or a cluster of Pods."),(0,r.kt)("p",null,"The ",(0,r.kt)("strong",{parentName:"p"},"Service Binding Controller"),", as the name suggests, is responsible for creating so-called ",(0,r.kt)("em",{parentName:"p"},"Service Bindings"),". A Service Binding represents ",(0,r.kt)("strong",{parentName:"p"},"a unique set of credentials")," connecting a database client, such as an application and a Service Instance, in this case a PostgreSQL instance. In the case of a8s PostgreSQL, a Service Binding contains a ",(0,r.kt)("strong",{parentName:"p"},"username/password")," combination as well as other information necessary to establish a connection such as the ",(0,r.kt)("strong",{parentName:"p"},"hostname"),"."),(0,r.kt)("p",null,"The ",(0,r.kt)("strong",{parentName:"p"},"Backup Manager")," is responsible for managing backup and restore requests and dispatching them to the ",(0,r.kt)("em",{parentName:"p"},"Backup Agents")," located alongside Postgres Service Instances. It is the Backup Agent of a Service Instance that actually triggers the execution, encryption, compression and streaming of backup and restore operations."),(0,r.kt)("p",null,"After ",(0,r.kt)("em",{parentName:"p"},"waiting for a8s Postgres Control Plane to become ready")," the message ",(0,r.kt)("inlineCode",{parentName:"p"},"\ud83c\udf89 The a8s Postgres Control Plane appears to be ready. All expected pods are running.")," indicates that ",(0,r.kt)("strong",{parentName:"p"},"the installation of a8s PostgreSQL was successful"),"."),(0,r.kt)("h2",{id:"step-2-creating-a-postgresql-cluster"},"Step 2: Creating a PostgreSQL Cluster"),(0,r.kt)("p",null,"In order to keep all tutorial resources in one place, create a Kubernetes ",(0,r.kt)("inlineCode",{parentName:"p"},"tutorial")," namespace:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl create namespace tutorial\n")),(0,r.kt)("p",null,"Now that the a8s PostgreSQL Operator and the ",(0,r.kt)("inlineCode",{parentName:"p"},"tutorial")," namespace is ready, it's time to create a database."),(0,r.kt)("p",null,"Using the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s")," CLI the process is as simple as:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg instance --name clustered-instance --replicas 3 -n tutorial\n")),(0,r.kt)("p",null,"This creates a clustered PostgreSQL instance named ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance")," represented as a StatefulSet with ",(0,r.kt)("inlineCode",{parentName:"p"},"3")," Pods. Each Pod runs a PostgreSQL process."),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"Note"),": The ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s CLI")," does not shield you the YAML specs is generated. Quite the opposite, it is intended to provide you with meaningful templates to start with. ",(0,r.kt)("strong",{parentName:"p"},"You can find all YAML specs generated by the ",(0,r.kt)("inlineCode",{parentName:"strong"},"a9s CLI")," in the ",(0,r.kt)("inlineCode",{parentName:"strong"},"usermanifests")," folder in your a9s working directory"),":"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"ls $(a9s demo pwd)/usermanifests\n")),(0,r.kt)("h3",{id:"inspecting-the-service-instance"},"Inspecting the Service Instance"),(0,r.kt)("p",null,"It's worth inspecting the PostgreSQL Service Instance to see what the a8s PostgreSQL Operator has created:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get postgresqls -n tutorial\n")),(0,r.kt)("p",null,"Output:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"NAME                 AGE\nclustered-instance   131m\n")),(0,r.kt)("p",null,"The ",(0,r.kt)("inlineCode",{parentName:"p"},"postgresql")," object named ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance"),", as the name suggests, represents your PostgreSQL instance. It is implemented by a set of Kubernetes Services and a StatefulSet."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get statefulsets -n tutorial\n")),(0,r.kt)("p",null,"The operator has created a Kubernetes StatefulSet with the name ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance"),":"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"NAME                 READY   AGE\nclustered-instance   3/3     89m\n")),(0,r.kt)("p",null,"And the StatefulSet, in turn, manages three Pods, namely:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get pods -n tutorial\n")),(0,r.kt)("p",null,"The following Pods:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"NAME                   READY   STATUS    RESTARTS   AGE\nclustered-instance-0   3/3     Running   0          70m\nclustered-instance-1   3/3     Running   0          68m\nclustered-instance-2   3/3     Running   0          66m\n")),(0,r.kt)("p",null,"Have a closer look at one of them:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl describe pod clustered-instance-0 -n tutorial\n")),(0,r.kt)("p",null,"Especially, look at the ",(0,r.kt)("inlineCode",{parentName:"p"},"Labels")," section in the output:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"Name:             clustered-instance-0\nNamespace:        tutorial\nPriority:         0\nService Account:  clustered-instance\nNode:             a8s-demo-m02/192.168.58.3\nStart Time:       Tue, 12 Mar 2024 08:15:39 +0100\nLabels:           a8s.a9s/dsi-group=postgresql.anynines.com\n                a8s.a9s/dsi-kind=Postgresql\n                a8s.a9s/dsi-name=clustered-instance\n                a8s.a9s/replication-role=master\n                apps.kubernetes.io/pod-index=0\n                controller-revision-hash=clustered-instance-749699f5b9\n                statefulset.kubernetes.io/pod-name=clustered-instance-0\n")),(0,r.kt)("p",null,"The label ",(0,r.kt)("inlineCode",{parentName:"p"},"a8s.a9s/replication-role=master")," indicates that the Pod ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance-0")," is the ",(0,r.kt)("strong",{parentName:"p"},"primary")," PostgreSQL server for the asynchronous streaming replication within the cluster. Don't worry if you are not familiar with this terminology. Just bare in mind that ",(0,r.kt)("strong",{parentName:"p"},"all data altering SQL statements always need to go to the primary Pod"),". There's a mechanism in place that will help with this."),(0,r.kt)("p",null,"By executing:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get services -n tutorial\n")),(0,r.kt)("p",null,"You will see a ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance-master")," Kubernetes service:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE\nclustered-instance-config   ClusterIP   None           <none>        <none>              74m\nclustered-instance-master   ClusterIP   10.105.7.211   <none>        5432/TCP,8008/TCP   75m\n")),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"The ",(0,r.kt)("inlineCode",{parentName:"strong"},"clustered-instance-master")," service provides a reference to the primary PostgreSQL server within the clustered Service Instance"),". As the cluster comes with failure-detection and automatic failover capabilities, the primary role may be assigned to another Pod in the cluster during leading election. However, the ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance-master")," service will be updated so that any application connecting through the ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance-master")," service automatically connects to the ",(0,r.kt)("strong",{parentName:"p"},"current")," primary."),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"Congratulations \ud83c\udf89"),", you've managed to create yourself a highly available PostgreSQL cluster using asynchronous streaming replication."),(0,r.kt)("h2",{id:"step-3-creating-a-service-binding"},"Step 3: Creating a Service Binding"),(0,r.kt)("p",null,"In order to prepare the deployment of an application, the database need to be configured to ",(0,r.kt)("strong",{parentName:"p"},"grant the application access to the PostgreSQL service instance"),". Granting an application running in Kubernetes access to a PostgreSQL database involves the following steps:"),(0,r.kt)("ol",null,(0,r.kt)("li",{parentName:"ol"},(0,r.kt)("p",{parentName:"li"},"Create a unique set of access credentials including a database role as well as a corresponding password. ")),(0,r.kt)("li",{parentName:"ol"},(0,r.kt)("p",{parentName:"li"},"Creating a Kubernetes Secret containing the credentials."))),(0,r.kt)("p",null,"The credential set should be unique to the application and the data service instance. So if a second application, such as a worker process, needs access, a separate credential set and Kubernetes Secret is to be created."),(0,r.kt)("p",null,"With a8s PostgreSQL the process of creating access credentials on-demand is referred to as creating ",(0,r.kt)("em",{parentName:"p"},"Service Bindings"),". In other words, ",(0,r.kt)("strong",{parentName:"p"},"a Service Binding in a8s PostgreSQL is a database role, password which is then stored in a Kubernetes Secret")," to be used by exactly one application."),(0,r.kt)("p",null,"Think about the implication of managing Service Bindings using the Kubernetes API. Instead of writing custom scripts connecting to the database, the creation of a database user is as simple as creating a Kubernetes object. Therefore, ",(0,r.kt)("strong",{parentName:"p"},"Service Bindings facilitate deployments to multiple Kubernetes environments describing application systems entirely using Kubernetes objects"),"."),(0,r.kt)("p",null,"Creating a Service Binding is easy:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg servicebinding --name sb-sample -n tutorial -i clustered-instance\n")),(0,r.kt)("p",null,"Have a look at the resources that have been generated:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get servicebindings -n tutorial\n")),(0,r.kt)("p",null,"Output:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"NAME        AGE\nsb-sample   6s\n")),(0,r.kt)("p",null,"The ",(0,r.kt)("inlineCode",{parentName:"p"},"servicebinding")," object named ",(0,r.kt)("inlineCode",{parentName:"p"},"sb-sample")," is owned by the a8s PostgreSQL Operator or, more precisely, the ServiceBindingController. As part of the Service Binding, a Kubernetes Secret has been created:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get secrets -n tutorial\n")),(0,r.kt)("p",null,"Output:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"NAME                                      TYPE     DATA   AGE\npostgres.credentials.clustered-instance   Opaque   2      9m16s\nsb-sample-service-binding                 Opaque   4      25s\nstandby.credentials.clustered-instance    Opaque   2      9m16s\n")),(0,r.kt)("p",null,"Investigate the Secret ",(0,r.kt)("inlineCode",{parentName:"p"},"sb-sample-service-binding"),":"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get secret sb-sample-service-binding -n tutorial -o yaml\n")),(0,r.kt)("p",null,"Output:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-yaml"},'apiVersion: v1\ndata:\n  database: YTlzX2FwcHNfZGVmYXVsdF9kYg==\n  instance_service: Y2x1c3RlcmVkLWluc3RhbmNlLW1hc3Rlci50dXRvcmlhbA==\n  password: bk1wNGI2WHdMeXUwYVkzWmF4ekExS1VURTNzM2xham4=\n  username: YThzLXNiLWN4cDZCMFRUQg==\nimmutable: true\nkind: Secret\nmetadata:\n  creationTimestamp: "2024-03-12T14:50:33Z"\n  finalizers:\n  - a8s.anynines.com/servicebinding.controller\n  labels:\n    service-binding: "true"\n  name: sb-sample-service-binding\n  namespace: tutorial\n  ownerReferences:\n  - apiVersion: servicebindings.anynines.com/v1beta3\n    blockOwnerDeletion: true\n    controller: true\n    kind: ServiceBinding\n    name: sb-sample\n    uid: e4636254-433a-4e82-a46b-e79fd7f25f58\n  resourceVersion: "2648"\n  uid: ebee4e29-4796-4e9a-8114-ec4d546644a9\ntype: Opaque\n')),(0,r.kt)("p",null,"Note that the values in the ",(0,r.kt)("inlineCode",{parentName:"p"},"data")," hash aren't readable right away as they are base64 encoded. Values can be decoded using the ",(0,r.kt)("inlineCode",{parentName:"p"},"base64")," command, for example:"),(0,r.kt)("p",null,(0,r.kt)("inlineCode",{parentName:"p"},"database:")),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'echo "YTlzX2FwcHNfZGVmYXVsdF9kYg==" | base64 --decode\na9s_apps_default_db\n')),(0,r.kt)("p",null,(0,r.kt)("inlineCode",{parentName:"p"},"instance_service:")),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'echo "Y2x1c3RlcmVkLWluc3RhbmNlLW1hc3Rlci50dXRvcmlhbA==" | base64 --decode\nclustered-instance-master.tutorial\n')),(0,r.kt)("p",null,"Given a Service name, the generic naming pattern in Kubernetes to derive its DNS entry is: ",(0,r.kt)("inlineCode",{parentName:"p"},"{service-name}.{namespace}.svc.{cluster-domain:cluster.local}"),"."),(0,r.kt)("p",null,"Assuming that your Kubernetes' cluster domain is the default ",(0,r.kt)("inlineCode",{parentName:"p"},"cluster.local"),", this means that the primary (formerly master) node of your PostgreSQL cluster is reachable via the DNS entry: ",(0,r.kt)("strong",{parentName:"p"},(0,r.kt)("inlineCode",{parentName:"strong"},"clustered-instance-master.tutorial.svc.cluster.local")),". "),(0,r.kt)("p",null,(0,r.kt)("inlineCode",{parentName:"p"},"username:")),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'echo "YThzLXNiLWN4cDZCMFRUQg==" | base64 --decode\na8s-sb-cxp6B0TTB\n')),(0,r.kt)("p",null,(0,r.kt)("inlineCode",{parentName:"p"},"password:")),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'echo "bk1wNGI2WHdMeXUwYVkzWmF4ekExS1VURTNzM2xham4=" | base64 --decode\nnMp4b6XwLyu0aY3ZaxzA1KUTE3s3lajn\n')),(0,r.kt)("p",null,"As you can see, the secret ",(0,r.kt)("inlineCode",{parentName:"p"},"sb-sample-service-binding")," contains all relevant information required by an application to connect to your PostgreSQL instance."),(0,r.kt)("h2",{id:"step-4-deploying-a-demo-application"},"Step 4: Deploying a Demo Application"),(0,r.kt)("p",null,"With the PostgreSQL database at hand, an exemplary application can be deployed. "),(0,r.kt)("p",null,"The demo app has already been checked out for you. Hence, installing it just a single command away:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl apply -k $(a9s demo pwd)/a8s-demo/demo-app -n tutorial\n")),(0,r.kt)("p",null,"Output:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"service/demo-app created\ndeployment.apps/demo-app created\n")),(0,r.kt)("p",null,"The demo app consists of a Kubernetes Service and a Deployment both named ",(0,r.kt)("inlineCode",{parentName:"p"},"demo-app"),"."),(0,r.kt)("p",null,"You can verify that the app is running by executing:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl get pods -n tutorial -l app=demo-app\n")),(0,r.kt)("p",null,"Output:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"NAME                        READY   STATUS    RESTARTS   AGE\ndemo-app-65f6dd4445-glgc4   1/1     Running   0          81s\n")),(0,r.kt)("p",null,"In order to access the app locally, create a port forward mapping the container port ",(0,r.kt)("inlineCode",{parentName:"p"},"3000")," your local machine's port ",(0,r.kt)("inlineCode",{parentName:"p"},"8080"),":"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"kubectl port-forward service/demo-app -n tutorial 8080:3000\n")),(0,r.kt)("p",null,"Then navigate your browser to: ",(0,r.kt)("a",{parentName:"p",href:"http://localhost:8080"},"http://localhost:8080")),(0,r.kt)("h2",{id:"step-5-interacting-with-postgresql"},"Step 5: Interacting with PostgreSQL"),(0,r.kt)("p",null,"Once you've created a PostgreSQL Service Instance, you can use the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s CLI")," to interact with it."),(0,r.kt)("h3",{id:"applying-a-local-sql-file"},"Applying a Local SQL File"),(0,r.kt)("p",null,"Although not the preferred way to load seed data into a production database, during development it might be handy to execute a SQL file to a PostgreSQL instance. This allows executing one or multiple SQL statements conveniently. "),(0,r.kt)("p",null,"Download an exemplary SQL file:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"wget https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/demo_data.sql\n")),(0,r.kt)("p",null,"Executing an SQL file is as simple as using the ",(0,r.kt)("inlineCode",{parentName:"p"},"--file")," option:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial\n")),(0,r.kt)("p",null,"The ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s CLI")," will determine the replication leader, upload, execute and delete the SQL file. "),(0,r.kt)("p",null,"The ",(0,r.kt)("inlineCode",{parentName:"p"},"--no-delete")," option can be used during debugging of erroneous SQL statements\nas the SQL file remains in the PostgreSQL Leader's Pod."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial --no-delete\n")),(0,r.kt)("p",null,"With the SQL file still available in the Pod, statements can be quickly altered and re-tested."),(0,r.kt)("h3",{id:"applying-an-sql-string"},"Applying an SQL String"),(0,r.kt)("p",null,"It is also possible to execute a SQL string containing one or several SQL statements by using the ",(0,r.kt)("inlineCode",{parentName:"p"},"--sql")," option:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"\n')),(0,r.kt)("p",null,"The output of the command will be printed on the screen, for example:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"Output from the Pod:\n    \ncount \n-------\n    10 \n(1 row)\n")),(0,r.kt)("p",null,"Again, the ",(0,r.kt)("inlineCode",{parentName:"p"},"pg apply")," commands are not meant to interact with production databases but may become handy during debugging and local development."),(0,r.kt)("p",null,"Be aware that these commands are executed by the privileged ",(0,r.kt)("inlineCode",{parentName:"p"},"postgres")," user. Schemas (tables) created by the ",(0,r.kt)("inlineCode",{parentName:"p"},"postgres")," user may not be accessible by roles (users) created in conjunction with Service Bindings. You will then have to grant access privileges to the Service Binding role."),(0,r.kt)("h2",{id:"optional-step-6-creating-and-restoring-a-backup"},"(Optional) Step 6: Creating and Restoring a Backup"),(0,r.kt)("p",null,"Assuming you have configured the backup store and provided access credentials to an AWS S3 compatible object store, try creating and restoring a backup for your application."),(0,r.kt)("h3",{id:"creating-a-backup"},"Creating a Backup"),(0,r.kt)("p",null,"Creating a backup can be achieved with a single command:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg backup --name clustered-backup-1 -i clustered-instance -n tutorial\n")),(0,r.kt)("p",null,"With a closer look at the output you will notice that a backup is also specified by a YAML specification and thus is done in a declarative way. You express that you want a backup to be created:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-YAML"},"apiVersion: backups.anynines.com/v1beta3\nkind: Backup\nmetadata:\n    name: clustered-backup-1\n    namespace: tutorial\nspec:\n    serviceInstance:\n    apiGroup: postgresql.anynines.com\n    kind: Postgresql\n    name: clustered-instance\n")),(0,r.kt)("p",null,"The a8s Backup Manager is the responsible for making the backup happen. It does that by locating the Service Instance ",(0,r.kt)("inlineCode",{parentName:"p"},"clustered-instance")," which also runs the ",(0,r.kt)("inlineCode",{parentName:"p"},"a8s Backup Agent"),". This agent is then executing the PostgreSQL backup command and, depending on its configuration, compressing, encrypting and streaming the backup to the backup object store (S3)."),(0,r.kt)("h3",{id:"restoring-a-backup"},"Restoring a Backup"),(0,r.kt)("p",null,"In order to experience the value of a backup, simulate a data loss by issueing the following ",(0,r.kt)("inlineCode",{parentName:"p"},"DELETE")," statement:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'a9s pg apply -i clustered-instance -n tutorial --sql "DELETE FROM posts"\n')),(0,r.kt)("p",null,"Verify the destructive effect on your data by counting the number of posts:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"\n')),(0,r.kt)("p",null,"And/or reloading the demo-app."),(0,r.kt)("p",null,"Once you've confirmed that all blog posts are gone, it's time to recover the data from the backup."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},"a9s create pg restore --name clustered-restore-1 -b clustered-backup-1 -i clustered-instance -n tutorial\n")),(0,r.kt)("p",null,"Again, apply the ",(0,r.kt)("inlineCode",{parentName:"p"},"COUNT")," or reload the website to see that the restore has brought back all blog posts."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre"},'a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"\n')),(0,r.kt)("p",null,"Some engineers say that a convenient backup/restore functionality at your disposal improves the quality of sleep by 37% \ud83d\ude09."),(0,r.kt)("h2",{id:"congratulations"},"Congratulations"),(0,r.kt)("p",null,"With just a few commands, you have created a local Kubernetes cluster, installed the a8s PostgreSQL Control Plane including all its dependencies. Furthermore, you have provisioned an PostgreSQL cluster consisting of three Pods providing you with an asynchronous streaming cluster supporting automatic failure detection, lead-election and failover. Deploying the demo application you've also experienced the convenience of Service Bindings and their automatic creation of Kubernetes Secrets. The backup and restore experiment then illustrated how effortless handling a production database can be."),(0,r.kt)("p",null,"Did you every think that running a production database as an application developer with full self-service could be so easy?"),(0,r.kt)("h2",{id:"what-to-do-next"},"What to do next?"),(0,r.kt)("p",null,"Wait, there's more to it! This hands-on tutorial merely scratched the surface. Did you see that the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s CLI")," has created many YAML manifests stored in the ",(0,r.kt)("inlineCode",{parentName:"p"},"usermanifests")," folder of your working directory? This is a good place to start tweaking your manifests and start your own experiments."),(0,r.kt)("p",null,"If you want to learn more about a8s PostgreSQL feel free to have a look at the documentation at TODO."),(0,r.kt)("p",null,"For more about the ",(0,r.kt)("inlineCode",{parentName:"p"},"a9s CLI")," have a look at ",(0,r.kt)("a",{parentName:"p",href:"https://github.com/anynines/a9s-cli-v2"},"https://github.com/anynines/a9s-cli-v2"),"."),(0,r.kt)("h2",{id:"links"},"Links"),(0,r.kt)("ol",null,(0,r.kt)("li",{parentName:"ol"},"a9s CLI documentation and source, ",(0,r.kt)("a",{parentName:"li",href:"https://github.com/anynines/a9s-cli-v2"},"https://github.com/anynines/a9s-cli-v2")," "),(0,r.kt)("li",{parentName:"ol"},"PostgreSQL documentation, Log-Shipping Standby Servers, ",(0,r.kt)("a",{parentName:"li",href:"https://www.postgresql.org/docs/current/warm-standby.html"},"https://www.postgresql.org/docs/current/warm-standby.html"))))}d.isMDXComponent=!0}}]);