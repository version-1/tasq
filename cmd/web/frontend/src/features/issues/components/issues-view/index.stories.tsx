import type { Meta, StoryObj } from "@storybook/react-vite";
import { asyncNoop, noop, storySummary } from "@/stories/fixtures";
import { IssuesView } from "./index";

const meta = {
  title: "Features/Issues/IssuesView",
  component: IssuesView,
  args: {
    summary: storySummary,
    onAddIssue: noop,
    onStatusChange: asyncNoop,
  },
} satisfies Meta<typeof IssuesView>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
