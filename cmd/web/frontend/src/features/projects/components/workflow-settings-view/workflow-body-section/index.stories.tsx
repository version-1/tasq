import type { Meta, StoryObj } from "@storybook/react-vite";
import { workflowFixtures } from "@/mocks/fixtures/workflows";
import { WorkflowBodySection } from "./index";

const meta = {
  title: "Features/Projects/WorkflowSettings/WorkflowBodySection",
  component: WorkflowBodySection,
  args: {
    body: workflowFixtures[0].body,
  },
} satisfies Meta<typeof WorkflowBodySection>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};

export const Empty: Story = {
  args: {
    body: "",
  },
};
