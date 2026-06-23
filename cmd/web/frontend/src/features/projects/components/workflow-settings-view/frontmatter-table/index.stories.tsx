import type { Meta, StoryObj } from "@storybook/react-vite";
import { workflowFixtures } from "@/mocks/fixtures/workflows";
import { FrontmatterTable } from "./index";
import { toFrontmatterRows } from "./rows";

const meta = {
  title: "Features/Projects/WorkflowSettings/FrontmatterTable",
  component: FrontmatterTable,
  args: {
    rows: toFrontmatterRows(workflowFixtures[0].frontmatter),
  },
} satisfies Meta<typeof FrontmatterTable>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};
