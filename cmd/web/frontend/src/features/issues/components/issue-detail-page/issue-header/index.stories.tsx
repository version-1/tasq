import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { IssueHeader } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/IssueHeader",
  component: IssueHeader,
  args: {
    disabled: false,
    issue: issueFixtures[1],
    onStatusChange: async () => undefined,
  },
} satisfies Meta<typeof IssueHeader>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
