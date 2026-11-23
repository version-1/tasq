import type { ReactNode } from "react";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http } from "msw";
import { setupServer } from "msw/node";
import { MemoryRouter } from "react-router-dom";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import { App } from "./App";
import "@/lib/i18n";
import { updateProjectDocumentArtifact } from "@/lib/api";
import { handlers } from "@/mocks/handlers";
import { projectFixtures } from "@/mocks/fixtures/projects";

vi.mock("./app/dashboard/page", () => ({ default: () => <div /> }));
vi.mock("./app/dashboard/stats/page", () => ({ default: () => <div /> }));
vi.mock("./app/dashboard/table/page", () => ({ default: () => <div /> }));
vi.mock("./app/not-found/page", () => ({ NotFoundRoute: () => <div /> }));
vi.mock("./app/issues/[id]/conversations/page", () => ({ default: () => <div /> }));
vi.mock("./app/issues/[id]/page", () => ({ default: () => <div /> }));
vi.mock("./app/issues/page", () => ({ default: () => <div /> }));
vi.mock("./app/issues/table/page", () => ({ default: () => <div /> }));
vi.mock("./app/projects/[projectKey]/settings/page", () => ({ default: () => <div /> }));
vi.mock("./app/settings/page", () => ({ default: () => <div /> }));
vi.mock("./components/layout", () => ({
  Layout: ({ children }: { children: ReactNode }) => <>{children}</>,
  useLayoutShellData: () => ({ activeProject: projectFixtures[0], loadState: { kind: "ready" } }),
}));
vi.mock("./components/layout/default", () => ({ DefaultLayout: ({ children }: { children: ReactNode }) => <>{children}</> }));
vi.mock("./components/layout/issue", () => ({ IssueDetailLayout: ({ children }: { children: ReactNode }) => <>{children}</> }));
vi.mock("./components/layout/project", () => ({ ProjectLayout: ({ children }: { children: ReactNode }) => <>{children}</> }));
vi.mock("./context/tabs", () => ({ TabsProvider: ({ activeKey, children }: { activeKey: string; children: ReactNode }) => <div data-active-key={activeKey} data-testid="tabs">{children}</div> }));

const server = setupServer(...handlers);
const listPath = "/projects/tasq/artifacts";

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function renderPath(path: string) {
  return render(<MemoryRouter initialEntries={[path]}><App /></MemoryRouter>);
}

describe("project artifact routes", () => {
  it.each([
    ["/projects/tasq/issues", "issues"],
    [listPath, "artifacts"],
  ])("sets %s active tab to %s", (path, activeKey) => {
    renderPath(path);
    expect(screen.getByTestId("tabs")).toHaveAttribute("data-active-key", activeKey);
  });

  it("opens a full-page document from the list and returns via the breadcrumb", async () => {
    const user = userEvent.setup();
    renderPath(listPath);

    await user.click(await screen.findByRole("link", { name: /Acceptance criteria/ }));
    expect(await screen.findByText("Preserve issue IDs")).toBeInTheDocument();
    expect(screen.queryByTestId("tabs")).not.toBeInTheDocument();
    const navigation = screen.getAllByRole("navigation");
    expect(navigation).toHaveLength(1);

    await user.click(within(navigation[0]).getByRole("link", { name: "Artifacts" }));
    expect(await screen.findByRole("link", { name: /Acceptance criteria/ })).toBeInTheDocument();
  });

  it("saves title and Markdown from a direct URL and reads the changes on a fresh load", async () => {
    const user = userEvent.setup();
    const path = `${listPath}/doc-2`;
    const page = renderPath(path);

    expect(await screen.findByText("All checks passed.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Artifacts" })).toHaveAttribute("href", listPath);
    await user.click(screen.getByRole("button", { name: "Edit document" }));
    const title = screen.getByRole("textbox", { name: "Title" });
    const body = screen.getByRole("textbox", { name: "Markdown body" });
    await user.clear(title);
    await user.type(title, "Updated report");
    await user.clear(body);
    await user.type(body, "**Verified**");

    server.use(http.patch("/tracker/api/v1/projects/1/artifacts/documents/doc-2", () => Response.error()));
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(title).toHaveValue("Updated report");
    expect(body).toHaveValue("**Verified**");

    server.resetHandlers();
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByRole("heading", { name: "Updated report", level: 1 })).toBeInTheDocument();
    page.unmount();

    renderPath(path);
    expect(await screen.findByRole("heading", { name: "Updated report", level: 1 })).toBeInTheDocument();
    expect(screen.getByText("Verified").tagName).toBe("STRONG");
  });

  it("preserves a title saved by another client when editing only the body", async () => {
    const user = userEvent.setup();
    renderPath(`${listPath}/doc-1`);

    await user.click(await screen.findByRole("button", { name: "Edit document" }));
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    await updateProjectDocumentArtifact(1, "doc-1", { title: "Concurrent title" });
    const body = screen.getByRole("textbox", { name: "Markdown body" });
    await user.clear(body);
    await user.type(body, "Revised criteria");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByRole("heading", { name: "Concurrent title", level: 1 })).toBeInTheDocument();
    expect(screen.getByText("Revised criteria")).toBeInTheDocument();
  });
});
