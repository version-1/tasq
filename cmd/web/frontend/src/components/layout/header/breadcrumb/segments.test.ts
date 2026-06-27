import { describe, expect, it } from "vitest";

import { breadcrumbSegmentsFromPathname } from "./segments";

describe("breadcrumbSegmentsFromPathname", () => {
  it.each([
    {
      pathname: "/dashboard",
      segments: [{ label: "Dashboard" }],
    },
    {
      pathname: "/dashboard/table",
      segments: [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Table" },
      ],
    },
    {
      pathname: "/dashboard/stats",
      segments: [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Stats" },
      ],
    },
    {
      pathname: "/projects/TASQ/issues",
      segments: [
        { label: "TASQ", href: "/projects/TASQ/issues" },
        { label: "Issues" },
      ],
    },
    {
      pathname: "/projects/TASQ/settings",
      segments: [
        { label: "TASQ", href: "/projects/TASQ/issues" },
        { label: "Settings" },
      ],
    },
    {
      pathname: "/projects/TASQ/table",
      segments: [
        { label: "TASQ", href: "/projects/TASQ/issues" },
        { label: "Table" },
      ],
    },
    {
      pathname: "/issues/25",
      segments: [
        { label: "Dashboard", href: "/dashboard" },
        { label: "#25" },
      ],
    },
    {
      pathname: "/issues/25/conversations",
      segments: [
        { label: "Dashboard", href: "/dashboard" },
        { label: "#25", href: "/issues/25" },
        { label: "Conversation" },
      ],
    },
    {
      pathname: "/issues/25/runs/7/conversations",
      segments: [
        { label: "Dashboard", href: "/dashboard" },
        { label: "#25", href: "/issues/25" },
        { label: "Conversation" },
      ],
    },
    {
      pathname: "/settings",
      segments: [{ label: "Settings" }],
    },
  ])("maps $pathname to breadcrumb segments", ({ pathname, segments }) => {
    expect(breadcrumbSegmentsFromPathname(pathname)).toEqual(segments);
  });

  it("decodes labels and keeps hrefs URL encoded", () => {
    expect(breadcrumbSegmentsFromPathname("/projects/My%20Project/issues")).toEqual([
      { label: "My Project", href: "/projects/My%20Project/issues" },
      { label: "Issues" },
    ]);
  });

  it("ignores trailing slashes, query strings, and hashes", () => {
    expect(breadcrumbSegmentsFromPathname("/issues/25/conversations/?tab=notes#latest")).toEqual([
      { label: "Dashboard", href: "/dashboard" },
      { label: "#25", href: "/issues/25" },
      { label: "Conversation" },
    ]);
  });

  it("returns no segments for unmapped paths", () => {
    expect(breadcrumbSegmentsFromPathname("/unknown")).toEqual([]);
    expect(breadcrumbSegmentsFromPathname("/issues")).toEqual([]);
    expect(breadcrumbSegmentsFromPathname("/issues/table")).toEqual([]);
  });
});
