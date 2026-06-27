import type { Meta, StoryObj } from "@storybook/react-vite";
import { PriorityBadge } from "./index";

const meta = {
  title: "Features/Issues/PriorityBadge",
  component: PriorityBadge,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    priority: "normal",
  },
} satisfies Meta<typeof PriorityBadge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Normal: Story = {};

export const AllPriorities: Story = {
  render: () => (
    <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
      {(["urgent", "high", "normal", "low"] as const).map((priority) => (
        <PriorityBadge key={priority} priority={priority} />
      ))}
    </div>
  ),
};
