import type { Meta, StoryObj } from "@storybook/react-vite";
import type { IssueStatus, Summary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { IssueBoard } from "./index";

const summary: Summary = {
  generatedAt: "2026-06-16T00:00:00.000Z",
  columns: issueStatuses.map((status) => ({
    status,
    title: status,
    issues: issueFixtures.filter((issue) => issue.status === status),
  })),
};

const emptySummary: Summary = {
  generatedAt: "2026-06-16T00:00:00.000Z",
  columns: issueStatuses.map((status) => ({
    status,
    title: status,
    issues: [],
  })),
};

const meta = {
  title: "Issue/Board",
  component: IssueBoard,
  args: {
    summary,
    onAddIssue: (_status?: IssueStatus) => undefined,
    onStatusChange: async () => undefined,
  },
} satisfies Meta<typeof IssueBoard>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};

export const Empty: Story = {
  args: {
    summary: emptySummary,
  },
};
