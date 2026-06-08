import { http, HttpResponse } from "msw";
import type {
  CreateIssueInput,
  CreateProjectInput,
  ErrorResponse as IssueTrackerErrorResponse,
  IssueStatesInput,
  IssueStatus,
  UpdateIssueInput,
} from "@/lib/generated/issue-tracker";
import type { ErrorResponse as OrchestratorErrorResponse } from "@/lib/generated/orchestrator";
import {
  buildOrchestratorIssueRuntime,
  buildSummary,
  createComment,
  createIssue,
  createProject,
  getOrchestratorConversation,
  getIssue,
  listComments,
  listIssueStates,
  listIssues,
  listProjects,
  updateIssue,
} from "./state";

const apiBase = "/tracker/api/v1";
const orchestratorBase = "/orchestrator/api/v1";

export const handlers = [
  http.get(`${apiBase}/health`, () => {
    return jsonOk({ status: "ok" });
  }),

  http.get(`${apiBase}/summary`, () => {
    return jsonOk(buildSummary());
  }),

  http.get(`${apiBase}/projects`, () => {
    return jsonOk(listProjects());
  }),

  http.post(`${apiBase}/projects`, async ({ request }) => {
    const input = await request.json() as CreateProjectInput;
    const project = createProject(input);

    if (!project) {
      return jsonError("projects.create.invalid_input", "Project name, key, and location are required.", 400);
    }

    return jsonOk(project, 201);
  }),

  http.get(`${apiBase}/issues`, ({ request }) => {
    const url = new URL(request.url);
    const projectID = optionalNumber(url.searchParams.get("project_id"));
    const states = parseStates(url.searchParams.get("states"));

    return jsonOk(listIssues({ projectID, states }));
  }),

  http.post(`${apiBase}/issues`, async ({ request }) => {
    const input = await request.json() as CreateIssueInput;
    const issue = createIssue(input);

    if (!issue) {
      return jsonError("issues.create.invalid_input", "Issue title and projectId are required.", 400);
    }

    return jsonOk(issue, 201);
  }),

  http.post(`${apiBase}/issues/states`, async ({ request }) => {
    const input = await request.json() as IssueStatesInput;
    return jsonOk(listIssueStates(input.ids));
  }),

  http.get(`${apiBase}/issues/:id`, ({ params }) => {
    const id = numericParam(params.id);
    const issue = getIssue(id);

    if (!issue) {
      return jsonError("issues.get.not_found", `Issue ${id} was not found.`, 404);
    }

    return jsonOk(issue);
  }),

  http.patch(`${apiBase}/issues/:id`, async ({ params, request }) => {
    const id = numericParam(params.id);
    const input = await request.json() as UpdateIssueInput;
    const issue = updateIssue(id, input);

    if (!issue) {
      return jsonError("issues.update.not_found", `Issue ${id} was not found.`, 404);
    }

    return jsonOk(issue);
  }),

  http.get(`${apiBase}/issues/:issueId/comments`, ({ params, request }) => {
    const issueID = numericParam(params.issueId);
    const url = new URL(request.url);
    const cursor = optionalNumber(url.searchParams.get("cursor")) ?? 0;
    const limit = optionalNumber(url.searchParams.get("limit")) ?? 20;
    const response = listComments(issueID, cursor, limit);

    if (!response) {
      return jsonError("comments.list.issue_not_found", `Issue ${issueID} was not found.`, 404);
    }

    return HttpResponse.json(response);
  }),

  http.post(`${apiBase}/issues/:issueId/comments`, async ({ params, request }) => {
    const issueID = numericParam(params.issueId);
    const input = await request.json() as { author: string; body: string; type?: "progress" | "blocker" | "handoff" | "general" };
    const comment = createComment(issueID, input);

    if (!comment) {
      return jsonError("comments.create.invalid_input", "Comment body, author, and issueId are required.", 400);
    }

    return jsonOk(comment, 201);
  }),

  http.get(`${orchestratorBase}/:issueIdentifier`, ({ params }) => {
    const issueID = issueIDFromIdentifier(params.issueIdentifier);
    if (!issueID) {
      return orchestratorError("invalid_issue_identifier", "issue identifier must use issue-<id>", 400);
    }

    const runtime = buildOrchestratorIssueRuntime(issueID);
    if (!runtime) {
      return orchestratorError("issue_not_found", `Issue ${issueID} was not found.`, 404);
    }

    return HttpResponse.json(runtime);
  }),

  http.get(`${orchestratorBase}/:issueIdentifier/runs/:runId/conversations`, ({ params }) => {
    const issueID = issueIDFromIdentifier(params.issueIdentifier);
    const runID = stringParam(params.runId);
    if (!issueID) {
      return orchestratorError("invalid_issue_identifier", "issue identifier must use issue-<id>", 400);
    }
    if (!runID) {
      return orchestratorError("invalid_run_id", "run id is required", 400);
    }

    const conversation = getOrchestratorConversation(issueID, runID);
    if (!conversation) {
      return orchestratorError("run_not_found", "run not found", 404);
    }

    return HttpResponse.json(conversation);
  }),
];

function jsonOk<T>(data: T, status = 200) {
  return HttpResponse.json({ data, meta: {} }, { status });
}

function jsonError(code: string, message: string, status: 400 | 404 | 500) {
  const body: IssueTrackerErrorResponse = {
    error: { code, message },
    meta: {},
  };

  return HttpResponse.json(body, { status });
}

function orchestratorError(code: string, message: string, status: 400 | 404 | 500) {
  const body: OrchestratorErrorResponse = {
    error: { code, message },
  };

  return HttpResponse.json(body, { status });
}

function parseStates(value: string | null): IssueStatus[] | undefined {
  if (!value) {
    return undefined;
  }

  return value.split(",").filter(Boolean) as IssueStatus[];
}

function optionalNumber(value: string | null): number | undefined {
  if (!value) {
    return undefined;
  }

  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function numericParam(value: string | readonly string[] | undefined): number {
  const rawValue = Array.isArray(value) ? value[0] : value;
  const parsed = Number.parseInt(rawValue ?? "", 10);
  return Number.isFinite(parsed) ? parsed : 0;
}

function stringParam(value: string | readonly string[] | undefined): string {
  return Array.isArray(value) ? value[0] ?? "" : value ?? "";
}

function issueIDFromIdentifier(value: string | readonly string[] | undefined): number | null {
  const identifier = stringParam(value);
  const match = /^issue-([1-9][0-9]*)$/.exec(identifier);
  if (!match) {
    return null;
  }
  return Number.parseInt(match[1], 10);
}
