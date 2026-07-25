import type { Meta, StoryObj } from "@storybook/react-vite";
import { storyComments } from "@/stories/fixtures";
import { CommentCard } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/CommentCard",
  component: CommentCard,
  args: {
    comment: storyComments[0],
  },
} satisfies Meta<typeof CommentCard>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
