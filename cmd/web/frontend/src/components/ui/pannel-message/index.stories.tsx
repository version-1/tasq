import type { Meta, StoryObj } from "@storybook/react-vite";
import { PanelMessage } from "./index";

const meta = {
  title: "UI/PanelMessage",
  component: PanelMessage,
  args: {
    title: "No issues found",
    detail: "Try changing the current project filter.",
  },
} satisfies Meta<typeof PanelMessage>;

export default meta;

type Story = StoryObj<typeof meta>;

export const WithDetail: Story = {};

export const TitleOnly: Story = {
  args: {
    detail: undefined,
  },
};
