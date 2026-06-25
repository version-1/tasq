import type { Meta, StoryObj } from "@storybook/react-vite";
import { RunRow } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/RunRow",
  component: RunRow,
  decorators: [
    (Story) => (
      <div style={{ border: "1px solid var(--border)", borderRadius: 8, overflow: "hidden" }}>
        <Story />
      </div>
    ),
  ],
  args: {
    issueID: 1,
    run: {
      run_id: "run-1-latest",
      thread_id: "thread_abc123",
      status: "running",
      attempt: 2,
      created_at: "2026-06-08T01:00:00.000Z",
      updated_at: "2026-06-08T02:00:00.000Z",
    },
  },
} satisfies Meta<typeof RunRow>;

export default meta;

type Story = StoryObj<typeof meta>;

export const WithThreadID: Story = {};

export const WithoutThreadID: Story = {
  args: {
    run: {
      run_id: "run-1-without-thread",
      status: "queued",
      attempt: 1,
      created_at: "2026-06-08T01:00:00.000Z",
      updated_at: "2026-06-08T01:05:00.000Z",
    },
  },
};
