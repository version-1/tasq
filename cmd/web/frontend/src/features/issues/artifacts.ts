import type { Artifact } from "@/lib/types";

export function pullRequestArtifact(artifacts: readonly Artifact[]): Artifact | undefined {
  return artifacts.find(
    (artifact) => artifact.type === "pull_request" && artifact.data_type === "url",
  );
}
