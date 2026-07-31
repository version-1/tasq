import type { Meta, StoryObj } from "@storybook/react";
import { ChangeRequestDialog } from "./index";

const meta = {
  title: "Features/Issues/ChangeRequestDialog",
  component: ChangeRequestDialog,
  args: {
    issueID: 42,
    issueTitle: "Refine issue detail",
    onCancel: () => undefined,
    onMoveIssueReady: async () => undefined,
    onSuccess: () => undefined,
    variant: "reject",
  },
} satisfies Meta<typeof ChangeRequestDialog>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ContinueWithComment: Story = {
  args: {
    variant: "continue",
  },
};
