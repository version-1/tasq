import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { OrchestratorConversation } from "@/lib/types";
import { ConversationPage } from "./index";

const api = vi.hoisted(() => ({
  fetchOrchestratorConversation: vi.fn(),
  fetchOrchestratorIssueRuntime: vi.fn(),
}));

vi.mock("@/lib/api", () => api);

function renderConversationPage(conversation: OrchestratorConversation) {
  api.fetchOrchestratorConversation.mockResolvedValue(conversation);

  render(
    <MemoryRouter initialEntries={[`/issues/${conversation.issue_id}/runs/${conversation.run_id}/conversations`]}>
      <Routes>
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={<ConversationPage />}
        />
      </Routes>
    </MemoryRouter>,
  );
}

function getCodeByText(container: HTMLElement, text: string) {
  return within(container).getByText((_, node) => (
    node?.tagName.toLowerCase() === "code" &&
    node.textContent?.replace(/\s+/g, " ").trim() === text
  ));
}

async function expandEventBody(user: ReturnType<typeof userEvent.setup>, item: HTMLElement) {
  await user.click(
    within(item).getByRole("button", {
      name: "Expand conversation event body",
    }),
  );
}

describe("ConversationPage", () => {
  beforeEach(() => {
    api.fetchOrchestratorConversation.mockReset();
    api.fetchOrchestratorIssueRuntime.mockReset();
  });

  it("renders item completed command execution content in chronological order", async () => {
    const user = userEvent.setup();

    renderConversationPage({
      issue_identifier: "issue-48",
      issue_id: "48",
      run_id: "run-1",
      events: [
        {
          at: "2026-06-15T01:00:00Z",
          event: "running",
          message: "runner started",
          payload_json: "",
        },
        {
          at: "2026-06-15T01:01:00Z",
          event: "item/completed",
          message: "item completed",
          payload_json: JSON.stringify({
            item: {
              type: "commandExecution",
              command: "npm test",
              aggregatedOutput: "## Command output\n\nDone",
              text: "extra text",
              exitCode: 2,
            },
          }),
        },
        {
          at: "2026-06-15T01:02:00Z",
          event: "turn_completed",
          message: "turn_id=turn-1",
          payload_json: JSON.stringify({ aggregatedOutput: "turn output" }),
        },
      ],
    });

    await screen.findByText("Conversation history · run-1");
    const items = screen.getAllByRole("listitem");

    expect(within(items[0]).getByText("running")).toBeInTheDocument();
    expect(within(items[1]).getByText("commandExecution")).toBeInTheDocument();
    expect(within(items[1]).getByText("npm test")).toBeInTheDocument();

    await expandEventBody(user, items[1]);

    expect(within(items[1]).getByRole("heading", { name: "Command output" })).toBeInTheDocument();
    expect(within(items[1]).getByText("Done")).toBeInTheDocument();
    expect(within(items[1]).getByText("extra text")).toBeInTheDocument();
    expect(within(items[1]).getByText("exit code 2")).toBeInTheDocument();
    expect(within(items[2]).getByText("Turn completed")).toBeInTheDocument();

    await expandEventBody(user, items[2]);

    expect(within(items[2]).getByText("turn output")).toBeInTheDocument();
  });

  it("renders text items without a zero exit code indicator", async () => {
    const user = userEvent.setup();

    renderConversationPage({
      issue_identifier: "issue-48",
      issue_id: "48",
      run_id: "run-1",
      events: [
        {
          at: "2026-06-15T01:00:00Z",
          event: "item/completed",
          message: "item completed",
          payload_json: JSON.stringify({
            item: {
              type: "text",
              text: "hello **world**",
              exitCode: 0,
            },
          }),
        },
      ],
    });

    const item = await screen.findByRole("listitem");

    expect(within(item).getByText("text")).toBeInTheDocument();
    await expandEventBody(user, item);
    expect(within(item).getByText("hello")).toBeInTheDocument();
    expect(within(item).getByText("world")).toBeInTheDocument();
    expect(within(item).queryByText(/exit code/)).not.toBeInTheDocument();
  });

  it("folds and expands event body content", async () => {
    const user = userEvent.setup();

    renderConversationPage({
      issue_identifier: "issue-48",
      issue_id: "48",
      run_id: "run-fold",
      events: [
        {
          at: "2026-06-15T01:00:00Z",
          event: "item/completed",
          message: "item completed",
          payload_json: JSON.stringify({
            item: {
              type: "text",
              text: "## Folded heading\n\nfoldable body content",
            },
          }),
        },
      ],
    });

    const item = await screen.findByRole("listitem");
    expect(
      within(item).getByRole("button", {
        name: "Expand conversation event body",
      }),
    ).toBeInTheDocument();
    expect(within(item).getByText("## Folded heading foldable body content")).toBeInTheDocument();
    expect(within(item).queryByRole("heading", { name: "Folded heading" })).not.toBeInTheDocument();

    await user.click(
      within(item).getByRole("button", {
        name: "Expand conversation event body",
      }),
    );

    expect(within(item).getByRole("heading", { name: "Folded heading" })).toBeVisible();
    expect(
      within(item).getByRole("button", {
        name: "Collapse conversation event body",
      }),
    ).toBeInTheDocument();
  });

  it("truncates folded event body preview to 100 characters", async () => {
    const longBody = `${"a".repeat(100)}b`;

    renderConversationPage({
      issue_identifier: "issue-48",
      issue_id: "48",
      run_id: "run-fold-preview",
      events: [
        {
          at: "2026-06-15T01:00:00Z",
          event: "item/completed",
          message: "item completed",
          payload_json: JSON.stringify({
            item: {
              type: "text",
              text: longBody,
            },
          }),
        },
      ],
    });

    const item = await screen.findByRole("listitem");

    expect(within(item).getByText(`${"a".repeat(100)}...`)).toBeInTheDocument();
  });

  it("uses a fallback header for unknown completed item types", async () => {
    renderConversationPage({
      issue_identifier: "issue-48",
      issue_id: "48",
      run_id: "run-1",
      events: [
        {
          at: "2026-06-15T01:00:00Z",
          event: "item/completed",
          message: "item completed",
          payload_json: JSON.stringify({
            item: {
              type: "customThing",
              aggregatedOutput: "custom body",
            },
          }),
        },
      ],
    });

    const item = await screen.findByRole("listitem");

    expect(within(item).getByText("customThing")).toBeInTheDocument();
    expect(within(item).getByText("custom body")).toBeInTheDocument();
  });

  it("renders command approval requests with the reason and command only", async () => {
    renderConversationPage({
      issue_identifier: "issue-49",
      issue_id: "49",
      run_id: "run-approval",
      events: [
        {
          at: "2026-06-15T03:00:00Z",
          event: "item/commandExecution/requestApproval",
          message: "approval requested",
          payload_json: JSON.stringify({
            reason: "needs elevated filesystem access",
            command: "npm run build",
            availableDecisions: ["approve", "deny"],
            proposedExecpolicyAmendment: { sandbox_permissions: "require_escalated" },
          }),
        },
      ],
    });

    const item = await screen.findByRole("listitem");
    const approvalRequest = within(item).getByRole("region", { name: "Approval request" });

    expect(within(item).getByText("approval requested")).toBeInTheDocument();
    expect(within(item).queryByRole("button", { name: "Expand conversation event body" })).not.toBeInTheDocument();
    expect(within(approvalRequest).getByText("needs elevated filesystem access")).toBeInTheDocument();
    expect(getCodeByText(approvalRequest, "npm run build")).toBeInTheDocument();
    expect(screen.queryByText("availableDecisions")).not.toBeInTheDocument();
    expect(screen.queryByText("proposedExecpolicyAmendment")).not.toBeInTheDocument();
  });

  it("finds approval request details nested under params", async () => {
    renderConversationPage({
      issue_identifier: "issue-49",
      issue_id: "49",
      run_id: "run-nested-approval",
      events: [
        {
          at: "2026-06-15T03:00:00Z",
          event: "item/commandExecution/requestApproval",
          message: "",
          payload_json: JSON.stringify({
            id: 1,
            method: "item/commandExecution/requestApproval",
            params: {
              command: "make test",
              reason: "needs command approval",
            },
          }),
        },
      ],
    });

    const item = await screen.findByRole("listitem");
    const approvalRequest = within(item).getByRole("region", { name: "Approval request" });

    expect(within(approvalRequest).getByText("needs command approval")).toBeInTheDocument();
    expect(getCodeByText(approvalRequest, "make test")).toBeInTheDocument();
  });

  it("renders token usage and rate limit update events", async () => {
    const user = userEvent.setup();

    renderConversationPage({
      issue_identifier: "issue-50",
      issue_id: "50",
      run_id: "run-usage",
      events: [
        {
          at: "2026-06-15T04:00:00Z",
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
        {
          at: "2026-06-15T04:01:00Z",
          event: "account/rateLimits/updated",
          message: "rate limits updated",
          payload_json: JSON.stringify({
            rateLimits: {
              limitId: "codex",
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
              planType: "pro",
              rateLimitReachedType: null,
            },
          }),
        },
      ],
    });

    await screen.findByText("Conversation history · run-usage");
    const items = screen.getAllByRole("listitem");

    expect(screen.getByText("token usage updated")).toBeInTheDocument();

    await expandEventBody(user, items[0]);

    expect(screen.getByRole("region", { name: "Token usage" })).toBeInTheDocument();
    expect(screen.getByText("8,712,831")).toBeInTheDocument();
    expect(screen.getByText("Context window 258,400 tokens")).toBeInTheDocument();
    expect(screen.getByText("rate limits updated")).toBeInTheDocument();

    await expandEventBody(user, items[1]);

    expect(screen.getByRole("region", { name: "Rate limits" })).toBeInTheDocument();
    expect(screen.getByText("codex")).toBeInTheDocument();
    expect(screen.getByText("pro")).toBeInTheDocument();
    expect(screen.getByText("Not reached")).toBeInTheDocument();
  });
});
