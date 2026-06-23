import type { Meta, StoryObj } from "@storybook/react-vite";
import { Layout } from "@/components/layout";
import { IssueDetailLayout } from "./index";

const meta = {
  title: "Layout/IssueDetailLayout",
  component: IssueDetailLayout,
  decorators: [
    (Story) => (
      <Layout>
        <Story />
      </Layout>
    ),
  ],
  args: {
    children: <section style={{ padding: 24 }}>Issue detail content</section>,
  },
} satisfies Meta<typeof IssueDetailLayout>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
