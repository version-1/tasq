import type { Meta, StoryObj } from "@storybook/react-vite";
import { RejectAction } from ".";

const meta = {
  title: "Features/Issues/RejectAction",
  component: RejectAction,
  args: {
    onOpenDialog: () => undefined,
    onSelectShortcut: async () => undefined,
  },
} satisfies Meta<typeof RejectAction>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
