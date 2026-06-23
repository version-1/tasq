import type { Meta, StoryObj } from "@storybook/react-vite";
import { MetaItem } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/MetaItem",
  component: MetaItem,
  args: {
    label: "Assignee",
    value: "frontend",
  },
} satisfies Meta<typeof MetaItem>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
