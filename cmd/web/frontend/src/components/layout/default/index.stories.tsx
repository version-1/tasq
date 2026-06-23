import type { Meta, StoryObj } from "@storybook/react-vite";
import { Layout } from "@/components/layout";
import { DefaultLayout } from "./index";

const meta = {
  title: "Layout/DefaultLayout",
  component: DefaultLayout,
  decorators: [
    (Story) => (
      <Layout>
        <Story />
      </Layout>
    ),
  ],
  args: {
    children: <section style={{ padding: 24 }}>Default route content</section>,
  },
} satisfies Meta<typeof DefaultLayout>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
