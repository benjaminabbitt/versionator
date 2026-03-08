import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'installation',
    'quick-start',
    {
      type: 'category',
      label: 'Concepts',
      items: [
        'concepts/version-file',
        'concepts/semver',
        'concepts/monorepo',
      ],
    },
    {
      type: 'category',
      label: 'Commands',
      link: {
        type: 'doc',
        id: 'commands/index',
      },
      items: [
        'commands/version',
        'commands/major',
        'commands/minor',
        'commands/patch',
        'commands/prefix',
        'commands/prerelease',
        'commands/metadata',
        'commands/tag',
        'commands/emit',
        'commands/config',
        'commands/custom',
        'commands/vars',
      ],
    },
    {
      type: 'category',
      label: 'Configuration',
      items: [
        'configuration/config-file',
        'configuration/shell-completion',
      ],
    },
    {
      type: 'category',
      label: 'Templates',
      link: {
        type: 'doc',
        id: 'templates/index',
      },
      items: [
        'templates/variables',
        'templates/prerelease',
        'templates/metadata',
      ],
    },
    {
      type: 'category',
      label: 'Integration',
      items: [
        'integration/git',
        'integration/cicd',
        'integration/makefiles',
        'integration/languages',
      ],
    },
    {
      type: 'category',
      label: 'Examples',
      items: [
        'examples/python',
        'examples/go',
        'examples/javascript',
        'examples/rust',
      ],
    },
  ],
};

export default sidebars;
