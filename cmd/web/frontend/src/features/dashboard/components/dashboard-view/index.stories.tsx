import type { Meta, StoryObj } from "@storybook/react-vite";
import { storySummary } from "@/stories/fixtures";
import { DashboardView } from "./index";

const meta = {
  title: "Features/Dashboard/DashboardView",
  component: DashboardView,
  args: {
    summary: storySummary,
    issues: storySummary.columns.flatMap((column) => column.issues),
    refreshIntervalMs: 3000,
  },
} satisfies Meta<typeof DashboardView>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
