import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Versionator',
  tagline: 'Semantic version management made simple',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://benjaminabbitt.github.io',
  baseUrl: '/versionator/',

  organizationName: 'benjaminabbitt',
  projectName: 'versionator',

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/benjaminabbitt/versionator/tree/master/docs/',
          routeBasePath: '/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Versionator',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          href: 'https://github.com/benjaminabbitt/versionator',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Getting Started',
              to: '/',
            },
            {
              label: 'Commands',
              to: '/commands',
            },
            {
              label: 'Templates',
              to: '/templates',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/benjaminabbitt/versionator',
            },
            {
              label: 'Releases',
              href: 'https://github.com/benjaminabbitt/versionator/releases',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Benjamin Abbitt. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'yaml', 'go', 'python', 'ruby', 'rust', 'java', 'kotlin', 'csharp', 'swift', 'php'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
