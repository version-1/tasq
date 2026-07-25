import type { Meta, StoryObj } from "@storybook/react-vite";
import { ChangeRequestCard } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/ChangeRequestCard",
  component: ChangeRequestCard,
  args: {
    changeRequest: {
      id: 3,
      issueId: 42,
      author: "qa",
      body: "Add a regression test for retrying after the ready update fails.",
      status: "resolved",
      createdAt: "2026-06-21T03:30:00.000Z",
      updatedAt: "2026-06-21T04:15:00.000Z",
      resolvedAt: "2026-06-21T04:15:00.000Z",
      resolvedByRunId: "run-review-fix-3",
      resultCommentId: 17,
    },
  },
} satisfies Meta<typeof ChangeRequestCard>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
