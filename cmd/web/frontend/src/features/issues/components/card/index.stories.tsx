import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import type { IssueSummary } from "@/lib/types";
import { IssueCard } from "./index";

function summaryIssue(index: number, commentCount = 0): IssueSummary {
  return {
    ...issueFixtures[index],
    stats: {
      commentCount,
    },
  };
}

const meta = {
  title: "Features/Issues/Card",
  component: IssueCard,
  decorators: [
    (Story) => (
      <div style={{ maxWidth: 360 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    issue: summaryIssue(1, 3),
    onStatusChange: async () => undefined,
    runCount: 2,
  },
} satisfies Meta<typeof IssueCard>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Readonly: Story = {
  args: {
    readonly: true,
  },
};

export const InProgressLocked: Story = {
  args: {
    issue: {
      ...summaryIssue(2),
      status: "in_progress",
    },
  },
};

export const ReviewTransitions: Story = {
  args: {
    issue: {
      ...summaryIssue(3),
      status: "review",
    },
  },
};

export const BacklogQuickAction: Story = {
  args: {
    issue: {
      ...summaryIssue(1),
      status: "backlog",
    },
  },
};

export const WithoutRunCount: Story = {
  args: {
    runCount: undefined,
  },
};

export const Urgent: Story = {
  args: {
    issue: {
      ...summaryIssue(1),
      id: 9,
      title: "Restore failed workspace orchestration",
      status: "backlog",
      priority: "urgent",
      assignee: "ops",
    },
  },
};
