import type { Meta, StoryObj } from "@storybook/react-vite";
import { Header } from "./index";

const meta = {
  title: "Layout/Header",
  component: Header,
  args: {
    activePage: "issues",
    projectName: "Product Website",
    issueCount: 24,
    pages: [
      { key: "issues", href: "/issues", titleKey: "header.board" },
      { key: "settings", href: "/settings", titleKey: "header.settings" },
    ],
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
