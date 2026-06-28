import type { Meta, StoryObj } from "@storybook/react-vite";
import { ProjectBadge } from "./index";

const meta = {
  title: "Features/Issues/ProjectBadge",
  component: ProjectBadge,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    projectKey: "TASQ",
  },
} satisfies Meta<typeof ProjectBadge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Small: Story = {
  args: {
    size: "small",
  },
};

export const LongProjectKey: Story = {
  args: {
    projectKey: "PRODUCT-WEBSITE-REDESIGN",
  },
};
