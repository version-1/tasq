import type { Meta, StoryObj } from "@storybook/react-vite";
import { ModalOutlet } from "./index";

const meta = {
  title: "UI/ModalOutlet",
  component: ModalOutlet,
  args: {
    children: <div role="dialog">Projected modal content</div>,
  },
} satisfies Meta<typeof ModalOutlet>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
