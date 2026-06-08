import { beforeEach, describe, expect, it, vi } from "vitest";
import { clearToasts, getToastsSnapshot } from "@/lib/toast";
import { createIssue, fetchSummary } from "./api";

const issueTracker = vi.hoisted(() => ({
  getApiV1Summary: vi.fn(),
  postApiV1Issues: vi.fn(),
  getApiV1IssuesId: vi.fn(),
  getApiV1IssuesIssueIdComments: vi.fn(),
  getApiV1Projects: vi.fn(),
  postApiV1Projects: vi.fn(),
  patchApiV1IssuesId: vi.fn(),
}));

vi.mock("@/lib/generated/issue-tracker", () => issueTracker);

describe("api error toast handling", () => {
  beforeEach(() => {
    clearToasts();
    vi.clearAllMocks();
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

    expect(getToastsSnapshot()).toMatchObject([
      {
        message: "Could not load the summary",
        type: "error",
      },
    ]);
  });

  it("falls back to the server message when the i18n error key is missing", async () => {
    issueTracker.getApiV1Summary.mockResolvedValue({
      status: 404,
      data: {
        error: {
          code: "summary.get.unknown",
          message: "server fallback message",
        },
        meta: {},
      },
    });

    await expect(fetchSummary()).rejects.toMatchObject({
      code: "summary.get.unknown",
      message: "server fallback message",
    });

    expect(getToastsSnapshot()).toMatchObject([
      {
        message: "server fallback message",
        type: "error",
      },
    ]);
  });

  it("does not show a toast when silent is true", async () => {
    issueTracker.postApiV1Issues.mockResolvedValue({
      status: 400,
      data: {
        error: {
          code: "issues.create.invalid_input",
          message: "issue input is invalid",
        },
        meta: {},
      },
    });

    await expect(
      createIssue(
        { projectId: 1, title: "", description: "", status: "backlog", priority: "normal" },
        { silent: true },
      ),
    ).rejects.toMatchObject({
      code: "issues.create.invalid_input",
    });

    expect(getToastsSnapshot()).toEqual([]);
  });
});
