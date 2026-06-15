import { render, screen, within } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { OrchestratorConversation } from "@/lib/types";
import { ConversationPage } from "./index";

const { fetchOrchestratorConversation, fetchOrchestratorIssueRuntime } = vi.hoisted(() => ({
  fetchOrchestratorConversation: vi.fn(),
  fetchOrchestratorIssueRuntime: vi.fn(),
}));

vi.mock("@/lib/api", () => ({
  fetchOrchestratorConversation,
  fetchOrchestratorIssueRuntime,
}));

function renderConversationPage(conversation: OrchestratorConversation) {
  fetchOrchestratorConversation.mockResolvedValue(conversation);

  render(
    <MemoryRouter initialEntries={[`/issues/${conversation.issue_id}/runs/${conversation.run_id}/conversations`]}>
      <Routes>
        <Route path="/issues/:id/runs/:runId/conversations" element={<ConversationPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("ConversationPage", () => {
  beforeEach(() => {
    fetchOrchestratorConversation.mockReset();
    fetchOrchestratorIssueRuntime.mockReset();
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

    const approvalRequest = await screen.findByRole("region", { name: "Approval request" });

    expect(
      within(approvalRequest).getByRole("heading", { name: "needs elevated filesystem access" }),
    ).toBeInTheDocument();
    expect(within(approvalRequest).getByText("npm run build")).toBeInTheDocument();
    expect(screen.getByText("approval requested")).toBeInTheDocument();
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

    const approvalRequest = await screen.findByRole("region", { name: "Approval request" });

    expect(
      within(approvalRequest).getByRole("heading", { name: "needs command approval" }),
    ).toBeInTheDocument();
    expect(within(approvalRequest).getByText("make test")).toBeInTheDocument();
  });
});
