import type { Meta, StoryObj } from "@storybook/react-vite";
import { workflowFixtures } from "@/mocks/fixtures/workflows";
import { FrontmatterSection } from "./index";

const meta = {
  title: "Features/Projects/WorkflowSettings/FrontmatterSection",
  component: FrontmatterSection,
  args: {
    frontmatter: workflowFixtures[0].frontmatter,
  },
} satisfies Meta<typeof FrontmatterSection>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};

export const Empty: Story = {
  args: {
    frontmatter: {},
  },
};
