import type { CSSProperties } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { noop } from "@/stories/fixtures";
import { Button } from "./index";

const meta = {
  title: "UI/Button",
  component: Button,
  args: {
    children: "Create task",
    onClick: noop,
    size: "default",
    variant: "primary",
  },
  argTypes: {
    size: { control: "inline-radio", options: ["default", "compact"] },
    variant: {
      control: "inline-radio",
      options: ["primary", "positive", "secondary", "tertiary"],
    },
  },
} satisfies Meta<typeof Button>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Primary: Story = {};

export const Positive: Story = {
  args: { children: "Ready", variant: "positive" },
};

export const Secondary: Story = {
  args: { children: "Resolve", variant: "secondary" },
};

export const Tertiary: Story = {
  args: { children: "Reject", variant: "tertiary" },
};

export const Compact: Story = {
  args: { children: "Done", size: "compact" },
};

export const Variants: Story = {
  render: (args) => (
    <div style={variantGridStyle}>
      <Button {...args} variant="primary">
        Primary
      </Button>
      <Button {...args} variant="positive">
        Positive
      </Button>
      <Button {...args} variant="secondary">
        Secondary
      </Button>
      <Button {...args} variant="tertiary">
        Tertiary
      </Button>
    </div>
  ),
};

const variantGridStyle: CSSProperties = {
  alignItems: "center",
  display: "flex",
  flexWrap: "wrap",
  gap: "12px",
};
