import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import "@/lib/i18n";
import { ArtifactsSection } from "./index";

const api = vi.hoisted(() => ({ fetchDocumentArtifacts: vi.fn(), fetchDocumentArtifact: vi.fn() }));
vi.mock("@/lib/api", () => api);

describe("ArtifactsSection", () => {
  beforeEach(() => {
    api.fetchDocumentArtifacts.mockReset();
    api.fetchDocumentArtifact.mockReset();
    api.fetchDocumentArtifacts.mockResolvedValue([]);
  });

  it("renders a pull request link in a safe new tab", () => {
    render(
      <ArtifactsSection
        issueID={42}
        artifacts={[
          {
            type: "pull_request",
            data_type: "url",
            data_value: "https://github.com/version-1/tasq/pull/14",
          },
        ]}
      />,
    );

    expect(screen.getByRole("heading", { name: "Artifacts" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Pull request" })).toHaveAttribute(
      "href",
      "https://github.com/version-1/tasq/pull/14",
    );
    expect(screen.getByRole("link", { name: "Pull request" })).toHaveAttribute("target", "_blank");
    expect(screen.getByRole("link", { name: "Pull request" })).toHaveAttribute(
      "rel",
      "noopener noreferrer",
    );
    expect(screen.getByText("version-1/tasq #14")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Pull request" }).querySelector(".lucide-git-pull-request"),
    ).toBeInTheDocument();
  });

  it("does not render an empty section when no pull request artifact exists", () => {
    const { container } = render(<ArtifactsSection artifacts={[]} issueID={42} />);

    return waitFor(() => expect(container).toBeEmptyDOMElement());
  });

  it("shows document loading while the list request is pending", () => {
    api.fetchDocumentArtifacts.mockReturnValue(new Promise(() => undefined));
    render(<ArtifactsSection artifacts={[]} issueID={42} />);

    expect(screen.getByRole("status")).toHaveTextContent("Loading documents...");
  });

  it("lists documents and renders selected Markdown", async () => {
    const user = userEvent.setup();
    api.fetchDocumentArtifacts.mockResolvedValue([{ id: "doc-1", issueId: 42, type: "work_report", title: "Report", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" }]);
    api.fetchDocumentArtifact.mockResolvedValue({ id: "doc-1", issueId: 42, type: "work_report", title: "Report", body: "# Safe report\n\n<script>alert(1)</script>", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" });
    render(<ArtifactsSection artifacts={[]} issueID={42} />);
    await user.click(await screen.findByRole("button", { name: /Report/ }));
    expect(await screen.findByRole("heading", { name: "Safe report" })).toBeInTheDocument();
    expect(screen.queryByText("alert(1)")).not.toBeInTheDocument();
  });

  it("renders long Markdown document content", async () => {
    const user = userEvent.setup();
    const body = `${"Long report content ".repeat(400)}\n\n## Verification`;
    api.fetchDocumentArtifacts.mockResolvedValue([{ id: "doc-2", issueId: 42, type: "work_report", title: "Long report", createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" }]);
    api.fetchDocumentArtifact.mockResolvedValue({ id: "doc-2", issueId: 42, type: "work_report", title: "Long report", body, createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" });
    render(<ArtifactsSection artifacts={[]} issueID={42} />);

    await user.click(await screen.findByRole("button", { name: /Long report/ }));
    expect(await screen.findByRole("heading", { name: "Verification" })).toBeInTheDocument();
    expect(screen.getByText(/Long report content Long report content/)).toBeInTheDocument();
  });

  it("shows an empty document state beside a pull request", async () => {
    render(<ArtifactsSection issueID={42} artifacts={[{ type: "pull_request", data_type: "url", data_value: "https://github.com/version-1/tasq/pull/14" }]} />);

    expect(await screen.findByText("No documents")).toBeInTheDocument();
  });

  it("shows an error when document loading fails", async () => {
    api.fetchDocumentArtifacts.mockRejectedValue(new Error("offline"));
    render(<ArtifactsSection artifacts={[]} issueID={42} />);
    expect(await screen.findByText("offline")).toHaveAttribute("role", "status");
  });
});
