import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { issueFixtures } from "@/mocks/fixtures/issues";
import { storySummary } from "@/stories/fixtures";
import { BasicInfoPanel } from "./index";

const issueOptions = storySummary.columns.flatMap((column) => column.issues);

function renderBasicInfoPanel(issue = issueFixtures[2]) {
  render(
    <MemoryRouter>
      <BasicInfoPanel
        disabled={false}
        issue={issue}
        issueOptions={issueOptions}
        onStatusChange={vi.fn().mockResolvedValue(undefined)}
      />
    </MemoryRouter>,
  );
}

describe("BasicInfoPanel", () => {
  it("renders dependency issue links with id and title", () => {
    renderBasicInfoPanel();

    const dependencyLink = screen.getByRole("link", {
      name: "#2 Wire issue board to generated client",
    });

    expect(dependencyLink).toHaveAttribute("href", "/issues/2");
  });

  it("shows an empty dependency state when the issue has no dependencies", () => {
    renderBasicInfoPanel(issueFixtures[0]);

    expect(screen.getByText("No dependencies")).toBeInTheDocument();
  });
});
