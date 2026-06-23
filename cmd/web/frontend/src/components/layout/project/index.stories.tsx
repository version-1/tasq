import type { Meta, StoryObj } from "@storybook/react-vite";
import { Layout } from "@/components/layout";
import { ProjectLayout } from "./index";

const meta = {
  title: "Layout/ProjectLayout",
  component: ProjectLayout,
  decorators: [
    (Story) => (
      <Layout>
        <Story />
      </Layout>
    ),
  ],
  args: {
    children: <section style={{ padding: 24 }}>Project route content</section>,
    showAddTaskButton: true,
  },
} satisfies Meta<typeof ProjectLayout>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithoutAddTask: Story = {
  args: {
    showAddTaskButton: false,
  },
};
