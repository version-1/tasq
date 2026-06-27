import type { Meta, StoryObj } from "@storybook/react-vite";
import { issueStatuses } from "@/lib/types";
import { IssueFilterOptions } from "./index";

const meta = {
  title: "Features/Issues/IssueFilterOptions",
  component: IssueFilterOptions,
  decorators: [
    (Story) => (
      <div style={{ padding: 16 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    allLabel: "All statuses",
    applyLabel: "Apply",
    cancelLabel: "Cancel",
    clearLabel: "Clear all",
    label: "Status",
    onChange: () => undefined,
    options: issueStatuses.map((status) => ({ label: status, value: status })),
    selectedCountLabel: (count: number) => `${count} selected`,
    selectedValues: ["backlog", "ready", "in_progress"],
  },
} satisfies Meta<typeof IssueFilterOptions>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  args: {
    selectedValues: [],
  },
};
