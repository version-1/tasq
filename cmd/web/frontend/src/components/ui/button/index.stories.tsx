import type { Meta, StoryObj } from "@storybook/react-vite";
import { noop } from "@/stories/fixtures";
import { Button } from "./index";

const meta = {
  title: "UI/Button",
  component: Button,
  args: {
    children: "Create task",
    onClick: noop,
  },
} satisfies Meta<typeof Button>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Primary: Story = {};
