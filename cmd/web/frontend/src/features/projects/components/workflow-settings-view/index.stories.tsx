import type { Meta, StoryObj } from "@storybook/react-vite";
import { workflowFixtures } from "@/mocks/fixtures/workflows";
import { WorkflowSettingsView } from "./index";

const meta = {
  title: "Features/Projects/WorkflowSettingsView",
  component: WorkflowSettingsView,
  args: {
    workflow: workflowFixtures[0],
  },
} satisfies Meta<typeof WorkflowSettingsView>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
