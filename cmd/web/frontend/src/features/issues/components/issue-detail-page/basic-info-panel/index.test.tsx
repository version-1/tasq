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
    renderBasicInfoPanel(issueFixtures[0]);

    const firstDependencyLink = screen.getByRole("link", {
      name: "#2 Wire issue board to generated client",
    });
    const secondDependencyLink = screen.getByRole("link", {
      name: "#3 Verify status transitions",
    });

    expect(firstDependencyLink).toHaveAttribute("href", "/issues/2");
    expect(secondDependencyLink).toHaveAttribute("href", "/issues/3");
  });

  it("shows an empty dependency state when the issue has no dependencies", () => {
    renderBasicInfoPanel(issueFixtures[1]);

    expect(screen.getByText("No dependencies")).toBeInTheDocument();
  });
});
