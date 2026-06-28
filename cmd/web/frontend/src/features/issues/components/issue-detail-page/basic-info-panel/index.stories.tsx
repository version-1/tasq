import type { Meta, StoryObj } from "@storybook/react-vite";
import { MemoryRouter } from "react-router-dom";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { storySummary } from "@/stories/fixtures";
import { BasicInfoPanel } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/BasicInfoPanel",
  component: BasicInfoPanel,
  decorators: [
    (Story) => (
      <MemoryRouter>
        <Story />
      </MemoryRouter>
    ),
  ],
  args: {
    disabled: false,
    issue: issueFixtures[2],
    issueOptions: storySummary.columns.flatMap((column) => column.issues),
    onStatusChange: async () => undefined,
  },
} satisfies Meta<typeof BasicInfoPanel>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
