import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'getting-started',
    'setup-guide',
    'cli-reference',
    {
      type: 'category',
      label: 'Architecture / Design',
      items: [
        'design/architecture',
        'design/api',
        'design/operations',
        'design/schema',
      ],
    },
  ],
};

export default sidebars;
