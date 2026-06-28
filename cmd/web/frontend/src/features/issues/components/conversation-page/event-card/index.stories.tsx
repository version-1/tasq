import type { Meta, StoryObj } from "@storybook/react-vite";
import type { OrchestratorConversationEvent } from "@/lib/types";
import { ApprovalRequestEventBody } from "./approval-request-event-body";
import { EventCard } from "./index";
import { ItemCompletedEventBody } from "./item-completed-event-body";
import { RateLimitsEventBody } from "./rate-limits-event-body";
import { StatusEventBody } from "./status-event-body";
import { TokenUsageEventBody } from "./token-usage-event-body";
import { TurnCompletedEventBody } from "./turn-completed-event-body";

const commandOutputEvent = {
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
} satisfies OrchestratorConversationEvent;

const approvalRequestEvent = {
  at: "2026-06-01T08:04:30.000Z",
  event: "item/commandExecution/requestApproval",
  message: "approval requested",
  payload_json: JSON.stringify({
    reason: "Need to run the frontend typechecker.",
    command: "npm run typecheck",
  }),
} satisfies OrchestratorConversationEvent;

const tokenUsageEvent = {
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
} satisfies OrchestratorConversationEvent;

const rateLimitsEvent = {
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
} satisfies OrchestratorConversationEvent;

const turnCompletedEvent = {
  at: "2026-06-01T08:07:02.000Z",
  event: "turn_completed",
  message: "turn completed",
  payload_json: JSON.stringify({
    aggregatedOutput:
      "## Summary\n\nThe run updated the conversation view.\n\n- Added folded event cards\n- Kept event details readable",
  }),
} satisfies OrchestratorConversationEvent;

const statusEvent = {
  at: "2026-06-01T08:03:02.000Z",
  event: "running",
  message: "runner started",
  payload_json: "",
} satisfies OrchestratorConversationEvent;

const meta = {
  title: "Features/Issues/ConversationPage/EventCard",
  component: EventCard,
  args: {
    event: commandOutputEvent,
  },
} satisfies Meta<typeof EventCard>;

export default meta;

type Story = StoryObj<typeof meta>;

export const CommandOutput: Story = {};

export const ApprovalRequest: Story = {
  args: {
    event: approvalRequestEvent,
  },
};

export const TokenUsage: Story = {
  args: {
    event: tokenUsageEvent,
  },
};

export const RateLimits: Story = {
  args: {
    event: rateLimitsEvent,
  },
};

export const TurnCompletedBody: Story = {
  args: {
    event: turnCompletedEvent,
  },
  render: ({ event }) => <TurnCompletedEventBody event={event} />,
};

export const ItemCompletedBody: Story = {
  args: {
    event: commandOutputEvent,
  },
  render: ({ event }) => <ItemCompletedEventBody event={event} />,
};

export const ApprovalRequestBody: Story = {
  args: {
    event: approvalRequestEvent,
  },
  render: ({ event }) => <ApprovalRequestEventBody event={event} />,
};

export const TokenUsageBody: Story = {
  args: {
    event: tokenUsageEvent,
  },
  render: ({ event }) => <TokenUsageEventBody event={event} />,
};

export const RateLimitsBody: Story = {
  args: {
    event: rateLimitsEvent,
  },
  render: ({ event }) => <RateLimitsEventBody event={event} />,
};

export const StatusBody: Story = {
  args: {
    event: statusEvent,
  },
  render: ({ event }) => <StatusEventBody event={event} />,
};
