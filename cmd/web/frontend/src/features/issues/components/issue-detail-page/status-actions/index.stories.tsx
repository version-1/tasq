import type { Meta, StoryObj } from "@storybook/react-vite";
import { asyncNoop } from "@/stories/fixtures";
import { StatusActions } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/StatusActions",
  component: StatusActions,
  args: {
    currentStatus: "in_progress",
    disabled: false,
    onStatusChange: asyncNoop,
  },
} satisfies Meta<typeof StatusActions>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Enabled: Story = {};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};
