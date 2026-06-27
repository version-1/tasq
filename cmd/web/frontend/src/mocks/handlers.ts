import { http, HttpResponse } from "msw";
import type {
  CreateIssueInput,
  CreateProjectInput,
  ErrorResponse as IssueTrackerErrorResponse,
  IssueStatesInput,
  IssueStatus,
  Priority,
  UpdateIssueInput,
} from "@/lib/generated/issue-tracker";
import type { ErrorResponse as OrchestratorErrorResponse } from "@/lib/generated/orchestrator";
import {
  buildOrchestratorIssueRuntime,
  buildSummary,
  countIssues,
  createComment,
  createIssue,
  createProject,
  getOrchestratorConversation,
  getIssue,
  getProjectWorkflow,
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

  http.get(`${apiBase}/projects/:id/workflow`, ({ params }) => {
    const id = numericParam(params.id);
    const workflow = getProjectWorkflow(id);

    if (!workflow) {
      return jsonError("projects.workflow.not_found", `Workflow for project ${id} was not found.`, 404);
    }

    return jsonOk(workflow);
  }),

  http.get(`${apiBase}/issues`, ({ request }) => {
    const url = new URL(request.url);
    const projectID = optionalNumber(url.searchParams.get("project_id"));
    const projectIDs = parseNumbers(url.searchParams.get("project_ids"));
    const priorities = parsePriorities(url.searchParams.get("priorities"));
    const states = parseStates(url.searchParams.get("states"));
    const assignee = optionalString(url.searchParams.get("assignee"));
    const limit = optionalNumber(url.searchParams.get("limit"));
    const offset = limit === undefined ? 0 : (optionalNumber(url.searchParams.get("offset")) ?? 0);
    const sortBy = parseIssueSortBy(url.searchParams.get("sort_by"));
    const sortDirection = parseSortDirection(url.searchParams.get("sort_direction"));
    const filters = { assignee, limit, offset, priorities, projectID, projectIDs, sortBy, sortDirection, states };
    const total = countIssues({ assignee, priorities, projectID, projectIDs, states });
    const nextOffset = limit && offset + limit < total ? offset + limit : null;

    return jsonOk(listIssues(filters), 200, { limit: limit ?? 0, offset, total, nextOffset });
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

function jsonOk<T>(data: T, status = 200, meta: Record<string, unknown> = {}) {
  return HttpResponse.json({ data, meta }, { status });
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

function parsePriorities(value: string | null): Priority[] | undefined {
  if (!value) {
    return undefined;
  }

  return value.split(",").filter(Boolean) as Priority[];
}

function parseNumbers(value: string | null): number[] | undefined {
  if (!value) {
    return undefined;
  }

  return value
    .split(",")
    .map((item) => optionalNumber(item))
    .filter((item): item is number => item !== undefined);
}

function optionalNumber(value: string | null): number | undefined {
  if (!value) {
    return undefined;
  }

  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function optionalString(value: string | null): string | undefined {
  const trimmed = value?.trim() ?? "";
  return trimmed === "" ? undefined : trimmed;
}

function parseIssueSortBy(value: string | null): "id" | "priority" | "created_at" | "updated_at" | undefined {
  if (value === "id" || value === "priority" || value === "created_at" || value === "updated_at") {
    return value;
  }
  return undefined;
}

function parseSortDirection(value: string | null): "asc" | "desc" | undefined {
  if (value === "asc" || value === "desc") {
    return value;
  }
  return undefined;
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
