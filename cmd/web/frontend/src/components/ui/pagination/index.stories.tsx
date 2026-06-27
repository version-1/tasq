import type { Meta, StoryObj } from "@storybook/react-vite";
import { Pagination } from "./index";

const meta = {
  title: "UI/Pagination",
  component: Pagination,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    nextDisabled: false,
    nextLabel: "Next",
    onNext: () => undefined,
    onPrevious: () => undefined,
    previousDisabled: false,
    previousLabel: "Previous",
    summary: "1-50 of 128 issues",
  },
} satisfies Meta<typeof Pagination>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const FirstPage: Story = {
  args: {
    previousDisabled: true,
  },
};

export const LastPage: Story = {
  args: {
    nextDisabled: true,
    summary: "101-128 of 128 issues",
  },
};
