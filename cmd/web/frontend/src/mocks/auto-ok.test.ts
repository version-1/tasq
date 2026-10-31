import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { setupServer } from "msw/node";
import type { Issue } from "@/lib/types";
import { handlers } from "./handlers";
import { updateIssue } from "./state";

const server = setupServer(...handlers);
const base = `${window.location.origin}/tracker/api/v1`;
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());

async function request(path: string, method = "GET", body?: unknown) {
  return fetch(`${base}${path}`, { method, headers: { "Content-Type": "application/json" }, body: body === undefined ? undefined : JSON.stringify(body) });
}

async function issueRequest(path: string, method = "GET", body?: unknown): Promise<Issue> {
  const response = await request(path, method, body);
  expect(response.status).toBeLessThan(400);
  return (await response.json()).data;
}

async function createBlockedIssue(dependencyIDs: number[] = []) {
  return issueRequest("/issues", "POST", { projectId: 1, title: "Mock Auto OK", status: "blocked", dependency_ids: dependencyIDs });
}

describe("Auto OK mock HTTP workflow", () => {
  it("resumes on enable, stops at five accepted replies, and resets on re-enable", async () => {
    let issue = await createBlockedIssue();
    expect(issue.autoOk).toBe(false);
    for (let count = 1; count <= 5; count++) {
      issue = await issueRequest(`/issues/${issue.id}`, "PATCH", count === 1 ? { autoOk: true } : { status: "blocked" });
      expect(issue).toMatchObject({ autoOkCount: count, autoOk: count < 5, autoOkPending: false, status: "in_progress" });
    }
    issue = await issueRequest(`/issues/${issue.id}`, "PATCH", { status: "blocked" });
    expect(issue).toMatchObject({ status: "blocked", autoOkCount: 5, autoOkStopReason: "limit" });
    const comments = await (await request(`/issues/${issue.id}/comments`)).json();
    expect(comments.data.filter((comment: { author: string }) => comment.author === "tasq-auto-ok")).toHaveLength(5);
    issue = await issueRequest(`/issues/${issue.id}`, "PATCH", { autoOk: true });
    expect(issue).toMatchObject({ autoOk: true, autoOkCount: 1, autoOkStopReason: "" });
  });

  it("waits for dependencies and supports cancelling an unsent reply", async () => {
    const dependency = await issueRequest("/issues", "POST", { projectId: 1, title: "Dependency" });
    let issue = await createBlockedIssue([dependency.id]);
    issue = await issueRequest(`/issues/${issue.id}`, "PATCH", { autoOk: true });
    expect(issue).toMatchObject({ autoOkPending: true, autoOkCount: 0, status: "ready" });
    issue = await issueRequest(`/issues/${issue.id}`, "PATCH", { autoOk: false });
    expect(issue).toMatchObject({ autoOkPending: false, autoOkCount: 0, status: "blocked" });
    await issueRequest(`/issues/${issue.id}`, "PATCH", { autoOk: true });
    await issueRequest(`/issues/${dependency.id}`, "PATCH", { status: "done" });
    await request("/summary");
    issue = await issueRequest(`/issues/${issue.id}`);
    expect(issue).toMatchObject({ autoOkPending: false, autoOkCount: 1, status: "in_progress" });
  });

  it("records failed delivery and ignores obsolete generations", async () => {
    let issue = await createBlockedIssue();
    issue = updateIssue(issue.id, { autoOk: true })!;
    const generation = issue.autoOkGeneration;
    await issueRequest(`/issues/${issue.id}/auto-ok-result`, "POST", { generation, phase: "claim" });
    await request("/summary");
    expect((await issueRequest(`/issues/${issue.id}`)).autoOkCount).toBe(0);
    issue = await issueRequest(`/issues/${issue.id}/auto-ok-result`, "POST", { generation, error: "Mock resume failed" });
    expect(issue).toMatchObject({ autoOk: false, autoOkCount: 0, status: "blocked", autoOkStopReason: "error" });
    const comments = await (await request(`/issues/${issue.id}/comments`)).json();
    expect(comments.data[0].body).toContain("Mock resume failed");
    issue = await issueRequest(`/issues/${issue.id}`, "PATCH", { autoOk: true });
    issue = await issueRequest(`/issues/${issue.id}/auto-ok-result`, "POST", { generation });
    expect(issue).toMatchObject({ autoOk: true, autoOkCount: 1 });
  });

  it("preserves claimed replies across lifecycle updates and rejects invalid requests", async () => {
    let issue = await createBlockedIssue();
    issue = updateIssue(issue.id, { autoOk: true })!;
    const generation = issue.autoOkGeneration;
    await issueRequest(`/issues/${issue.id}/auto-ok-result`, "POST", { generation, phase: "claim" });
    updateIssue(issue.id, { status: "blocked" });
    issue = await issueRequest(`/issues/${issue.id}/auto-ok-result`, "POST", { generation });
    expect(issue).toMatchObject({ autoOkCount: 1, autoOkPending: true, status: "ready" });
    expect((await request(`/issues/${issue.id}/auto-ok-result`, "POST", { generation, phase: "cliam" })).status).toBe(400);
    expect((await request("/issues/999999/auto-ok-result", "POST", { generation })).status).toBe(404);
  });
});
