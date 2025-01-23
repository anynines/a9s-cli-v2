"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[4023],{6714:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>d,contentTitle:()=>t,default:()=>h,frontMatter:()=>r,metadata:()=>c,toc:()=>o});var i=s(4848),l=s(8453);const r={id:"a9s-cli-klutch",title:"a9s CLI Klutch",tags:["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind","klutch"],keywords:["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind","klutch"]},t="klutch Stack",c={id:"a9s-cli-klutch",title:"a9s CLI Klutch",description:"Create a local Klutch Control Plane Cluster using Kind, including the a8s stack. Deploy an App Cluster and bind resources to the Control Plane Cluster.",source:"@site/versioned_docs/version-0.14.1/a9s-cli-klutch.md",sourceDirName:".",slug:"/a9s-cli-klutch",permalink:"/docs/a9s-cli-klutch",draft:!1,unlisted:!1,tags:[{inline:!0,label:"a9s cli",permalink:"/docs/tags/a-9-s-cli"},{inline:!0,label:"a9s hub",permalink:"/docs/tags/a-9-s-hub"},{inline:!0,label:"a9s data services",permalink:"/docs/tags/a-9-s-data-services"},{inline:!0,label:"a8s data services",permalink:"/docs/tags/a-8-s-data-services"},{inline:!0,label:"a9s postgres",permalink:"/docs/tags/a-9-s-postgres"},{inline:!0,label:"a8s postgres",permalink:"/docs/tags/a-8-s-postgres"},{inline:!0,label:"data service",permalink:"/docs/tags/data-service"},{inline:!0,label:"introduction",permalink:"/docs/tags/introduction"},{inline:!0,label:"kubernetes",permalink:"/docs/tags/kubernetes"},{inline:!0,label:"minikube",permalink:"/docs/tags/minikube"},{inline:!0,label:"kind",permalink:"/docs/tags/kind"},{inline:!0,label:"klutch",permalink:"/docs/tags/klutch"}],version:"0.14.1",frontMatter:{id:"a9s-cli-klutch",title:"a9s CLI Klutch",tags:["a9s cli","a9s hub","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","kubernetes","minikube","kind","klutch"],keywords:["a9s cli","a9s hub","a9s platform","a9s data services","a8s data services","a9s postgres","a8s postgres","data service","introduction","postgresql","kubernetes","minikube","kind","klutch"]}},d={},o=[{value:"Prerequisites",id:"prerequisites",level:2},{value:"Installing the <code>kubectl-bind</code> plugin:",id:"installing-the-kubectl-bind-plugin",level:3},{value:"Running on Linux",id:"running-on-linux",level:3},{value:"Commands",id:"commands",level:2},{value:"1. <code>deploy</code>",id:"1-deploy",level:3},{value:"2. <code>bind</code>",id:"2-bind",level:3},{value:"3. <code>delete</code>",id:"3-delete",level:3}];function a(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",ul:"ul",...(0,l.R)(),...e.components};return(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(n.h1,{id:"klutch-stack",children:"klutch Stack"}),"\n",(0,i.jsxs)(n.p,{children:["Create a local Klutch Control Plane Cluster using ",(0,i.jsx)(n.code,{children:"Kind"}),", including the ",(0,i.jsx)(n.code,{children:"a8s"})," stack. Deploy an App Cluster and ",(0,i.jsx)(n.strong,{children:"bind"})," resources to the Control Plane Cluster.\nThis will allow you to use ",(0,i.jsx)(n.code,{children:"a8s"})," resource instances such as ",(0,i.jsx)(n.code,{children:"postgresql"})," on the App Cluster, which will run on the Control Plane Cluster."]}),"\n",(0,i.jsx)(n.h2,{id:"prerequisites",children:"Prerequisites"}),"\n",(0,i.jsxs)(n.ul,{children:["\n",(0,i.jsxs)(n.li,{children:[(0,i.jsx)(n.a,{href:"/docs/a9s-cli#prerequisites",children:"General prerequisites"})," are met."]}),"\n",(0,i.jsxs)(n.li,{children:["Install ",(0,i.jsx)(n.a,{href:"https://helm.sh/docs/intro/install/",children:"Helm"}),"."]}),"\n",(0,i.jsxs)(n.li,{children:["Install ",(0,i.jsx)(n.code,{children:"kubectl-bind"})," plugin version 1.3.0 or higher (see below)."]}),"\n",(0,i.jsxs)(n.li,{children:["On ",(0,i.jsx)(n.strong,{children:"linux"}),", docker must be runnable without sudo. See the ",(0,i.jsx)(n.a,{href:"https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user",children:"docker documentation"})," for further details."]}),"\n"]}),"\n",(0,i.jsxs)(n.h3,{id:"installing-the-kubectl-bind-plugin",children:["Installing the ",(0,i.jsx)(n.code,{children:"kubectl-bind"})," plugin:"]}),"\n",(0,i.jsxs)(n.p,{children:["Download a binary for your platform with the following URL, make it executable and place it in a location in your ",(0,i.jsx)(n.code,{children:"PATH"}),":"]}),"\n",(0,i.jsx)(n.p,{children:(0,i.jsx)(n.code,{children:"https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/v1.3.0/$OS-$ARCH/kubectl-bind"})}),"\n",(0,i.jsxs)(n.p,{children:["Replace ",(0,i.jsx)(n.code,{children:"OS"})," and ",(0,i.jsx)(n.code,{children:"ARCH"})," with values for your platform, e.g. ",(0,i.jsx)(n.code,{children:"darwin-arm64"})," or ",(0,i.jsx)(n.code,{children:"linux-amd64"}),". You can also use the following script to achieve this:"]}),"\n",(0,i.jsx)(n.pre,{children:(0,i.jsx)(n.code,{className:"language-bash",children:'RELEASE="v1.3.0"\nOS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o kubectl-bind https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/$RELEASE/$OS-$ARCH/kubectl-bind\n\nsudo chmod 755 kubectl-bind\nsudo mv kubectl-bind /usr/local/bin\n'})}),"\n",(0,i.jsx)(n.h3,{id:"running-on-linux",children:"Running on Linux"}),"\n",(0,i.jsxs)(n.p,{children:["To avoid issues with ",(0,i.jsx)(n.code,{children:"Kind"})," on Linux, increase the ",(0,i.jsx)(n.code,{children:"inotify"})," resource limits as described ",(0,i.jsx)(n.a,{href:"https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files",children:"here"}),"."]}),"\n",(0,i.jsx)(n.h2,{id:"commands",children:"Commands"}),"\n",(0,i.jsxs)(n.h3,{id:"1-deploy",children:["1. ",(0,i.jsx)(n.code,{children:"deploy"})]}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Usage"}),":"]}),"\n",(0,i.jsx)(n.pre,{children:(0,i.jsx)(n.code,{className:"language-bash",children:"a9s klutch deploy [options]\n"})}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Options"}),":"]}),"\n",(0,i.jsxs)(n.table,{children:[(0,i.jsx)(n.thead,{children:(0,i.jsxs)(n.tr,{children:[(0,i.jsx)(n.th,{children:"Flag"}),(0,i.jsx)(n.th,{children:"Description"}),(0,i.jsx)(n.th,{children:"Example"})]})}),(0,i.jsxs)(n.tbody,{children:[(0,i.jsxs)(n.tr,{children:[(0,i.jsxs)(n.td,{children:[(0,i.jsx)(n.code,{children:"-y"}),", ",(0,i.jsx)(n.code,{children:"--yes"})]}),(0,i.jsx)(n.td,{children:"Skip confirmation prompts"}),(0,i.jsx)(n.td,{children:(0,i.jsx)(n.code,{children:"a9s klutch deploy --yes"})})]}),(0,i.jsxs)(n.tr,{children:[(0,i.jsx)(n.td,{children:(0,i.jsx)(n.code,{children:"--port"})}),(0,i.jsxs)(n.td,{children:["The port to expose the Control Plane Cluster on. Defaults to ",(0,i.jsx)(n.code,{children:"8080"}),"."]}),(0,i.jsx)(n.td,{children:(0,i.jsx)(n.code,{children:"a9s klutch deploy --port 8080"})})]})]})]}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Description"}),":"]}),"\n",(0,i.jsxs)(n.p,{children:["This command deploys a ",(0,i.jsx)(n.code,{children:"Kind"})," cluster named ",(0,i.jsx)(n.code,{children:"klutch-control-plane"})," and installs the required\ncomponents for Klutch. These components include:"]}),"\n",(0,i.jsxs)(n.ul,{children:["\n",(0,i.jsxs)(n.li,{children:["The ",(0,i.jsx)(n.code,{children:"klutch-bind"})," backend and ",(0,i.jsx)(n.a,{href:"https://dexidp.io/",children:"Dex Idp"})," as a dummy OICD provider."]}),"\n",(0,i.jsx)(n.li,{children:"Crossplane and the anynines configuration packages."}),"\n",(0,i.jsxs)(n.li,{children:["The complete ",(0,i.jsx)(n.code,{children:"a8s"})," stack including ",(0,i.jsx)(n.code,{children:"Postgresql"})," operator, backup, restore and service binding capabilities."]}),"\n"]}),"\n",(0,i.jsxs)(n.p,{children:["In addition to the Control Plane Cluster, an App Cluster named ",(0,i.jsx)(n.code,{children:"klutch-app"})," is deployed. This cluster can be used for the ",(0,i.jsx)(n.code,{children:"a9s klutch bind"})," command to bind resources to the Control Plane Cluster."]}),"\n",(0,i.jsx)(n.p,{children:"The Control Plane Cluster exports the following resources for binding:"}),"\n",(0,i.jsxs)(n.ul,{children:["\n",(0,i.jsx)(n.li,{children:(0,i.jsx)(n.code,{children:"postgresqlinstance.anynines.com"})}),"\n",(0,i.jsx)(n.li,{children:(0,i.jsx)(n.code,{children:"servicebinding.anynines.com"})}),"\n",(0,i.jsx)(n.li,{children:(0,i.jsx)(n.code,{children:"backup.anynines.com"})}),"\n",(0,i.jsx)(n.li,{children:(0,i.jsx)(n.code,{children:"restore.anynines.com"})}),"\n"]}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Important"}),": For technical reasons, the Control Plane Cluster is exposed on the local network using the local IP address. If your IP or network changes, the Control Plane Cluster may become unreachable and will have to be redeployed."]}),"\n",(0,i.jsxs)(n.h3,{id:"2-bind",children:["2. ",(0,i.jsx)(n.code,{children:"bind"})]}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Usage"}),":"]}),"\n",(0,i.jsx)(n.pre,{children:(0,i.jsx)(n.code,{children:"a9s klutch bind [options]\n"})}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Options"}),":"]}),"\n",(0,i.jsxs)(n.table,{children:[(0,i.jsx)(n.thead,{children:(0,i.jsxs)(n.tr,{children:[(0,i.jsx)(n.th,{children:"Flag"}),(0,i.jsx)(n.th,{children:"Description"}),(0,i.jsx)(n.th,{children:"Example"})]})}),(0,i.jsx)(n.tbody,{children:(0,i.jsxs)(n.tr,{children:[(0,i.jsxs)(n.td,{children:[(0,i.jsx)(n.code,{children:"-y"}),", ",(0,i.jsx)(n.code,{children:"--yes"})]}),(0,i.jsx)(n.td,{children:"Skip confirmation prompts"}),(0,i.jsx)(n.td,{children:(0,i.jsx)(n.code,{children:"a9s klutch bind --yes"})})]})})]}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Description"}),":"]}),"\n",(0,i.jsxs)(n.p,{children:["This command will invoke ",(0,i.jsx)(n.code,{children:"kubectl bind"})," in order to bind a resource exported by the Control Plane Cluster. This process will open a browser window for you where you can authenticate with the dummy dex OIDC provider using these credentials:"]}),"\n",(0,i.jsxs)(n.p,{children:["Email: ",(0,i.jsx)(n.code,{children:"admin@example.com"})]}),"\n",(0,i.jsxs)(n.p,{children:["Password: ",(0,i.jsx)(n.code,{children:"password"})]}),"\n",(0,i.jsxs)(n.p,{children:["After logging in, grant access, and then ",(0,i.jsx)(n.strong,{children:"choose the resource you would like to bind"}),". Once this is done, return to your terminal and wait for the process to finish."]}),"\n",(0,i.jsxs)(n.p,{children:["After the ",(0,i.jsx)(n.code,{children:"bind"})," command has succeeded, you can deploy instances of the chosen resource on your App Cluster, which will run in the Control Plane Cluster. The command will print an example manifest for the resource you bound that you can apply to the App Cluster with ",(0,i.jsx)(n.code,{children:"kubectl"}),". You can do this easily by copying the printed yaml and using a heredoc, like so:"]}),"\n",(0,i.jsx)(n.pre,{children:(0,i.jsx)(n.code,{className:"language-bash",children:"kubectl apply -f - <<EOF\n<paste your manifests>\nEOF\n"})}),"\n",(0,i.jsxs)(n.h3,{id:"3-delete",children:["3. ",(0,i.jsx)(n.code,{children:"delete"})]}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Usage"}),":"]}),"\n",(0,i.jsx)(n.pre,{children:(0,i.jsx)(n.code,{className:"language-bash",children:"a9s klutch delete [options]\n"})}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Options"}),":"]}),"\n",(0,i.jsxs)(n.table,{children:[(0,i.jsx)(n.thead,{children:(0,i.jsxs)(n.tr,{children:[(0,i.jsx)(n.th,{children:"Flag"}),(0,i.jsx)(n.th,{children:"Description"}),(0,i.jsx)(n.th,{children:"Example"})]})}),(0,i.jsx)(n.tbody,{children:(0,i.jsxs)(n.tr,{children:[(0,i.jsxs)(n.td,{children:[(0,i.jsx)(n.code,{children:"-y"}),", ",(0,i.jsx)(n.code,{children:"--yes"})]}),(0,i.jsx)(n.td,{children:"Skip confirmation prompts"}),(0,i.jsx)(n.td,{children:(0,i.jsx)(n.code,{children:"a9s klutch delete --yes"})})]})})]}),"\n",(0,i.jsxs)(n.p,{children:[(0,i.jsx)(n.strong,{children:"Description"}),":"]}),"\n",(0,i.jsx)(n.p,{children:"This command deletes the Control Plane and App clusters."})]})}function h(e={}){const{wrapper:n}={...(0,l.R)(),...e.components};return n?(0,i.jsx)(n,{...e,children:(0,i.jsx)(a,{...e})}):a(e)}},8453:(e,n,s)=>{s.d(n,{R:()=>t,x:()=>c});var i=s(6540);const l={},r=i.createContext(l);function t(e){const n=i.useContext(r);return i.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function c(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(l):e.components||l:t(e.components),i.createElement(r.Provider,{value:n},e.children)}}}]);