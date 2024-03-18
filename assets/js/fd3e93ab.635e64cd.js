"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[3098],{4137:function(e,t,n){n.d(t,{Zo:function(){return c},kt:function(){return b}});var r=n(7294);function a(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?o(Object(n),!0).forEach((function(t){a(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function s(e,t){if(null==e)return{};var n,r,a=function(e,t){if(null==e)return{};var n,r,a={},o=Object.keys(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var l=r.createContext({}),u=function(e){var t=r.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},c=function(e){var t=u(e.components);return r.createElement(l.Provider,{value:t},e.children)},p="mdxType",d={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},f=r.forwardRef((function(e,t){var n=e.components,a=e.mdxType,o=e.originalType,l=e.parentName,c=s(e,["components","mdxType","originalType","parentName"]),p=u(n),f=a,b=p["".concat(l,".").concat(f)]||p[f]||d[f]||o;return n?r.createElement(b,i(i({ref:t},c),{},{components:n})):r.createElement(b,i({ref:t},c))}));function b(e,t){var n=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var o=n.length,i=new Array(o);i[0]=f;var s={};for(var l in t)hasOwnProperty.call(t,l)&&(s[l]=t[l]);s.originalType=e,s[p]="string"==typeof e?e:a,i[1]=s;for(var u=2;u<o;u++)i[u]=n[u];return r.createElement.apply(null,i)}return r.createElement.apply(null,n)}f.displayName="MDXCreateElement"},2462:function(e,t,n){n.r(t),n.d(t,{assets:function(){return l},contentTitle:function(){return i},default:function(){return d},frontMatter:function(){return o},metadata:function(){return s},toc:function(){return u}});var r=n(3117),a=(n(7294),n(4137));const o={id:"hands-on-tutorials-index",title:"Hands-On Tutorials",tags:["a9s CLI","tutorials","a9s Hub"],keywords:["a9s CLI","tutorials","a9s Hub"]},i="Hands-On-Tutorials",s={unversionedId:"hands-on-tutorials/hands-on-tutorials-index",id:"hands-on-tutorials/hands-on-tutorials-index",title:"Hands-On Tutorials",description:"The hands-on tutorials guide you through practical experiments using the a9s CLI to learn about Kubernetes, data services and application development.",source:"@site/docs/hands-on-tutorials/index.md",sourceDirName:"hands-on-tutorials",slug:"/hands-on-tutorials/",permalink:"/docs/develop/hands-on-tutorials/",draft:!1,tags:[{label:"a9s CLI",permalink:"/docs/develop/tags/a-9-s-cli"},{label:"tutorials",permalink:"/docs/develop/tags/tutorials"},{label:"a9s Hub",permalink:"/docs/develop/tags/a-9-s-hub"}],version:"current",frontMatter:{id:"hands-on-tutorials-index",title:"Hands-On Tutorials",tags:["a9s CLI","tutorials","a9s Hub"],keywords:["a9s CLI","tutorials","a9s Hub"]},sidebar:"tutorialSidebar",previous:{title:"a9s CLI",permalink:"/docs/develop/a9s-cli"},next:{title:"Deploying a Demo App using a8s PostgreSQL",permalink:"/docs/develop/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli"}},l={},u=[{value:"Deploying an application with PostgreSQL to a local Kubernetes cluster.",id:"deploying-an-application-with-postgresql-to-a-local-kubernetes-cluster",level:2}],c={toc:u},p="wrapper";function d(e){let{components:t,...n}=e;return(0,a.kt)(p,(0,r.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"hands-on-tutorials"},"Hands-On-Tutorials"),(0,a.kt)("p",null,"The hands-on tutorials guide you through practical experiments using the ",(0,a.kt)("inlineCode",{parentName:"p"},"a9s")," CLI to learn about Kubernetes, data services and application development."),(0,a.kt)("h2",{id:"deploying-an-application-with-postgresql-to-a-local-kubernetes-cluster"},"Deploying an application with PostgreSQL to a local Kubernetes cluster."),(0,a.kt)("p",null,"In this tutorial you will use the ",(0,a.kt)("inlineCode",{parentName:"p"},"a9s")," CLI to provision a local Kubernetes cluster using Kind or Minikube, install a PostgreSQL operator, deploy a demo application, load data into the database as well as perform backup and restore operations."),(0,a.kt)("p",null,(0,a.kt)("a",{parentName:"p",href:"/docs/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli/"},"Go to the PostgreSQL Tutorial")))}d.isMDXComponent=!0}}]);