import type { Meta, StoryObj } from "@storybook/react-vite";
import type { OrchestratorConversation, OrchestratorIssueRun } from "@/lib/types";
import { ConversationTab, defaultConversationMessageTypes } from "./index";

const runs: OrchestratorIssueRun[] = [
  {
    run_id: "run-latest",
    thread_id: "thread-latest",
    status: "succeeded",
    attempt: 2,
    created_at: "2026-06-21T03:00:00.000Z",
    updated_at: "2026-06-21T03:10:00.000Z",
  },
  {
    run_id: "run-previous",
    status: "failed",
    attempt: 1,
    created_at: "2026-06-20T03:00:00.000Z",
    updated_at: "2026-06-20T03:10:00.000Z",
  },
];

const conversation: OrchestratorConversation = {
  issue_identifier: "issue-42",
  issue_id: "42",
  run_id: "run-latest",
  events: [
    {
      at: "2026-06-21T03:00:00.000Z",
      event: "running",
      message: "runner started",
      payload_json: "",
    },
    {
      at: "2026-06-21T03:05:00.000Z",
      event: "item/completed",
      message: "item completed",
      payload_json: JSON.stringify({
        item: {
          type: "commandExecution",
          command: "npm test",
          aggregatedOutput: "All tests passed.",
          exitCode: 0,
        },
      }),
    },
  ],
};

const meta = {
  title: "Features/Issues/IssueDetail/ConversationTab",
  component: ConversationTab,
  args: {
    conversation,
    error: "",
    isLoading: false,
    messageTypes: [...defaultConversationMessageTypes],
    onMessageTypesChange: () => undefined,
    onRunChange: () => undefined,
    runs,
    runsError: "",
    runsLoading: false,
    selectedRunID: "run-latest",
  },
} satisfies Meta<typeof ConversationTab>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};

export const Empty: Story = {
  args: {
    conversation: null,
    runs: [],
    selectedRunID: "",
  },
};
