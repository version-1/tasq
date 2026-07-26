import type { Meta, StoryObj } from "@storybook/react-vite";
import type { OrchestratorConversationEvent } from "@/lib/types";
import { ItemCompletedEventBody } from "./body/item-completed";
import { EventBodyPreview } from "./body/preview";
import { RateLimitsEventBody } from "./body/rate-limits";
import { StatusEventBody } from "./body/status";
import { TokenUsageEventBody } from "./body/token-usage";
import { EventHeader } from "./event-header";
import { EventCard } from "./index";

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

export const HeaderFolded: Story = {
  args: {
    event: commandOutputEvent,
  },
  render: ({ event }) => (
    <EventHeader
      bodyID="storybook-conversation-event-body"
      event={event}
      isBodyOpen={false}
      onToggleBody={() => undefined}
    />
  ),
};

export const FoldedPreview: Story = {
  args: {
    event: commandOutputEvent,
  },
  render: ({ event }) => <EventBodyPreview event={event} />,
};

export const ItemCompletedBody: Story = {
  args: {
    event: commandOutputEvent,
  },
  render: ({ event }) => <ItemCompletedEventBody event={event} />,
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
