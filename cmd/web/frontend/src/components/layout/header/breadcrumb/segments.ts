export type BreadcrumbSegment = {
  label: string;
  href?: string;
};

export function breadcrumbSegmentsFromPathname(pathname: string): BreadcrumbSegment[] {
  const segments = pathname.split(/[?#]/, 1)[0].split("/").filter(Boolean);

  if (matchesRoute(segments, ["dashboard"])) {
    return [{ label: "Dashboard" }];
  }

  if (matchesRoute(segments, ["dashboard", "table"])) {
    return [
      { label: "Dashboard", href: "/dashboard" },
      { label: "Table" },
    ];
  }

  if (matchesRoute(segments, ["dashboard", "stats"])) {
    return [
      { label: "Dashboard", href: "/dashboard" },
      { label: "Stats" },
    ];
  }

  if (matchesRoute(segments, ["projects", ":projectKey", "issues"])) {
    const projectKey = segments[1];
    return [
      { label: decodePathSegment(projectKey), href: `/projects/${projectKey}/issues` },
      { label: "Issues" },
    ];
  }

  if (matchesRoute(segments, ["projects", ":projectKey", "settings"])) {
    const projectKey = segments[1];
    return [
      { label: decodePathSegment(projectKey), href: `/projects/${projectKey}/issues` },
      { label: "Settings" },
    ];
  }

  if (matchesRoute(segments, ["projects", ":projectKey", "table"])) {
    const projectKey = segments[1];
    return [
      { label: decodePathSegment(projectKey), href: `/projects/${projectKey}/issues` },
      { label: "Table" },
    ];
  }

  if (
    matchesRoute(segments, ["issues", ":issueID"]) &&
    isIssueIDSegment(segments[1])
  ) {
    const issueID = segments[1];
    return [
      { label: "Dashboard", href: "/dashboard" },
      { label: `#${decodePathSegment(issueID)}` },
    ];
  }

  if (
    (
      matchesRoute(segments, ["issues", ":issueID", "conversations"]) ||
      matchesRoute(segments, ["issues", ":issueID", "runs", ":runID", "conversations"])
    ) &&
    isIssueIDSegment(segments[1])
  ) {
    const issueID = segments[1];
    return [
      { label: "Dashboard", href: "/dashboard" },
      { label: `#${decodePathSegment(issueID)}`, href: `/issues/${issueID}` },
      { label: "Conversation" },
    ];
  }

  if (matchesRoute(segments, ["settings"])) {
    return [{ label: "Settings" }];
  }

  return [];
}

function matchesRoute(segments: string[], route: string[]): boolean {
  return (
    segments.length === route.length &&
    route.every((routeSegment, index) => routeSegment.startsWith(":") || routeSegment === segments[index])
  );
}

function decodePathSegment(segment: string): string {
  try {
    return decodeURIComponent(segment);
  } catch {
    return segment;
  }
}

function isIssueIDSegment(segment: string): boolean {
  return /^\d+$/.test(segment);
}
