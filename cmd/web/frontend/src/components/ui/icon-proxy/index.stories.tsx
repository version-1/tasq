import type { Meta, StoryObj } from "@storybook/react-vite";
import { IconProxy } from "./index";

const meta = {
  title: "UI/IconProxy",
  component: IconProxy,
  args: {
    name: "square-kanban",
    size: 24,
  },
} satisfies Meta<typeof IconProxy>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Kanban: Story = {};

export const Settings: Story = {
  args: {
    name: "settings",
  },
};
