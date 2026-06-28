import type { Meta, StoryObj } from "@storybook/react-vite";
import { asyncNoop, noop, storyQueueStatusForIssue } from "@/stories/fixtures";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { projectFixtures } from "@/mocks/fixtures/projects";
import { AddIssueDialog } from "./index";

const meta = {
  title: "Layout/Dialog/AddIssue",
  component: AddIssueDialog,
  args: {
    dependencyOptions: issueFixtures
      .map((issue) => ({
        ...issue,
        queueStatus: storyQueueStatusForIssue(issue),
        stats: { commentCount: 0 },
      })),
    error: "",
    initialStatus: "ready",
    project: projectFixtures[0],
    projects: projectFixtures,
    onCancel: noop,
    onSubmit: asyncNoop,
  },
} satisfies Meta<typeof AddIssueDialog>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithError: Story = {
  args: {
    error: "Unable to create issue.",
  },
};
