import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'competitors',
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
        'commands/init',
        'commands/major',
        'commands/minor',
        'commands/patch',
        'commands/prefix',
        'commands/prerelease',
        'commands/metadata',
        'commands/release',
        'commands/emit',
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
        'integration/binary-embedding',
        {
          type: 'category',
          label: 'Languages',
          link: {
            type: 'doc',
            id: 'integration/languages/index',
          },
          items: [
            'integration/languages/go',
            'integration/languages/rust',
            'integration/languages/c',
            'integration/languages/cpp',
            'integration/languages/java',
            'integration/languages/kotlin',
            'integration/languages/csharp',
            'integration/languages/swift',
            'integration/languages/python',
            'integration/languages/javascript',
            'integration/languages/typescript',
            'integration/languages/ruby',
            'integration/languages/docker',
          ],
        },
        'integration/git',
        {
          type: 'category',
          label: 'CI/CD',
          link: {
            type: 'doc',
            id: 'integration/cicd/index',
          },
          items: [
            'integration/cicd/github-actions',
            'integration/cicd/gitlab-ci',
            'integration/cicd/azure-devops',
            'integration/cicd/jenkins',
            'integration/cicd/circleci',
          ],
        },
        'integration/makefiles',
      ],
    },
  ],
};

export default sidebars;
