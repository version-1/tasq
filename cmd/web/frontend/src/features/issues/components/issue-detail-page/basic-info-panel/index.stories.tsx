import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { BasicInfoPanel } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/BasicInfoPanel",
  component: BasicInfoPanel,
  args: {
    disabled: false,
    issue: issueFixtures[1],
    onStatusChange: async () => undefined,
  },
} satisfies Meta<typeof BasicInfoPanel>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
