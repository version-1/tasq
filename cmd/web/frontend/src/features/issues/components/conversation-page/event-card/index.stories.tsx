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
