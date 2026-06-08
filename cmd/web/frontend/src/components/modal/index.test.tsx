import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ModalOutlet } from "@/components/modal";

describe("ModalOutlet", () => {
  it("renders children through a portal slot", () => {
    render(
      <ModalOutlet>
        <div role="dialog">Projected modal content</div>
      </ModalOutlet>,
    );

    expect(screen.getByRole("dialog", { name: "" })).toBeTruthy();
    expect(screen.getByText("Projected modal content")).toBeTruthy();
  });
});
