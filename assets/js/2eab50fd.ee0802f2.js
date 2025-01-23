"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[180],{3239:(e,s,n)=>{n.r(s),n.d(s,{assets:()=>r,contentTitle:()=>l,default:()=>h,frontMatter:()=>t,metadata:()=>c,toc:()=>o});var a=n(4848),i=n(8453);const t={id:"a9s-cli",title:"a9s CLI",tags:["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind","klutch"],keywords:["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind","klutch"]},l="a9s CLI",c={id:"a9s-cli",title:"a9s CLI",description:"anynines provides a command line tool called a9s to facilitate application development, devops tasks and interact with selected anynines products.",source:"@site/versioned_docs/version-0.14.1/a9s-cli-index.md",sourceDirName:".",slug:"/a9s-cli",permalink:"/docs/a9s-cli",draft:!1,unlisted:!1,tags:[{inline:!0,label:"a9s cli",permalink:"/docs/tags/a-9-s-cli"},{inline:!0,label:"a9s hub",permalink:"/docs/tags/a-9-s-hub"},{inline:!0,label:"a9s data services",permalink:"/docs/tags/a-9-s-data-services"},{inline:!0,label:"a8s data services",permalink:"/docs/tags/a-8-s-data-services"},{inline:!0,label:"a9s postgres",permalink:"/docs/tags/a-9-s-postgres"},{inline:!0,label:"a8s postgres",permalink:"/docs/tags/a-8-s-postgres"},{inline:!0,label:"data service",permalink:"/docs/tags/data-service"},{inline:!0,label:"introduction",permalink:"/docs/tags/introduction"},{inline:!0,label:"kubernetes",permalink:"/docs/tags/kubernetes"},{inline:!0,label:"minikube",permalink:"/docs/tags/minikube"},{inline:!0,label:"kind",permalink:"/docs/tags/kind"},{inline:!0,label:"klutch",permalink:"/docs/tags/klutch"}],version:"0.14.1",frontMatter:{id:"a9s-cli",title:"a9s CLI",tags:["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind","klutch"],keywords:["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind","klutch"]}},r={},o=[{value:"Prerequisites",id:"prerequisites",level:2},{value:"Installing the CLI",id:"installing-the-cli",level:2},{value:"Using the CLI",id:"using-the-cli",level:2},{value:"Use Cases",id:"use-cases",level:2},{value:"<code>a8s</code> Stack",id:"a8s-stack",level:3},{value:"Go to the a8s Stack documentation",id:"go-to-the-a8s-stack-documentation",level:3},{value:"<code>klutch</code> Stack",id:"klutch-stack",level:3},{value:"Go to the klutch Stack documentation",id:"go-to-the-klutch-stack-documentation",level:3}];function d(e){const s={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",ul:"ul",...(0,i.R)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsx)(s.h1,{id:"a9s-cli",children:"a9s CLI"}),"\n",(0,a.jsxs)(s.p,{children:["anynines provides a command line tool called ",(0,a.jsx)(s.code,{children:"a9s"})," to facilitate application development, devops tasks and interact with selected anynines products."]}),"\n",(0,a.jsx)(s.h2,{id:"prerequisites",children:"Prerequisites"}),"\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsx)(s.li,{children:"MacOS / Linux."}),"\n",(0,a.jsx)(s.li,{children:"Using the backup/restore feature of a8s PostgreSQL requires an S3 compatible endpoint."}),"\n",(0,a.jsxs)(s.li,{children:["Install Go (if you want ",(0,a.jsx)(s.code,{children:"go env"})," to identify your OS and arch)."]}),"\n",(0,a.jsx)(s.li,{children:"Install Git."}),"\n",(0,a.jsx)(s.li,{children:"Install Docker."}),"\n",(0,a.jsx)(s.li,{children:"Install Kubectl."}),"\n",(0,a.jsx)(s.li,{children:"Install Kind and/or Minikube."}),"\n"]}),"\n",(0,a.jsx)(s.h2,{id:"installing-the-cli",children:"Installing the CLI"}),"\n",(0,a.jsxs)(s.p,{children:["In order to install the ",(0,a.jsx)(s.code,{children:"a9s"})," CLI execute the following shell script:"]}),"\n",(0,a.jsx)(s.pre,{children:(0,a.jsx)(s.code,{className:"language-bash",children:"OS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o a9s https://github.com/anynines/a9s-cli-v2/releases/download/v0.14.1/a9s-cli-v2_${OS}_${ARCH}.tar.gz\n    \nsudo chmod 755 a9s\nsudo mv a9s /usr/local/bin\n"})}),"\n",(0,a.jsxs)(s.p,{children:["This will download the ",(0,a.jsx)(s.code,{children:"a9s"})," binary suitable for your architecture and move it to ",(0,a.jsx)(s.code,{children:"/usr/local/bin"}),".\nDepending on your system you have to adjust the ",(0,a.jsx)(s.code,{children:"PATH"})," variable or move the binary to a folder that's already in the ",(0,a.jsx)(s.code,{children:"PATH"}),"."]}),"\n",(0,a.jsx)(s.h2,{id:"using-the-cli",children:"Using the CLI"}),"\n",(0,a.jsx)(s.pre,{children:(0,a.jsx)(s.code,{className:"language-bash",children:"a9s\n"})}),"\n",(0,a.jsx)(s.h2,{id:"use-cases",children:"Use Cases"}),"\n",(0,a.jsxs)(s.p,{children:["The ",(0,a.jsx)(s.code,{children:"a9s"})," CLI can be used to install and use the following stacks:"]}),"\n",(0,a.jsxs)(s.h3,{id:"a8s-stack",children:[(0,a.jsx)(s.code,{children:"a8s"})," Stack"]}),"\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsxs)(s.li,{children:["Install a local Kubernetes cluster (",(0,a.jsx)(s.code,{children:"minikube"})," or ",(0,a.jsx)(s.code,{children:"kind"}),")."]}),"\n",(0,a.jsxs)(s.li,{children:["Install the ",(0,a.jsx)(s.a,{href:"https://cert-manager.io/",children:"cert-manager"}),"."]}),"\n",(0,a.jsx)(s.li,{children:"Install a local Minio object store for storing Backups."}),"\n",(0,a.jsxs)(s.li,{children:["Install the a8s PostgreSQL Operator PostgreSQL supporting","\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsxs)(s.li,{children:["creating dedicated PostgreSQL clusters with","\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsx)(s.li,{children:"synchronous and asynchronous streaming replication."}),"\n",(0,a.jsx)(s.li,{children:"automatic failure detection and automatic failover."}),"\n"]}),"\n"]}),"\n",(0,a.jsx)(s.li,{children:"backup and restore capabilities storing backups in an S3 compatible object store such as AWS S3 or Minio."}),"\n",(0,a.jsx)(s.li,{children:"ability to easily create database users and Kubernetes Secrets by using the Service Bindings abstraction"}),"\n"]}),"\n"]}),"\n",(0,a.jsxs)(s.li,{children:["Easily apply ",(0,a.jsx)(s.code,{children:".sql"})," files and SQL commands to PostgreSQL clusters."]}),"\n"]}),"\n",(0,a.jsx)(s.h3,{id:"go-to-the-a8s-stack-documentation",children:(0,a.jsx)(s.a,{href:"/docs/a9s-cli-a8s/",children:"Go to the a8s Stack documentation"})}),"\n",(0,a.jsxs)(s.h3,{id:"klutch-stack",children:[(0,a.jsx)(s.code,{children:"klutch"})," Stack"]}),"\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsxs)(s.li,{children:["Install a local Klutch Control Plane Cluster using ",(0,a.jsx)(s.code,{children:"kind"})]}),"\n",(0,a.jsx)(s.li,{children:"Install Crossplane and the a8s stack on the Control Plane Cluster"}),"\n",(0,a.jsx)(s.li,{children:"Bind resources from an App Cluster to the Control Plane Cluster"}),"\n"]}),"\n",(0,a.jsx)(s.h3,{id:"go-to-the-klutch-stack-documentation",children:(0,a.jsx)(s.a,{href:"/docs/a9s-cli-klutch/",children:"Go to the klutch Stack documentation"})})]})}function h(e={}){const{wrapper:s}={...(0,i.R)(),...e.components};return s?(0,a.jsx)(s,{...e,children:(0,a.jsx)(d,{...e})}):d(e)}},8453:(e,s,n)=>{n.d(s,{R:()=>l,x:()=>c});var a=n(6540);const i={},t=a.createContext(i);function l(e){const s=a.useContext(t);return a.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function c(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(i):e.components||i:l(e.components),a.createElement(t.Provider,{value:s},e.children)}}}]);