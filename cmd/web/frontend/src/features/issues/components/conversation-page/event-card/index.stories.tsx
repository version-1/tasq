import type { Meta, StoryObj } from "@storybook/react-vite";
import { EventCard } from "./index";

const meta = {
  title: "Features/Issues/ConversationPage/EventCard",
  component: EventCard,
  args: {
    event: {
      at: "2026-06-01T08:05:12.000Z",
      event: "item/completed",
      message: "command completed",
      payload_json: JSON.stringify({
        item: {
          type: "commandExecution",
          command: "npm test -- conversation",
          aggregatedOutput: "## Command output\n\nAll conversation tests passed.",
          exitCode: 0,
        },
      }),
    },
  },
} satisfies Meta<typeof EventCard>;

export default meta;

type Story = StoryObj<typeof meta>;

export const CommandOutput: Story = {};

export const ApprovalRequest: Story = {
  args: {
    event: {
      at: "2026-06-01T08:04:30.000Z",
      event: "item/commandExecution/requestApproval",
      message: "approval requested",
      payload_json: JSON.stringify({
        reason: "Need to run the frontend typechecker.",
        command: "npm run typecheck",
      }),
    },
  },
};

export const TokenUsage: Story = {
  args: {
    event: {
      at: "2026-06-01T08:06:49.000Z",
      event: "thread/tokenUsage/updated",
      message: "token usage updated",
      payload_json: JSON.stringify({
        tokenUsage: {
          total: {
            totalTokens: 8712831,
            inputTokens: 8684367,
            cachedInputTokens: 8415104,
            outputTokens: 28464,
            reasoningOutputTokens: 5186,
          },
          last: {
            totalTokens: 146759,
            inputTokens: 146356,
            cachedInputTokens: 145280,
            outputTokens: 403,
            reasoningOutputTokens: 203,
          },
          modelContextWindow: 258400,
        },
      }),
    },
  },
};

export const RateLimits: Story = {
  args: {
    event: {
      at: "2026-06-01T08:06:54.000Z",
      event: "account/rateLimits/updated",
      message: "rate limits updated",
      payload_json: JSON.stringify({
        rateLimits: {
          limitId: "codex",
          limitName: null,
          primary: {
            usedPercent: 13,
            windowDurationMins: 300,
            resetsAt: 1781415835,
          },
          secondary: {
            usedPercent: 13,
            windowDurationMins: 10080,
            resetsAt: 1781572462,
          },
          credits: null,
          individualLimit: null,
          planType: "pro",
          rateLimitReachedType: null,
        },
      }),
    },
  },
};
