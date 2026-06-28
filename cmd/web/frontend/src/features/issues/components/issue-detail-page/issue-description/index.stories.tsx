import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { IssueDescription } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/IssueDescription",
  component: IssueDescription,
  args: {
    issue: issueFixtures[0],
    onSave: async () => undefined,
  },
} satisfies Meta<typeof IssueDescription>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
