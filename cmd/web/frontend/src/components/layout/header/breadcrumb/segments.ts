export type BreadcrumbSegment = {
  label: string;
  href?: string;
};

export function breadcrumbSegmentsFromPathname(pathname: string): BreadcrumbSegment[] {
  const segments = pathname.split(/[?#]/, 1)[0].split("/").filter(Boolean);

  if (matchesRoute(segments, ["dashboard"])) {
    return [{ label: "Dashboard" }];
  }

  if (matchesRoute(segments, ["issues"])) {
    return [{ label: "Issues" }];
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

  if (matchesRoute(segments, ["issues", ":issueID"])) {
    const issueID = segments[1];
    return [
      { label: "Issues", href: "/issues" },
      { label: `#${decodePathSegment(issueID)}` },
    ];
  }

  if (
    matchesRoute(segments, ["issues", ":issueID", "conversations"]) ||
    matchesRoute(segments, ["issues", ":issueID", "runs", ":runID", "conversations"])
  ) {
    const issueID = segments[1];
    return [
      { label: "Issues", href: "/issues" },
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
