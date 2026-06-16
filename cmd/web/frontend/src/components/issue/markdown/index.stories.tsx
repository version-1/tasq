import type { Meta, StoryObj } from "@storybook/react-vite";
import { Markdown } from "./index";

const meta = {
  title: "Issue/Markdown",
  component: Markdown,
  decorators: [
    (Story) => (
      <div style={{ maxWidth: 720 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    content: `# Issue description preview

This story shows how issue descriptions render when they include common Markdown patterns. It should be long enough to inspect spacing, line height, wrapping, and dense content inside the same component.

## Implementation notes

Render **Markdown** from issue descriptions while keeping generated links and unsafe markup contained.

- Preserve paragraphs and lists.
- Rewrite attachment URLs such as attachment://demo-image.
- Keep empty descriptions readable.
- Support _emphasis_, **strong text**, and inline \`code\` without changing the surrounding paragraph rhythm.

### Context

The issue detail view often includes planning notes, reviewer instructions, command snippets, and acceptance criteria in one body. Long prose should wrap naturally, keep readable spacing between sections, and avoid pushing table content outside the viewport.

> Reviewer note: use this sample to check how quoted text appears next to headings, lists, tables, and task checkboxes.

#### Steps to verify

1. Open the issue detail story and scan the first viewport.
2. Confirm that headings have a clear hierarchy.
3. Check the table borders and cell padding.
4. Resize the viewport and verify that long lines wrap instead of overlapping.

### Command sample

\`\`\`bash
npm run typecheck
npm test
\`\`\`

### Link and attachment sample

See [Storybook documentation](https://storybook.js.org/) for component review workflows.

![Generated attachment](attachment://demo-image)

### Review matrix

| Area | Owner | Status |
| --- | --- | --- |
| Board rendering | frontend | Ready |
| Detail page | web | In progress |
| Mock data | qa | Review |

### Checklist

- [x] Render paragraphs and lists.
- [x] Keep links sanitized.
- [ ] Confirm task list checkbox styling.
- [ ] Review table spacing on narrow screens.`,
    emptyText: "No description.",
  },
} satisfies Meta<typeof Markdown>;

export default meta;

type Story = StoryObj<typeof meta>;

export const RichText: Story = {};

export const Empty: Story = {
  args: {
    content: "",
  },
};
