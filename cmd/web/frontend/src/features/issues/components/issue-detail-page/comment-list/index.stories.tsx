import type { Meta, StoryObj } from "@storybook/react-vite";
import { noop, storyComments } from "@/stories/fixtures";
import { CommentList } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/CommentList",
  component: CommentList,
  args: {
    comments: storyComments,
    error: "",
    hasMore: true,
    isLoading: false,
    onLoadMore: noop,
  },
} satisfies Meta<typeof CommentList>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};

export const Empty: Story = {
  args: {
    comments: [],
    hasMore: false,
  },
};
