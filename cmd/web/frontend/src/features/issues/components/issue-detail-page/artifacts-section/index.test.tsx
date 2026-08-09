import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import "@/lib/i18n";
import { ArtifactsSection } from "./index";

describe("ArtifactsSection", () => {
  it("renders a pull request link in a safe new tab", () => {
    render(
      <ArtifactsSection
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
  });

  it("does not render an empty section when no pull request artifact exists", () => {
    const { container } = render(<ArtifactsSection artifacts={[]} />);

    expect(container).toBeEmptyDOMElement();
  });
});
