import type { Meta, StoryObj } from "@storybook/react-vite";
import { CommentTypeBadge, commentTypes } from "./index";

const meta = {
  title: "Features/Issues/CommentTypeBadge",
  component: CommentTypeBadge,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    type: "progress",
  },
} satisfies Meta<typeof CommentTypeBadge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Progress: Story = {};

export const AllTypes: Story = {
  render: () => (
    <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
      {commentTypes.map((type) => (
        <CommentTypeBadge key={type} type={type} />
      ))}
    </div>
  ),
};
