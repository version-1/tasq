import type { Meta, StoryObj } from "@storybook/react-vite";
import { noop } from "@/stories/fixtures";
import { AddProjectDialog } from "./index";

const meta = {
  title: "Layout/Dialog/AddProject",
  component: AddProjectDialog,
  args: {
    onCancel: noop,
  },
} satisfies Meta<typeof AddProjectDialog>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
