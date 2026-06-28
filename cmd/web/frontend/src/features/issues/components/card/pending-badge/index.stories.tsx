import type { Meta, StoryObj } from "@storybook/react-vite";
import { PendingBadge } from "./index";

const meta = {
  title: "Features/Issues/Card/PendingBadge",
  component: PendingBadge,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof PendingBadge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
