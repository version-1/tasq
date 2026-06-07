import type { Meta, StoryObj } from "@storybook/react-vite";
import { Header } from "./index";

const meta = {
  title: "Layout/Header",
  component: Header,
  args: {
    activePage: "issues",
    projectName: "Product Website",
    issueCount: 24,
    showViewNavigation: true,
    onAddTask: () => undefined,
  },
} satisfies Meta<typeof Header>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Issues: Story = {};

export const Loading: Story = {
  args: {
    issueCount: null,
  },
};
