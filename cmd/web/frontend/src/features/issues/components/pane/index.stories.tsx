import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import type { IssueSummary } from "@/lib/types";
import { storyQueueStatusForIssue } from "@/stories/fixtures";
import { IssueDetail } from "./index";

function summaryIssue(index: number, commentCount = 0): IssueSummary {
  const issue = issueFixtures[index];
  return {
    ...issue,
    queueStatus: storyQueueStatusForIssue(issue),
    stats: {
      commentCount,
    },
  };
}

const meta = {
  title: "Features/Issues/Pane",
  component: IssueDetail,
  decorators: [
    (Story) => (
      <div style={{ minHeight: 520, position: "relative" }}>
        <Story />
      </div>
    ),
  ],
  args: {
    issue: summaryIssue(0, 2),
    onStatusChange: async () => undefined,
  },
} satisfies Meta<typeof IssueDetail>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Unassigned: Story = {
  args: {
    issue: summaryIssue(3),
  },
};

export const EmptyDescription: Story = {
  args: {
    issue: {
      ...summaryIssue(0),
      description: "",
    },
  },
};
