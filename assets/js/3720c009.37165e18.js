"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[3751],{1618:function(e,t,n){n.r(t),n.d(t,{default:function(){return m}});var a=n(7294),l=n(4334),r=n(5565),c=n(5463),s=n(3702),u=n(7919),o=n(6426),i=n(3647);function m(e){let{tags:t}=e;const n=(0,r.M)();return a.createElement(c.FG,{className:(0,l.Z)(s.k.wrapper.docsPages,s.k.page.docsTagsListPage)},a.createElement(c.d,{title:n}),a.createElement(i.Z,{tag:"doc_tags_list"}),a.createElement(u.Z,null,a.createElement("div",{className:"container margin-vert--lg"},a.createElement("div",{className:"row"},a.createElement("main",{className:"col col--8 col--offset-2"},a.createElement("h1",null,n),a.createElement(o.Z,{tags:t}))))))}},3852:function(e,t,n){n.d(t,{Z:function(){return s}});var a=n(7294),l=n(4334),r=n(3699),c={tag:"tag_zVej",tagRegular:"tagRegular_sFm0",tagWithCount:"tagWithCount_h2kH"};function s(e){let{permalink:t,label:n,count:s}=e;return a.createElement(r.Z,{href:t,className:(0,l.Z)(c.tag,s?c.tagWithCount:c.tagRegular)},n,s&&a.createElement("span",null,s))}},6426:function(e,t,n){n.d(t,{Z:function(){return u}});var a=n(7294),l=n(5565),r=n(3852),c={tag:"tag_Nnez"};function s(e){let{letterEntry:t}=e;return a.createElement("article",null,a.createElement("h2",null,t.letter),a.createElement("ul",{className:"padding--none"},t.tags.map((e=>a.createElement("li",{key:e.permalink,className:c.tag},a.createElement(r.Z,e))))),a.createElement("hr",null))}function u(e){let{tags:t}=e;const n=(0,l.P)(t);return a.createElement("section",{className:"margin-vert--lg"},n.map((e=>a.createElement(s,{key:e.letter,letterEntry:e}))))}},5565:function(e,t,n){n.d(t,{M:function(){return l},P:function(){return r}});var a=n(7325);const l=()=>(0,a.I)({id:"theme.tags.tagsPageTitle",message:"Tags",description:"The title of the tag list page"});function r(e){const t={};return Object.values(e).forEach((e=>{const n=function(e){return e[0].toUpperCase()}(e.label);t[n]??=[],t[n].push(e)})),Object.entries(t).sort(((e,t)=>{let[n]=e,[a]=t;return n.localeCompare(a)})).map((e=>{let[t,n]=e;return{letter:t,tags:n.sort(((e,t)=>e.label.localeCompare(t.label)))}}))}}}]);