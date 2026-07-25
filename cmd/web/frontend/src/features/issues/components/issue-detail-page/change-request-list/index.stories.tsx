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
      {
        id: 2,
        issueId: 42,
        author: "reviewer",
        body: "Update the screenshots after the modal copy changes.",
        status: "in_progress",
        createdAt: "2026-06-21T03:00:00.000Z",
        updatedAt: "2026-06-21T03:10:00.000Z",
        resolvedAt: null,
        resolvedByRunId: null,
        resultCommentId: null,
      },
      {
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
      {
        id: 4,
        issueId: 42,
        author: "reviewer",
        body: "This request was superseded by the newer retry requirement.",
        status: "canceled",
        createdAt: "2026-06-21T04:30:00.000Z",
        updatedAt: "2026-06-21T04:45:00.000Z",
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
