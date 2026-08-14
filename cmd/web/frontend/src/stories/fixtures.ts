import type {
  Comment,
  Issue,
  IssueStatus,
  IssueSummary,
  OrchestratorIssueRun,
  QueueStatus,
  Summary,
} from "@/lib/types";
import type { LayoutShellData } from "@/components/layout";
import { issueStatuses } from "@/lib/types";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { commentFixtures } from "@/mocks/fixtures/comments";
import { projectFixtures } from "@/mocks/fixtures/projects";

export const noop = () => undefined;
export const asyncNoop = async () => undefined;

export function storyIssue(index = 0): IssueSummary {
  const issue = issueFixtures[index];
  return {
    ...issue,
    queueStatus: storyQueueStatusForIssue(issue),
    stats: {
      commentCount: index + 1,
    },
  };
}

export function storyQueueStatusForIssue(issue: Issue): QueueStatus {
  if (issue.status === "backlog") {
    return "backlog";
  }
  if (issue.status === "ready") {
    return issue.dependency_ids.some(hasActiveDependency) ? "pending" : "queued";
  }
  if (issue.status === "in_progress") {
    return "processing";
  }
  if (issue.status === "done") {
    return "completed";
  }
  return "inactive";
}

function hasActiveDependency(dependencyID: number): boolean {
  const dependency = issueFixtures.find((issue) => issue.id === dependencyID);
  return dependency !== undefined && !isSatisfiedDependencyStatus(dependency.status);
}

function isSatisfiedDependencyStatus(status: IssueStatus): boolean {
  return (
    status === "done" ||
    status === "cancelled" ||
    status === "duplicate"
  );
}

export const storySummary: Summary = {
  generatedAt: "2026-06-16T00:00:00.000Z",
  columns: issueStatuses.map((status) => ({
    status,
    title: status,
    issues: issueFixtures
      .filter((issue) => issue.status === status)
      .map((issue, index) => ({
        ...issue,
        queueStatus: storyQueueStatusForIssue(issue),
        stats: {
          commentCount: index + 1,
        },
      })),
  })),
};

export const storyComments: Comment[] = commentFixtures;

export const storyRuns: OrchestratorIssueRun[] = [
  {
    run_id: "run-1-latest",
    status: "running",
    attempt: 1,
    thread_id: "thread-story-1",
    created_at: "2026-06-16T00:00:00.000Z",
    updated_at: "2026-06-16T00:05:00.000Z",
  },
];

export const storyShellData: LayoutShellData = {
  activePage: "issues",
  activeProject: projectFixtures[0],
  addIssueError: "",
  addIssueInitialStatus: "backlog" as IssueStatus,
  deleteProjectError: "",
  isIssueDetailPage: false,
  isDeletingProject: false,
  isMovingRejectedIssue: false,
  isMovingResolvedIssue: false,
  isProjectIssueScope: false,
  issues: storySummary.columns.flatMap((column) => column.issues),
  layoutData: {
    summary: storySummary,
    issues: storySummary.columns.flatMap((column) => column.issues),
    selectedIssue: storyIssue(0),
    refreshIntervalMs: 3000,
    language: "ja",
    onRefreshIntervalChange: noop,
    onLanguageChange: noop,
    onSelectIssue: noop,
    onAddIssue: noop,
    onRejectIssue: noop,
    onRejectShortcut: asyncNoop,
    onResolveIssue: noop,
    onStatusChange: asyncNoop,
  },
  loadState: {
    kind: "ready",
    projects: projectFixtures,
    summary: storySummary,
  },
  projects: projectFixtures,
  rejectIssue: null,
  rejectIssueError: "",
  rejectRequestRecovery: { body: "", requestCreated: false },
  resolveIssue: null,
  resolveIssueError: "",
  summary: storySummary,
  title: "Tasq",
  onIssueDetailTitleChange: noop,
  onAddIssue: noop,
  onAddProject: noop,
  onCloseModal: noop,
  onCreateIssue: asyncNoop,
  onDeleteProject: noop,
  onConfirmDeleteProject: asyncNoop,
  onMoveRejectedIssueReady: asyncNoop,
  onMoveResolvedIssueReady: asyncNoop,
};
