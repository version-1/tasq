import { useQuery } from "@tanstack/react-query";
import { fetchOrchestratorIssueRuntime } from "@/lib/api";

const issueThreadIDQueryKey = (issueID: number) => ["issues", issueID, "thread-id"] as const;

export function useIssueThreadID(issueID: number, enabled: boolean) {
  const query = useQuery({
    queryKey: issueThreadIDQueryKey(issueID),
    queryFn: async (): Promise<string | null> => {
      try {
        const runtime = await fetchOrchestratorIssueRuntime(issueID, { silent: true });
        return threadIDFromRuns(runtime.runs);
      } catch {
        return null;
      }
    },
    enabled,
    gcTime: Infinity,
    staleTime: Infinity,
  });

  return {
    isThreadIDLoading: query.isLoading,
    threadID: query.data ?? null,
  };
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
