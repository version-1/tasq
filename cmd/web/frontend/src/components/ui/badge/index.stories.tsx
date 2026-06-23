import type { Meta, StoryObj } from "@storybook/react-vite";
import { Badge } from "./index";

const meta = {
  title: "UI/Badge",
  component: Badge,
  args: {
    children: "Tasq",
    variant: "project",
  },
} satisfies Meta<typeof Badge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Project: Story = {};

export const HighPriority: Story = {
  args: {
    children: "High",
    variant: "priority-high",
  },
};
