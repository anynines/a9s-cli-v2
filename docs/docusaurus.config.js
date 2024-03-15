// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

const Versions = require("./versions.json");
const usercentricsScript = ({ NODE_ENV }) => {
  const config = {
    id: "usercentrics-cmp",
    src: "https://app.usercentrics.eu/browser-ui/latest/loader.js",
    "data-settings-id": "Rkzv9fbcQ",
    async: true
  };

  if (NODE_ENV != "production") {
    config["data-version"] = "preview";
  }

  return config;
};

const googleTagScript = () => {
  const config = {
    type: "text/plain",
    src: "https://www.googletagmanager.com/gtag/js?id=GTM-NZZ5ZVC",
    "data-usercentrics": "Google Tag Manager",
    async: true
  };
  return config;
};

const scripts = [usercentricsScript(process.env), googleTagScript()];

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "a9s CLI Documentation",
  tagline: "a9s CLI Documentation",
  url: "https://hub.anynines.com",
  baseUrl: "/",
  trailingSlash: true,
  onBrokenLinks: "warn",
  onBrokenMarkdownLinks: "warn",
  favicon: "/img/favicon.ico",
  organizationName: "anynines", // Usually your GitHub org/user name.
  projectName: "anynines-docs", // Usually your repo name.

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
    localeConfigs: {
      en: {
        label: "English",
        direction: "ltr"
      }
    }
  },

  // See: https://github.com/cmfcmf/docusaurus-search-local
  plugins: [
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      {
        indexBlog: false
      }
    ]
  ],

  scripts: scripts,

  presets: [
    [
      "@docusaurus/preset-classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          versions: {
            current: {
              label: "Develop",
              path: "develop",
              banner: "none"
            }
          },
          sidebarPath: require.resolve("./sidebars.js")
        },
        blog: {
          showReadingTime: true,
          blogTitle: "Changelog",
          path: "changelog",
          blogSidebarTitle: "Versions",
          // Please change this to your repo.
          routeBasePath: "/changelog",
          include: ["*.md", "*.mdx"]
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css")
        }
      })
    ]
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        logo: {
          alt: "anynines Logo",
          src: "img/logo.svg",
          srcDark: "img/logoDark.svg"
        },
        items: [
          {
            type: "doc",
            docId: "a9s-cli",
            position: "left",
            label: "a9s CLI Docs"
          },
          {
            type: "doc",
            docId: "hands-on-tutorials/hands-on-tutorials-index",
            position: "left",
            label: "Tutorials"
          },
          //{ to: "/changelog", label: "Changelog", position: "left" },

          // right
          {
            type: "docsVersionDropdown",
            position: "right",
            dropdownActiveClassDisabled: false // true
          }
        ]
      },
      metadata: [{ name: "docusaurus_tag", content: "default" }], // container tag for local search, keep this to default
      footer: {
        style: "dark",
        links: [
          {
            title: "Documentation",
            items: [
              {
                label: "a9s CLI",
                to: "/docs/develop/a9s-cli"
              },
              {
                label: "a9s Data Services",
                to: "https://docs.anynines.com"
              },
              {
                label: "a9s Data Services for K8s",
                to: "https://docs.k8s.anynines.com/"
              },
              {
                label: "Documentation Tags",
                to: "/docs/tags"
              }
            ]
          },
          {
            title: "Products",
            items: [
              {
                label: "Platform",
                href: "https://paas.anynines.com/"
              },
              {
                label: "Data Services",
                href: "https://www.anynines.com/data-services"
              },
              {
                label: "Enterprise Operation",
                href: "https://www.anynines.com/platform-operations"
              }
            ]
          },
          {
            title: "About",
            items: [
              {
                label: "Blog",
                href: "https://blog.anynines.com/"
              },
              {
                label: "Team",
                href: "https://www.anynines.com/team"
              },
              {
                label: "Career",
                href: "https://www.anynines.com/career"
              },
              {
                label: "Contact",
                href: "https://www.anynines.com/contact"
              }
            ]
          },
          {
            title: "Legal",
            items: [
              {
                label: "Imprint",
                href: "https://www.anynines.com/imprint"
              },
              {
                label: "Privacy Policy",
                href: "https://www.anynines.com/data-privacy"
              }
            ]
          },
          {
            title: "Social Media",
            items: [
              {
                label: "Github",
                href: "https://github.com/anynines"
              },
              {
                label: "Twitter",
                href: "https://twitter.com/anynines?lang=en"
              },
              {
                label: "Facebook",
                href: "https://de-de.facebook.com/anyninescom/"
              },
              {
                label: "Medium",
                href: "https://anynines.medium.com/"
              }
            ]
          }
        ],
        copyright: `Copyright © ${new Date().getFullYear()} anynines`
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme
      }
    })
};

module.exports = config;
