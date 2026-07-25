import type { Meta, StoryObj } from "@storybook/react";
import { ChangeRequestList } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/ChangeRequestList",
  component: ChangeRequestList,
  args: {
    changeRequests: [
      {
        id: 1,
        issueId: 42,
        author: "reviewer",
        body: "Please cover the empty state before marking this done.",
        status: "open",
        createdAt: "2026-06-21T02:30:00.000Z",
        updatedAt: "2026-06-21T02:30:00.000Z",
        resolvedAt: null,
        resolvedByRunId: null,
        resultCommentId: null,
      },
    ],
    error: "",
    isLoading: false,
  },
} satisfies Meta<typeof ChangeRequestList>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
