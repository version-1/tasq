import type { Meta, StoryObj } from "@storybook/react-vite";
import { asyncNoop, noop } from "@/stories/fixtures";
import { projectFixtures } from "@/mocks/fixtures/projects";
import { AddIssueDialog } from "./index";

const meta = {
  title: "Layout/AddIssueDialog",
  component: AddIssueDialog,
  args: {
    error: "",
    initialStatus: "ready",
    project: projectFixtures[0],
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
