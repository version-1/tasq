import { act, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import "@/lib/i18n";
import { DocumentArtifactsView } from "./index";

const api = vi.hoisted(() => ({
  fetchProjectDocumentArtifact: vi.fn(),
  fetchProjectDocumentArtifacts: vi.fn(),
}));
vi.mock("@/lib/api", () => api);

function renderView(documentID?: string) {
  return render(
    <MemoryRouter>
      <DocumentArtifactsView documentID={documentID} projectID={3} projectKey="tasq" />
    </MemoryRouter>,
  );
}

describe("DocumentArtifactsView", () => {
  beforeEach(() => {
    api.fetchProjectDocumentArtifact.mockReset();
    api.fetchProjectDocumentArtifacts.mockReset();
  });

  it("lists project-owned documents that have no issue link", async () => {
    api.fetchProjectDocumentArtifacts.mockResolvedValue([
      { id: "independent", projectId: 3, type: "reference_document", title: "Project charter", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" },
    ]);

    renderView();

    expect(await screen.findByRole("link", { name: /Project charter/ })).toHaveAttribute(
      "href",
      "/projects/tasq/artifacts/independent",
    );
    expect(api.fetchProjectDocumentArtifact).not.toHaveBeenCalled();
  });

  it("loads the document selected by a deep link", async () => {
    api.fetchProjectDocumentArtifacts.mockResolvedValue([
      { id: "doc-1", projectId: 3, type: "work_report", title: "Report", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" },
    ]);
    api.fetchProjectDocumentArtifact.mockResolvedValue({
      id: "doc-1", projectId: 3, type: "work_report", title: "Report", body: "# Verification", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z",
    });

    renderView("doc-1");

    expect(await screen.findByRole("heading", { name: "Verification" })).toBeInTheDocument();
    expect(api.fetchProjectDocumentArtifact).toHaveBeenCalledWith(3, "doc-1", { silent: true });
  });

  it("shows a not-found error for an invalid deep link", async () => {
    api.fetchProjectDocumentArtifacts.mockResolvedValue([
      { id: "first", projectId: 3, type: "work_report", title: "First", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" },
      { id: "second", projectId: 3, type: "work_report", title: "Second", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" },
    ]);
    api.fetchProjectDocumentArtifact.mockRejectedValue(new Error("Document was not found."));

    renderView("missing");

    await waitFor(() => expect(screen.getByRole("status")).toHaveTextContent("Document was not found."));
  });

  it("does not render a stale document after the deep link changes", async () => {
    let resolveFirst: (value: { id: string; projectId: number; type: "work_report"; title: string; body: string; createdAt: string; updatedAt: string }) => void;
    const first = new Promise<{ id: string; projectId: number; type: "work_report"; title: string; body: string; createdAt: string; updatedAt: string }>((resolve) => { resolveFirst = resolve; });
    api.fetchProjectDocumentArtifacts.mockResolvedValue([
      { id: "first", projectId: 3, type: "work_report", title: "First", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" },
      { id: "second", projectId: 3, type: "work_report", title: "Second", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" },
    ]);
    api.fetchProjectDocumentArtifact.mockImplementation((_projectID: number, id: string) => id === "first" ? first : Promise.resolve({
      id: "second", projectId: 3, type: "work_report", title: "Second", body: "# Second", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z",
    }));

    const { rerender } = render(<MemoryRouter><DocumentArtifactsView documentID="first" projectID={3} projectKey="tasq" /></MemoryRouter>);
    await waitFor(() => expect(api.fetchProjectDocumentArtifact).toHaveBeenCalledWith(3, "first", { silent: true }));
    rerender(<MemoryRouter><DocumentArtifactsView documentID="second" projectID={3} projectKey="tasq" /></MemoryRouter>);
    await waitFor(() => expect(within(screen.getByRole("article")).getByRole("heading", { name: "Second", level: 2 })).toBeInTheDocument());

    await act(async () => resolveFirst!({ id: "first", projectId: 3, type: "work_report", title: "First", body: "# First", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" }));
    expect(within(screen.getByRole("article")).queryByRole("heading", { name: "First", level: 2 })).not.toBeInTheDocument();
  });
});
