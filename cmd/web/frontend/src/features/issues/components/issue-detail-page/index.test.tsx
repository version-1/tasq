import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
  AttachmentListResponse,
  CommentListResponse,
  Issue,
  OrchestratorConversation,
  OrchestratorIssueRuntime,
} from "@/lib/types";
import { IssueDetailPage } from "./index";

const api = vi.hoisted(() => ({
  attachmentContentURL: vi.fn((id: string) => `/tracker/api/v1/attachments/${id}/content`),
  fetchComments: vi.fn(),
  fetchIssue: vi.fn(),
  fetchIssueAttachments: vi.fn(),
  fetchOrchestratorConversation: vi.fn(),
  fetchOrchestratorIssueRuntime: vi.fn(),
  updateIssueStatus: vi.fn(),
}));

vi.mock("@/lib/api", () => api);

function renderIssueDetail(initialEntry = "/issues/42") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/issues/:id" element={<IssueDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("IssueDetailPage", () => {
  beforeEach(() => {
    api.attachmentContentURL.mockClear();
    api.fetchComments.mockReset();
    api.fetchIssue.mockReset();
    api.fetchIssueAttachments.mockReset();
    api.fetchOrchestratorConversation.mockReset();
    api.fetchOrchestratorIssueRuntime.mockReset();
    api.updateIssueStatus.mockReset();

    api.fetchIssue.mockResolvedValue(issue);
    api.fetchIssueAttachments.mockResolvedValue(attachmentsResponse);
    api.fetchOrchestratorIssueRuntime.mockResolvedValue(runtime);
    api.fetchComments.mockResolvedValue(commentsResponse);
    api.fetchOrchestratorConversation.mockResolvedValue(conversation);
  });

  it("shows the details tab by default without loading comments", async () => {
    renderIssueDetail();

    expect(await screen.findByRole("heading", { name: "Attachments" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Refine issue detail" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "screenshot.png" })).toHaveAttribute(
      "href",
      "/tracker/api/v1/attachments/attachment-1/content",
    );
    expect(screen.queryByRole("heading", { name: "Runs" })).not.toBeInTheDocument();
    expect(api.fetchIssueAttachments).toHaveBeenCalledWith(42, { silent: true });
    expect(api.fetchComments).not.toHaveBeenCalled();
  });

  it("treats a null attachment list as empty", async () => {
    api.fetchIssueAttachments.mockResolvedValueOnce({ data: null, meta: {} });

    renderIssueDetail();

    expect(await screen.findByRole("heading", { name: "Attachments" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Refine issue detail" })).not.toBeInTheDocument();
    expect(await screen.findByText("No attachments")).toBeInTheDocument();
  });

  it("updates status from the basic info grid", async () => {
    const user = userEvent.setup();
    api.updateIssueStatus.mockResolvedValueOnce({ ...issue, status: "in_progress" });

    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Change status" }));
    await user.click(screen.getByRole("button", { name: "In Progress" }));
    expect(api.updateIssueStatus).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(api.updateIssueStatus).toHaveBeenCalledWith(42, "in_progress");
    expect(screen.getByRole("button", { name: "Change status" })).toHaveTextContent("in_progress");
  });

  it("loads comments when the comments tab is selected", async () => {
    renderIssueDetail("/issues/42?tab=comments");

    expect(await screen.findByText("Looks good from design review.")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Runs" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Refine issue detail" })).not.toBeInTheDocument();
    expect(api.fetchComments).toHaveBeenCalledWith(42, undefined, 20, { silent: true });
  });

  it("loads the latest run conversation and filters message types", async () => {
    const user = userEvent.setup();
    renderIssueDetail("/issues/42?tab=conversation");

    await waitFor(() => {
      expect(api.fetchOrchestratorConversation).toHaveBeenCalledWith(42, "run-latest", {
        silent: true,
      });
    });
    expect(screen.queryByRole("heading", { name: "Refine issue detail" })).not.toBeInTheDocument();
    expect(await screen.findByText("runner started")).toBeInTheDocument();
    expect(screen.getByText("run failed")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Message types:/ }));
    await user.click(screen.getByLabelText("running"));
    await user.click(screen.getByRole("button", { name: "Apply" }));

    const items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(1);
    expect(within(items[0]).getByText("runner started")).toBeInTheDocument();
    expect(screen.queryByText("run failed")).not.toBeInTheDocument();
  });
});

const issue: Issue = {
  id: 42,
  projectId: 1,
  projectKey: "TASQ",
  title: "Refine issue detail",
  description: "Show details in tabs.",
  status: "ready",
  priority: "high",
  assignee: "frontend",
  dependency_ids: [],
  createdAt: "2026-06-20T00:00:00.000Z",
  updatedAt: "2026-06-21T00:00:00.000Z",
};

const attachmentsResponse: AttachmentListResponse = {
  data: [
    {
      id: "attachment-1",
      entityType: "issue",
      entityId: "42",
      filename: "screenshot.png",
      contentType: "image/png",
      size: 2048,
      createdAt: "2026-06-20T01:00:00.000Z",
    },
  ],
  meta: {},
};

const commentsResponse: CommentListResponse = {
  data: [
    {
      id: 1,
      issueId: 42,
      author: "designer",
      type: "general",
      body: "Looks good from design review.",
      createdAt: "2026-06-21T02:00:00.000Z",
    },
  ],
  meta: {
    cursor: 0,
    limit: 20,
    nextCursor: null,
  },
};

const runtime: OrchestratorIssueRuntime = {
  issue_identifier: "issue-42",
  issue_id: "42",
  status: "succeeded",
  workspace: {
    path: "/tmp/tasq",
  },
  attempts: {
    restart_count: 0,
    current_retry_attempt: 1,
  },
  runs: [
    {
      run_id: "run-latest",
      status: "failed",
      attempt: 2,
      created_at: "2026-06-21T03:00:00.000Z",
      updated_at: "2026-06-21T03:10:00.000Z",
    },
  ],
  running: null,
  retry: null,
  logs: {
    codex_session_logs: [],
  },
  recent_events: [],
  last_error: null,
  tracked: {},
};

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
      at: "2026-06-21T03:10:00.000Z",
      event: "failed",
      message: "run failed",
      payload_json: "",
    },
  ],
};
