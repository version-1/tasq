import type { Meta, StoryObj } from "@storybook/react-vite";
import { asyncNoop, noop } from "@/stories/fixtures";
import { DeleteProjectDialog } from "./index";

const meta = {
  title: "Layout/Dialog/DeleteProject",
  component: DeleteProjectDialog,
  args: {
    error: "",
    isDeleting: false,
    project: {
      createdAt: "2026-06-08T00:00:00Z",
      description: "Product work",
      id: 7,
      key: "product",
      location: "/workspace/product",
      name: "Product Website",
      workflowChecksum: "",
      updatedAt: "2026-06-08T00:00:00Z",
    },
    onCancel: noop,
    onConfirm: asyncNoop,
  },
} satisfies Meta<typeof DeleteProjectDialog>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const RunningRunsError: Story = {
  args: {
    error: "This project still has running runs. Stop them before deleting the project.",
  },
};
