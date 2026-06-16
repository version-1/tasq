import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { IssueDetail } from "./index";

const meta = {
  title: "Issue/Detail",
  component: IssueDetail,
  decorators: [
    (Story) => (
      <div style={{ minHeight: 520, position: "relative" }}>
        <Story />
      </div>
    ),
  ],
  args: {
    issue: issueFixtures[0],
    onStatusChange: async () => undefined,
  },
} satisfies Meta<typeof IssueDetail>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Unassigned: Story = {
  args: {
    issue: issueFixtures[3],
  },
};

export const EmptyDescription: Story = {
  args: {
    issue: {
      ...issueFixtures[0],
      description: "",
    },
  },
};
