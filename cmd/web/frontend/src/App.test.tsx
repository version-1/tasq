import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { App } from "./App";

vi.mock("./app/dashboard/page", () => ({ default: () => <div /> }));
vi.mock("./app/dashboard/stats/page", () => ({ default: () => <div /> }));
vi.mock("./app/dashboard/table/page", () => ({ default: () => <div /> }));
vi.mock("./app/not-found/page", () => ({ NotFoundRoute: () => <div /> }));
vi.mock("./app/issues/[id]/conversations/page", () => ({ default: () => <div /> }));
vi.mock("./app/issues/[id]/page", () => ({ default: () => <div /> }));
vi.mock("./app/issues/page", () => ({ default: () => <div /> }));
vi.mock("./app/issues/table/page", () => ({ default: () => <div /> }));
vi.mock("./app/projects/[projectKey]/artifacts/page", () => ({ default: () => <div /> }));
vi.mock("./app/projects/[projectKey]/settings/page", () => ({ default: () => <div /> }));
vi.mock("./app/settings/page", () => ({ default: () => <div /> }));
vi.mock("./components/layout", () => ({ Layout: ({ children }: { children: ReactNode }) => <>{children}</> }));
vi.mock("./components/layout/default", () => ({ DefaultLayout: ({ children }: { children: ReactNode }) => <>{children}</> }));
vi.mock("./components/layout/issue", () => ({ IssueDetailLayout: ({ children }: { children: ReactNode }) => <>{children}</> }));
vi.mock("./components/layout/project", () => ({ ProjectLayout: ({ children }: { children: ReactNode }) => <>{children}</> }));
vi.mock("./context/tabs", () => ({ TabsProvider: ({ activeKey, children }: { activeKey: string; children: ReactNode }) => <div data-active-key={activeKey} data-testid="tabs">{children}</div> }));

describe("project route tabs", () => {
  it.each([
    ["/projects/tasq/issues", "issues"],
    ["/projects/tasq/artifacts", "artifacts"],
    ["/projects/tasq/artifacts/doc-1", "artifacts"],
  ])("sets %s active tab to %s", (path, activeKey) => {
    render(<MemoryRouter initialEntries={[path]}><App /></MemoryRouter>);
    expect(screen.getByTestId("tabs")).toHaveAttribute("data-active-key", activeKey);
  });
});
