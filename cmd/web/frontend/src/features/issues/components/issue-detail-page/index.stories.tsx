import type { Meta, StoryObj } from "@storybook/react-vite";
import { IssueDetailPage } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/IssueDetailPage",
  component: IssueDetailPage,
} satisfies Meta<typeof IssueDetailPage>;

export default meta;

type Story = StoryObj<typeof meta>;

export const RouteContainer: Story = {};
