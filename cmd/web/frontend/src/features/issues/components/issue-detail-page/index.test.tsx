import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
  AttachmentListResponse,
  ChangeRequestListResponse,
  CommentListResponse,
  Issue,
  OrchestratorConversation,
  OrchestratorIssueRuntime,
} from "@/lib/types";
import { IssueDetailPage } from "./index";

const api = vi.hoisted(() => ({
  attachmentContentURL: vi.fn((id: string) => `/tracker/api/v1/attachments/${id}/content`),
  createChangeRequest: vi.fn(),
  fetchChangeRequests: vi.fn(),
  fetchComments: vi.fn(),
  fetchIssue: vi.fn(),
  fetchIssueAttachments: vi.fn(),
  fetchOrchestratorConversation: vi.fn(),
  fetchOrchestratorIssueRuntime: vi.fn(),
  updateIssueDescription: vi.fn(),
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
    api.createChangeRequest.mockReset();
    api.fetchChangeRequests.mockReset();
    api.fetchComments.mockReset();
    api.fetchIssue.mockReset();
    api.fetchIssueAttachments.mockReset();
    api.fetchOrchestratorConversation.mockReset();
    api.fetchOrchestratorIssueRuntime.mockReset();
    api.updateIssueDescription.mockReset();
    api.updateIssueStatus.mockReset();

    api.fetchIssue.mockResolvedValue(issue);
    api.fetchIssueAttachments.mockResolvedValue(attachmentsResponse);
    api.fetchOrchestratorIssueRuntime.mockResolvedValue(runtime);
    api.fetchChangeRequests.mockResolvedValue(changeRequestsResponse);
    api.fetchComments.mockResolvedValue(commentsResponse);
    api.fetchOrchestratorConversation.mockResolvedValue(conversation);
    api.createChangeRequest.mockResolvedValue(changeRequest);
    api.updateIssueDescription.mockResolvedValue({ ...issue, description: "Updated **description**" });
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
    expect(api.fetchChangeRequests).not.toHaveBeenCalled();
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

  it("rejects a review issue with a change request", async () => {
    const user = userEvent.setup();
    api.fetchIssue.mockResolvedValueOnce({ ...issue, status: "review" });
    api.updateIssueStatus.mockResolvedValueOnce({ ...issue, status: "ready" });

    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Reject" }));
    const dialog = screen.getByRole("dialog", { name: "Reject #42" });
    await user.type(screen.getByRole("textbox", { name: "Request" }), "  Please cover the empty state  ");
    await user.click(within(dialog).getByRole("button", { name: "Reject" }));

    expect(api.createChangeRequest).toHaveBeenCalledWith(42, {
      author: "reviewer",
      body: "Please cover the empty state",
    }, { silent: true });
    expect(api.updateIssueStatus).toHaveBeenCalledWith(42, "ready", { silent: true });
    expect(await screen.findByRole("button", { name: "Change status" })).toHaveTextContent("ready");
    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "Reject #42" })).not.toBeInTheDocument();
    });
  });

  it("requires reject request body before creating a change request", async () => {
    const user = userEvent.setup();
    api.fetchIssue.mockResolvedValueOnce({ ...issue, status: "review" });

    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Reject" }));
    const dialog = screen.getByRole("dialog", { name: "Reject #42" });
    await user.click(within(dialog).getByRole("button", { name: "Reject" }));

    expect(await screen.findByText("Enter a request")).toBeInTheDocument();
    expect(api.createChangeRequest).not.toHaveBeenCalled();
    expect(api.updateIssueStatus).not.toHaveBeenCalled();
  });

  it("does not move the issue when change request creation fails", async () => {
    const user = userEvent.setup();
    api.fetchIssue.mockResolvedValueOnce({ ...issue, status: "review" });
    api.createChangeRequest.mockRejectedValueOnce(new Error("failed to create request"));

    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Reject" }));
    const dialog = screen.getByRole("dialog", { name: "Reject #42" });
    await user.type(screen.getByRole("textbox", { name: "Request" }), "Fix the tests");
    await user.click(within(dialog).getByRole("button", { name: "Reject" }));

    expect(await screen.findByText("failed to create request")).toBeInTheDocument();
    expect(api.updateIssueStatus).not.toHaveBeenCalled();
    expect(screen.getByRole("dialog", { name: "Reject #42" })).toBeInTheDocument();
  });

  it("retries only the ready update when rejecting after a status update failure", async () => {
    const user = userEvent.setup();
    api.fetchIssue.mockResolvedValueOnce({ ...issue, status: "review" });
    api.updateIssueStatus
      .mockRejectedValueOnce(new Error("failed to update status"))
      .mockResolvedValueOnce({ ...issue, status: "ready" });

    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Reject" }));
    const dialog = screen.getByRole("dialog", { name: "Reject #42" });
    await user.type(screen.getByRole("textbox", { name: "Request" }), "Fix review feedback");
    await user.click(within(dialog).getByRole("button", { name: "Reject" }));
    expect(await screen.findByText("failed to update status")).toBeInTheDocument();

    await user.click(within(dialog).getByRole("button", { name: "Reject" }));

    expect(api.createChangeRequest).toHaveBeenCalledTimes(1);
    expect(api.updateIssueStatus).toHaveBeenCalledTimes(2);
    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "Reject #42" })).not.toBeInTheDocument();
    });
  });

  it("edits markdown descriptions", async () => {
    const user = userEvent.setup();
    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Edit description" }));
    await user.clear(screen.getByRole("textbox", { name: "Description" }));
    await user.type(screen.getByRole("textbox", { name: "Description" }), "Updated **description**");
    await user.click(screen.getByRole("tab", { name: "Preview" }));
    expect(screen.getByText("description")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(api.updateIssueDescription).toHaveBeenCalledWith(42, "Updated **description**");
    expect(
      await screen.findByText((_, element) =>
        element?.tagName === "P" && element.textContent === "Updated description",
      ),
    ).toBeInTheDocument();
  });

  it("keeps markdown edit content when saving fails", async () => {
    const user = userEvent.setup();
    api.updateIssueDescription.mockRejectedValueOnce(new Error("failed to update"));
    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Edit description" }));
    await user.clear(screen.getByRole("textbox", { name: "Description" }));
    await user.type(screen.getByRole("textbox", { name: "Description" }), "Draft after failure");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("failed to update")).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Description" })).toHaveValue("Draft after failure");
  });

  it("saves an empty markdown description", async () => {
    const user = userEvent.setup();
    api.updateIssueDescription.mockResolvedValueOnce({ ...issue, description: "" });
    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Edit description" }));
    await user.clear(screen.getByRole("textbox", { name: "Description" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(api.updateIssueDescription).toHaveBeenCalledWith(42, "");
    expect(await screen.findByText("No description")).toBeInTheDocument();
  });

  it("cancels markdown description edits", async () => {
    const user = userEvent.setup();
    renderIssueDetail();

    await user.click(await screen.findByRole("button", { name: "Edit description" }));
    await user.clear(screen.getByRole("textbox", { name: "Description" }));
    await user.type(screen.getByRole("textbox", { name: "Description" }), "Discard this");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(screen.queryByDisplayValue("Discard this")).not.toBeInTheDocument();
    expect(screen.getByText("Show details in tabs.")).toBeInTheDocument();
  });

  it("loads comments when the comments tab is selected", async () => {
    renderIssueDetail("/issues/42?tab=comments");

    expect(await screen.findByText("Looks good from design review.")).toBeInTheDocument();
    expect(screen.getByText("Please cover the empty state before resolving.")).toBeInTheDocument();
    expect(screen.getByText("Open")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Runs" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Refine issue detail" })).not.toBeInTheDocument();
    expect(api.fetchChangeRequests).toHaveBeenCalledWith(42, 100, { silent: true });
    expect(api.fetchComments).toHaveBeenCalledWith(42, undefined, 20, { silent: true });
  });

  it("shows a change request load error without skipping comments", async () => {
    api.fetchChangeRequests.mockRejectedValueOnce(new Error("failed to load requests"));

    renderIssueDetail("/issues/42?tab=comments");

    expect(await screen.findByText("failed to load requests")).toBeTruthy();
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

const changeRequest = {
  id: 9,
  issueId: 42,
  author: "reviewer",
  body: "Please cover the empty state",
  status: "open",
  createdAt: "2026-06-21T02:30:00.000Z",
  updatedAt: "2026-06-21T02:30:00.000Z",
  resolvedAt: null,
  resolvedByRunId: null,
  resultCommentId: null,
};

const changeRequestsResponse: ChangeRequestListResponse = {
  data: [
    {
      id: 9,
      issueId: 42,
      author: "reviewer",
      body: "Please cover the empty state before resolving.",
      status: "open",
      createdAt: "2026-06-21T01:30:00.000Z",
      updatedAt: "2026-06-21T01:30:00.000Z",
      resolvedAt: null,
      resolvedByRunId: null,
      resultCommentId: null,
    },
  ],
  meta: {
    limit: 100,
    status: "",
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
