import type { Meta, StoryObj } from "@storybook/react-vite";
import { ChangeRequestStatusBadge, changeRequestStatuses } from "./index";

const meta = {
  title: "Features/Issues/ChangeRequestStatusBadge",
  component: ChangeRequestStatusBadge,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    status: "open",
  },
} satisfies Meta<typeof ChangeRequestStatusBadge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Open: Story = {};

export const AllStatuses: Story = {
  render: () => (
    <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
      {changeRequestStatuses.map((status) => (
        <ChangeRequestStatusBadge key={status} status={status} />
      ))}
    </div>
  ),
};
