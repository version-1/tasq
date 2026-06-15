import { render, screen, within } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ConversationPage } from "./index";

const api = vi.hoisted(() => ({
  fetchOrchestratorConversation: vi.fn(),
  fetchOrchestratorIssueRuntime: vi.fn(),
}));

vi.mock("@/lib/api", () => api);

function renderConversationPage() {
  render(
    <MemoryRouter initialEntries={["/issues/48/runs/run-1/conversations"]}>
      <Routes>
        <Route
          path="/issues/:id/runs/:runId/conversations"
          element={<ConversationPage />}
        />
      </Routes>
    </MemoryRouter>,
  );
}

describe("ConversationPage item completed events", () => {
  beforeEach(() => {
    api.fetchOrchestratorConversation.mockReset();
    api.fetchOrchestratorIssueRuntime.mockReset();
  });

  it("renders item completed command execution content in chronological order", async () => {
    api.fetchOrchestratorConversation.mockResolvedValue({
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

    renderConversationPage();

    await screen.findByText("Conversation history · run-1");
    const items = screen.getAllByRole("listitem");

    expect(within(items[0]).getByText("running")).toBeInTheDocument();
    expect(within(items[1]).getByText("commandExecution")).toBeInTheDocument();
    expect(within(items[1]).getByText("npm test")).toBeInTheDocument();
    expect(within(items[1]).getByRole("heading", { name: "Command output" })).toBeInTheDocument();
    expect(within(items[1]).getByText("Done")).toBeInTheDocument();
    expect(within(items[1]).getByText("extra text")).toBeInTheDocument();
    expect(within(items[1]).getByText("exit code 2")).toBeInTheDocument();
    expect(within(items[2]).getByText("Turn completed")).toBeInTheDocument();
    expect(within(items[2]).getByText("turn output")).toBeInTheDocument();
  });

  it("renders text items without a zero exit code indicator", async () => {
    api.fetchOrchestratorConversation.mockResolvedValue({
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

    renderConversationPage();

    const item = await screen.findByRole("listitem");

    expect(within(item).getByText("text")).toBeInTheDocument();
    expect(within(item).getByText("hello")).toBeInTheDocument();
    expect(within(item).getByText("world")).toBeInTheDocument();
    expect(within(item).queryByText(/exit code/)).not.toBeInTheDocument();
  });

  it("uses a fallback header for unknown completed item types", async () => {
    api.fetchOrchestratorConversation.mockResolvedValue({
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

    renderConversationPage();

    const item = await screen.findByRole("listitem");

    expect(within(item).getByText("customThing")).toBeInTheDocument();
    expect(within(item).getByText("custom body")).toBeInTheDocument();
  });
});
