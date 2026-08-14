import type { Meta, StoryObj } from "@storybook/react";
import { ResolveIssueDialog } from "./index";

const meta = {
  title: "Features/Issues/ResolveIssueDialog",
  component: ResolveIssueDialog,
  args: {
    issueID: 42,
    issueTitle: "Restore CI access",
    loadLatestBlocker: async () => ({
      id: 18,
      issueId: 42,
      author: "codex",
      type: "blocker",
      body: "Approval is required to update the protected CI configuration. Approve the workflow file change so the agent can restore the failing checks.",
      createdAt: "2026-08-14T03:30:00.000Z",
    }),
    onCancel: () => undefined,
    onMoveIssueReady: async () => undefined,
    onSuccess: () => undefined,
  },
} satisfies Meta<typeof ResolveIssueDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
