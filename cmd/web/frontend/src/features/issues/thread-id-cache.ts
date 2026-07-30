import { fetchOrchestratorIssueRuntime } from "@/lib/api";

type ThreadIDLoader = (issueID: number) => Promise<string | null>;

class IssueThreadIDCache {
  private readonly entries = new Map<number, Promise<string | null>>();

  load(issueID: number, loader: ThreadIDLoader): Promise<string | null> {
    const cachedThreadID = this.entries.get(issueID);
    if (cachedThreadID) {
      return cachedThreadID;
    }

    const threadID = loader(issueID);
    this.entries.set(issueID, threadID);
    return threadID;
  }

  clear(): void {
    this.entries.clear();
  }
}

const issueThreadIDCache = new IssueThreadIDCache();

export function loadIssueThreadID(issueID: number): Promise<string | null> {
  return issueThreadIDCache.load(issueID, fetchIssueThreadID);
}

async function fetchIssueThreadID(issueID: number): Promise<string | null> {
  try {
    const runtime = await fetchOrchestratorIssueRuntime(issueID, { silent: true });
    return threadIDFromRuns(runtime.runs);
  } catch {
    return null;
  }
}

function threadIDFromRuns(runs: readonly { readonly thread_id?: string }[]): string | null {
  for (const run of runs) {
    const threadID = run.thread_id?.trim();
    if (threadID) {
      return threadID;
    }
  }
  return null;
}

/** @internal Keeps the module-scoped cache isolated between unit tests. */
export function clearIssueThreadIDCacheForTest(): void {
  issueThreadIDCache.clear();
}
