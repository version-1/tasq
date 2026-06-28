import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { issueStatuses, type IssueStatus, type IssueSummary, type QueueStatus } from "@/lib/types";
import { storyQueueStatusForIssue } from "@/stories/fixtures";
import { IssueCard } from "./index";

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

function issueVariant({
  id,
  priority = "normal",
  queueStatus,
  status,
  title,
}: {
  id: number;
  priority?: IssueSummary["priority"];
  queueStatus?: QueueStatus;
  status: IssueStatus;
  title: string;
}): IssueSummary {
  return {
    ...summaryIssue(1, id % 4),
    id,
    priority,
    queueStatus: queueStatus ?? queueStatusForStatus(status),
    status,
    title,
  };
}

function queueStatusForStatus(status: IssueStatus): QueueStatus {
  switch (status) {
    case "backlog":
      return "backlog";
    case "ready":
      return "queued";
    case "in_progress":
      return "processing";
    case "done":
      return "completed";
    default:
      return "inactive";
  }
}

const meta = {
  title: "Features/Issues/Card",
  component: IssueCard,
  decorators: [
    (Story) => (
      <div style={{ maxWidth: 760 }}>
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
      queueStatus: "processing",
    },
  },
};

export const ReviewTransitions: Story = {
  args: {
    issue: {
      ...summaryIssue(3),
      status: "review",
      queueStatus: "inactive",
    },
  },
};

export const ReadyPending: Story = {
  args: {
    issue: {
      ...summaryIssue(1),
      status: "ready",
      queueStatus: "pending",
    },
  },
};

export const AllStatuses: Story = {
  render: (args) => (
    <div style={{ display: "grid", gap: 16, gridTemplateColumns: "repeat(2, minmax(0, 360px))" }}>
      {issueStatuses.map((status, index) => (
        <IssueCard
          key={status}
          {...args}
          issue={issueVariant({
            id: index + 30,
            priority: index % 3 === 0 ? "high" : index % 3 === 1 ? "normal" : "low",
            status,
            title: `${status} issue card`,
          })}
        />
      ))}
    </div>
  ),
};

export const QueueStates: Story = {
  render: (args) => (
    <div style={{ display: "grid", gap: 16, gridTemplateColumns: "repeat(2, minmax(0, 360px))" }}>
      {[
        { queueStatus: "pending", status: "ready", title: "Ready issue blocked by dependency" },
        { queueStatus: "queued", status: "ready", title: "Ready issue queued for work" },
        { queueStatus: "processing", status: "in_progress", title: "In progress issue" },
        { queueStatus: "completed", status: "done", title: "Done issue" },
        { queueStatus: "inactive", status: "review", title: "Review issue" },
        { queueStatus: "backlog", status: "backlog", title: "Backlog issue" },
      ].map((item, index) => (
        <IssueCard
          key={item.queueStatus}
          {...args}
          issue={issueVariant({
            id: index + 60,
            queueStatus: item.queueStatus as QueueStatus,
            status: item.status as IssueStatus,
            title: item.title,
          })}
        />
      ))}
    </div>
  ),
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
