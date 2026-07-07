import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'getting-started/overview',
        'getting-started/quickstart',
        'getting-started/setup-guide',
        {
          type: 'category',
          label: 'Concepts',
          items: [
            'getting-started/concepts/overview',
            'getting-started/concepts/orchestrator',
            'getting-started/concepts/issue-tracker',
            'getting-started/concepts/web',
            'getting-started/concepts/tq-cli',
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/codex-autonomy-setup',
        'guides/recover-blocked-session',
        'guides/workflow-configuration',
        'guides/web-ui-operations',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'reference/cli-reference',
        'reference/api',
        'reference/configuration',
        'reference/schema',
      ],
    },
    {
      type: 'category',
      label: 'Contributing',
      items: [
        'contributing/development-setup',
        'contributing/running-locally',
        'contributing/testing',
      ],
    },
  ],
};

export default sidebars;
