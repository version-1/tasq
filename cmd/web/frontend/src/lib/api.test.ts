import { beforeEach, describe, expect, it, vi } from "vitest";
import { toastStore } from "@/lib/toast";
import { fetchProjectWorkflow, fetchSummary } from "./api";

const issueTracker = vi.hoisted(() => ({
  deleteApiV1ProjectsId: vi.fn(),
  getApiV1Summary: vi.fn(),
  postApiV1Issues: vi.fn(),
  getApiV1IssuesId: vi.fn(),
  getApiV1IssuesIssueIdComments: vi.fn(),
  getApiV1Projects: vi.fn(),
  getApiV1ProjectsIdWorkflow: vi.fn(),
  postApiV1Projects: vi.fn(),
  patchApiV1IssuesId: vi.fn(),
}));

vi.mock("@/lib/generated/issue-tracker", () => issueTracker);

vi.mock("@/lib/generated/orchestrator", () => ({
  getApiV1IssueIdentifier: vi.fn(),
  getApiV1IssueIdentifierRunsRunIdConversations: vi.fn(),
}));

describe("api toast handling", () => {
  beforeEach(() => {
    toastStore.clear();
    issueTracker.getApiV1Summary.mockReset();
    issueTracker.getApiV1ProjectsIdWorkflow.mockReset();
  });

  it("shows an i18n error toast for failed responses by default", async () => {
    issueTracker.getApiV1Summary.mockResolvedValue({
      status: 500,
      data: {
        error: {
          code: "summary.get.internal_error",
          message: "summary failed",
        },
        meta: {},
      },
    });

    await expect(fetchSummary()).rejects.toMatchObject({
      code: "summary.get.internal_error",
      message: "summary failed",
    });

    expect(toastStore.getSnapshot()).toMatchObject([
      {
        message: "Could not load the summary",
        type: "error",
      },
    ]);
  });

  it("does not show a toast when silent is true", async () => {
    issueTracker.getApiV1Summary.mockResolvedValue({
      status: 500,
      data: {
        error: {
          code: "summary.get.internal_error",
          message: "summary failed",
        },
        meta: {},
      },
    });

    await expect(fetchSummary({ silent: true })).rejects.toMatchObject({
      code: "summary.get.internal_error",
    });

    expect(toastStore.getSnapshot()).toEqual([]);
  });

  it("unwraps project workflow responses", async () => {
    issueTracker.getApiV1ProjectsIdWorkflow.mockResolvedValue({
      status: 200,
      data: {
        data: {
          projectId: 1,
          frontmatter: { tasq: { tracker: { default_status: "ready" } } },
          body: "## Workflow",
          checksum: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
          createdAt: "2026-06-01T00:00:00.000Z",
          updatedAt: "2026-06-22T00:00:00.000Z",
        },
        meta: {},
      },
    });

    await expect(fetchProjectWorkflow(1)).resolves.toMatchObject({
      projectId: 1,
      body: "## Workflow",
      frontmatter: { tasq: { tracker: { default_status: "ready" } } },
    });
  });
});
