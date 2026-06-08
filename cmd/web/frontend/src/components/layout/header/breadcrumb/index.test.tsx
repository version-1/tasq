import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { Breadcrumb } from "./index";

function renderBreadcrumb(pathname: string) {
  render(
    <MemoryRouter initialEntries={[pathname]}>
      <Breadcrumb />
    </MemoryRouter>,
  );
}

describe("Breadcrumb", () => {
  it("renders URL-based breadcrumb links with the current page as plain text", () => {
    renderBreadcrumb("/issues/24/conversations");

    const breadcrumb = screen.getByRole("navigation", { name: "Breadcrumb" });
    expect(within(breadcrumb).getByRole("link", { name: "Issues" })).toHaveAttribute("href", "/issues");
    expect(within(breadcrumb).getByRole("link", { name: "#24" })).toHaveAttribute("href", "/issues/24");
    expect(within(breadcrumb).getByText("Conversation")).toHaveAttribute("aria-current", "page");
  });

  it("renders the project key before the issues segment for project-scoped routes", () => {
    renderBreadcrumb("/projects/TASQ/issues");

    const breadcrumb = screen.getByRole("navigation", { name: "Breadcrumb" });
    expect(within(breadcrumb).getByRole("link", { name: "TASQ" })).toHaveAttribute(
      "href",
      "/projects/TASQ/issues",
    );
    expect(within(breadcrumb).getByText("Issues")).toHaveAttribute("aria-current", "page");
  });
});
