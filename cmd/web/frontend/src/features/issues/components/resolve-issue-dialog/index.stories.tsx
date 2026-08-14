import type { Meta, StoryObj } from "@storybook/react";
import { ResolveIssueDialog } from "./index";

const meta = {
  title: "Features/Issues/ResolveIssueDialog",
  component: ResolveIssueDialog,
  args: {
    issueID: 42,
    issueTitle: "Restore CI access",
    onCancel: () => undefined,
    onMoveIssueReady: async () => undefined,
    onSuccess: () => undefined,
  },
} satisfies Meta<typeof ResolveIssueDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
