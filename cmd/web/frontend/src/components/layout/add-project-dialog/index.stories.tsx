import type { Meta, StoryObj } from "@storybook/react-vite";
import { asyncNoop, noop } from "@/stories/fixtures";
import { AddProjectDialog } from "./index";

const meta = {
  title: "Layout/AddProjectDialog",
  component: AddProjectDialog,
  args: {
    error: "",
    onCancel: noop,
    onSubmit: asyncNoop,
  },
} satisfies Meta<typeof AddProjectDialog>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithError: Story = {
  args: {
    error: "Location must be an absolute path.",
  },
};
