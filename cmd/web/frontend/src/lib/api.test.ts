import { beforeEach, describe, expect, it, vi } from "vitest";
import { toastStore } from "@/lib/toast";
import { fetchSummary } from "./api";

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

vi.mock("@/lib/generated/orchestrator", () => ({
  getApiV1IssueIdentifier: vi.fn(),
  getApiV1IssueIdentifierRunsRunIdConversations: vi.fn(),
}));

describe("api toast handling", () => {
  beforeEach(() => {
    toastStore.clear();
    issueTracker.getApiV1Summary.mockReset();
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
});
