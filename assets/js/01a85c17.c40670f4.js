"use strict";(self.webpackChunkanynines_docs=self.webpackChunkanynines_docs||[]).push([[4013],{2506:function(e,t,a){a.d(t,{Z:function(){return E}});var n=a(7294),l=a(4334),r=a(7919),i=a(3488),s=a(3699),c=a(7325),m={sidebar:"sidebar_re4s",sidebarItemTitle:"sidebarItemTitle_pO2u",sidebarItemList:"sidebarItemList_Yudw",sidebarItem:"sidebarItem__DBe",sidebarItemLink:"sidebarItemLink_mo7H",sidebarItemLinkActive:"sidebarItemLinkActive_I1ZP"};function o(e){let{sidebar:t}=e;return n.createElement("aside",{className:"col col--3"},n.createElement("nav",{className:(0,l.Z)(m.sidebar,"thin-scrollbar"),"aria-label":(0,c.I)({id:"theme.blog.sidebar.navAriaLabel",message:"Blog recent posts navigation",description:"The ARIA label for recent posts in the blog sidebar"})},n.createElement("div",{className:(0,l.Z)(m.sidebarItemTitle,"margin-bottom--md")},t.title),n.createElement("ul",{className:(0,l.Z)(m.sidebarItemList,"clean-list")},t.items.map((e=>n.createElement("li",{key:e.permalink,className:m.sidebarItem},n.createElement(s.Z,{isNavLink:!0,to:e.permalink,className:m.sidebarItemLink,activeClassName:m.sidebarItemLinkActive},e.title)))))))}var u=a(3086);function d(e){let{sidebar:t}=e;return n.createElement("ul",{className:"menu__list"},t.items.map((e=>n.createElement("li",{key:e.permalink,className:"menu__list-item"},n.createElement(s.Z,{isNavLink:!0,to:e.permalink,className:"menu__link",activeClassName:"menu__link--active"},e.title)))))}function g(e){return n.createElement(u.Zo,{component:d,props:e})}function b(e){let{sidebar:t}=e;const a=(0,i.i)();return t?.items.length?"mobile"===a?n.createElement(g,{sidebar:t}):n.createElement(o,{sidebar:t}):null}function E(e){const{sidebar:t,toc:a,children:i,...s}=e,c=t&&t.items.length>0;return n.createElement(r.Z,s,n.createElement("div",{className:"container margin-vert--lg"},n.createElement("div",{className:"row"},n.createElement(b,{sidebar:t}),n.createElement("main",{className:(0,l.Z)("col",{"col--7":c,"col--9 col--offset-1":!c}),itemScope:!0,itemType:"http://schema.org/Blog"},i),a&&n.createElement("div",{className:"col col--2"},a))))}},3977:function(e,t,a){a.r(t),a.d(t,{default:function(){return u}});var n=a(7294),l=a(4334),r=a(5565),i=a(5463),s=a(3702),c=a(2506),m=a(6426),o=a(3647);function u(e){let{tags:t,sidebar:a}=e;const u=(0,r.M)();return n.createElement(i.FG,{className:(0,l.Z)(s.k.wrapper.blogPages,s.k.page.blogTagsListPage)},n.createElement(i.d,{title:u}),n.createElement(o.Z,{tag:"blog_tags_list"}),n.createElement(c.Z,{sidebar:a},n.createElement("h1",null,u),n.createElement(m.Z,{tags:t})))}},3852:function(e,t,a){a.d(t,{Z:function(){return s}});var n=a(7294),l=a(4334),r=a(3699),i={tag:"tag_zVej",tagRegular:"tagRegular_sFm0",tagWithCount:"tagWithCount_h2kH"};function s(e){let{permalink:t,label:a,count:s}=e;return n.createElement(r.Z,{href:t,className:(0,l.Z)(i.tag,s?i.tagWithCount:i.tagRegular)},a,s&&n.createElement("span",null,s))}},6426:function(e,t,a){a.d(t,{Z:function(){return c}});var n=a(7294),l=a(5565),r=a(3852),i={tag:"tag_Nnez"};function s(e){let{letterEntry:t}=e;return n.createElement("article",null,n.createElement("h2",null,t.letter),n.createElement("ul",{className:"padding--none"},t.tags.map((e=>n.createElement("li",{key:e.permalink,className:i.tag},n.createElement(r.Z,e))))),n.createElement("hr",null))}function c(e){let{tags:t}=e;const a=(0,l.P)(t);return n.createElement("section",{className:"margin-vert--lg"},a.map((e=>n.createElement(s,{key:e.letter,letterEntry:e}))))}},5565:function(e,t,a){a.d(t,{M:function(){return l},P:function(){return r}});var n=a(7325);const l=()=>(0,n.I)({id:"theme.tags.tagsPageTitle",message:"Tags",description:"The title of the tag list page"});function r(e){const t={};return Object.values(e).forEach((e=>{const a=function(e){return e[0].toUpperCase()}(e.label);t[a]??=[],t[a].push(e)})),Object.entries(t).sort(((e,t)=>{let[a]=e,[n]=t;return a.localeCompare(n)})).map((e=>{let[t,a]=e;return{letter:t,tags:a.sort(((e,t)=>e.label.localeCompare(t.label)))}}))}}}]);