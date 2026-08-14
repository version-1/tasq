export type BuildInfo = {
  version: string;
  commit: string | null;
};

const PLACEHOLDER_PREFIX = "__TASQ_";

function readMetadata(name: string): string | null {
  const value = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)?.content;
  if (!value || value.startsWith(PLACEHOLDER_PREFIX)) {
    return null;
  }
  return value;
}

function formatVersion(version: string): string {
  if (version === "dev" || version.startsWith("v")) {
    return version;
  }
  return `v${version}`;
}

export function getBuildInfo(): BuildInfo {
  const commit = readMetadata("tasq-commit");
  return {
    version: formatVersion(readMetadata("tasq-version") ?? "dev"),
    commit: commit === "unknown" ? null : commit,
  };
}
