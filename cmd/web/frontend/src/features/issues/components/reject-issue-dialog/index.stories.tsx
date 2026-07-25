import type { Meta, StoryObj } from "@storybook/react";
import { RejectIssueDialog } from "./index";

const meta = {
  title: "Features/Issues/RejectIssueDialog",
  component: RejectIssueDialog,
  args: {
    issueID: 42,
    issueTitle: "Refine issue detail",
    onCancel: () => undefined,
    onMoveIssueReady: async () => undefined,
    onSuccess: () => undefined,
  },
} satisfies Meta<typeof RejectIssueDialog>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
