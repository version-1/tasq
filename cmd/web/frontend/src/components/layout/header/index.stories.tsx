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
          { key: "issues", href: "/issues", titleKey: "header.board" },
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
