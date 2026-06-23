import type { Meta, StoryObj } from "@storybook/react-vite";
import { storyShellData } from "@/stories/fixtures";
import { Layout, ShellLayout } from "./index";

const meta = {
  title: "Layout/AppLayout",
  component: Layout,
} satisfies Meta<typeof Layout>;

export default meta;

type Story = StoryObj<typeof meta>;

export const ShellReady: Story = {
  args: {
    children: <section style={{ padding: 24 }}>Shell story content</section>,
  },
  render: () => (
    <ShellLayout shellData={storyShellData}>
      <section style={{ padding: 24 }}>
        <h2>Story content</h2>
        <p>Route content rendered inside the application shell.</p>
      </section>
    </ShellLayout>
  ),
};

export const ProviderLayout: Story = {
  args: {
    children: <section style={{ padding: 24 }}>Layout provider story</section>,
  },
};
