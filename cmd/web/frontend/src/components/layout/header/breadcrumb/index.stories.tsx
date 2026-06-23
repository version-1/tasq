import type { Meta, StoryObj } from "@storybook/react-vite";
import { Breadcrumb } from "./index";

const meta = {
  title: "Layout/Breadcrumb",
  component: Breadcrumb,
  parameters: {
    reactRouter: {
      routePath: "/issues/4",
    },
  },
} satisfies Meta<typeof Breadcrumb>;

export default meta;

type Story = StoryObj<typeof meta>;

export const IssueDetail: Story = {};
