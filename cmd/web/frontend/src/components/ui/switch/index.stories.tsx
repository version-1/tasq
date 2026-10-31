import type { Meta, StoryObj } from "@storybook/react-vite";
import { Switch } from "./index";

const meta = {
  title: "UI/Switch",
  component: Switch,
  args: {
    "aria-label": "Dark mode",
    checked: true,
  },
} satisfies Meta<typeof Switch>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Checked: Story = {};

export const Unchecked: Story = {
  args: {
    checked: false,
  },
};

export const Decorative: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 16 }}>
      <Switch decorative checked />
      <Switch decorative checked={false} />
      <Switch decorative checked disabled />
    </div>
  ),
};
