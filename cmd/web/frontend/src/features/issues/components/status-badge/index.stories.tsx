import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueStatuses } from "@/lib/types";
import { StatusBadge } from "./index";

const meta = {
  title: "Features/Issues/StatusBadge",
  component: StatusBadge,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    status: "ready",
  },
} satisfies Meta<typeof StatusBadge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Ready: Story = {};

export const AllStatuses: Story = {
  render: () => (
    <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
      {issueStatuses.map((status) => (
        <StatusBadge key={status} status={status} />
      ))}
    </div>
  ),
};
