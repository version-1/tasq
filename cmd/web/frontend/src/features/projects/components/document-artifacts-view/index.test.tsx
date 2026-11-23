import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import "@/lib/i18n";
import { DocumentArtifactsView } from "./index";

const api = vi.hoisted(() => ({
  fetchProjectDocumentArtifact: vi.fn(),
  fetchProjectDocumentArtifacts: vi.fn(),
}));
vi.mock("@/lib/api", () => api);

function renderView() {
  return render(
    <MemoryRouter>
      <DocumentArtifactsView projectID={3} projectKey="tasq" />
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
});
