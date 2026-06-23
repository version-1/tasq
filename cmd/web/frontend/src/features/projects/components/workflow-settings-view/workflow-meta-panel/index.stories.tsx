import type { Meta, StoryObj } from "@storybook/react-vite";
import { workflowFixtures } from "@/mocks/fixtures/workflows";
import { WorkflowMetaPanel } from "./index";

const meta = {
  title: "Features/Projects/WorkflowSettings/WorkflowMetaPanel",
  component: WorkflowMetaPanel,
  args: {
    updatedAt: workflowFixtures[0].updatedAt,
  },
} satisfies Meta<typeof WorkflowMetaPanel>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
