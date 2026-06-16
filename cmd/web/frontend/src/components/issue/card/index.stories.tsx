import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { IssueCard } from "./index";

const meta = {
  title: "Issue/Card",
  component: IssueCard,
  decorators: [
    (Story) => (
      <div style={{ maxWidth: 360 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    issue: issueFixtures[1],
    commentCount: 3,
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
      ...issueFixtures[2],
      status: "in_progress",
    },
  },
};

export const ReviewTransitions: Story = {
  args: {
    issue: {
      ...issueFixtures[3],
      status: "review",
    },
  },
};

export const BacklogQuickAction: Story = {
  args: {
    issue: {
      ...issueFixtures[1],
      status: "backlog",
    },
  },
};

export const WithoutMetrics: Story = {
  args: {
    commentCount: undefined,
    runCount: undefined,
  },
};

export const Urgent: Story = {
  args: {
    issue: {
      ...issueFixtures[1],
      id: 9,
      title: "Restore failed workspace orchestration",
      status: "backlog",
      priority: "urgent",
      assignee: "ops",
    },
  },
};
