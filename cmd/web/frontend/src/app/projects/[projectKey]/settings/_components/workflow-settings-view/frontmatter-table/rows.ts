import type { ProjectWorkflowFrontmatter } from "@/lib/types";

export type FrontmatterRow = {
  depth: number;
  id: string;
  kind: "branch" | "leaf";
  key: string;
  value: string;
};

export function toFrontmatterRows(value: ProjectWorkflowFrontmatter): FrontmatterRow[] {
  return Object.entries(value).flatMap(([key, nestedValue]) =>
    flattenFrontmatterValue({
      depth: 0,
      key,
      path: key,
      value: nestedValue,
    }),
  );
}

function flattenFrontmatterValue({
  depth,
  key,
  path,
  value,
}: {
  depth: number;
  key: string;
  path: string;
  value: unknown;
}): FrontmatterRow[] {
  if (isRecord(value)) {
    const entries = Object.entries(value);
    const row: FrontmatterRow = {
      depth,
      id: path,
      key,
      kind: entries.length > 0 ? "branch" : "leaf",
      value: entries.length > 0 ? "{...}" : "{}",
    };
    return [
      row,
      ...entries.flatMap(([nestedKey, nestedValue]) =>
        flattenFrontmatterValue({
          depth: depth + 1,
          key: nestedKey,
          path: `${path}.${nestedKey}`,
          value: nestedValue,
        }),
      ),
    ];
  }

  if (Array.isArray(value)) {
    const row: FrontmatterRow = {
      depth,
      id: path,
      key,
      kind: value.length > 0 ? "branch" : "leaf",
      value: value.length > 0 ? "[...]" : "[]",
    };
    return [
      row,
      ...value.flatMap((item, index) =>
        flattenFrontmatterValue({
          depth: depth + 1,
          key: `[${index}]`,
          path: `${path}[${index}]`,
          value: item,
        }),
      ),
    ];
  }

  return [{
    depth,
    id: path,
    key,
    kind: "leaf",
    value: formatScalar(value),
  }];
}

function isRecord(value: unknown): value is ProjectWorkflowFrontmatter {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function formatScalar(value: unknown): string {
  if (value === null) {
    return "null";
  }
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  return JSON.stringify(value);
}
