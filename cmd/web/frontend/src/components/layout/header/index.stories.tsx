import type { Meta, StoryObj } from "@storybook/react-vite";
import { TabsProvider } from "@/context/tabs";
import { Header } from "./index";

const meta = {
  title: "Layout/Header",
  component: Header,
  decorators: [
    (Story) => (
      <TabsProvider
        activeKey="issues"
        pages={[
          { key: "issues", href: "/dashboard", titleKey: "header.board" },
          { key: "settings", href: "/settings", titleKey: "header.settings" },
        ]}
      >
        <Story />
      </TabsProvider>
    ),
  ],
  args: {
    activePage: "issues",
    projectName: "Product Website",
    issueCount: 24,
    language: "ja",
    showViewNavigation: true,
    onAddTask: () => undefined,
    onLanguageChange: () => undefined,
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
