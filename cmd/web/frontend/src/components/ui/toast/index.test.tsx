import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";
import { toastStore } from "@/lib/toast";
import { ToastStack } from "./index";

describe("ToastStack", () => {
  afterEach(() => {
    toastStore.clear();
  });

  it("renders toast messages and supports manual dismissal", async () => {
    const user = userEvent.setup();
    toastStore.error({ message: "Could not refresh" });

    render(<ToastStack />);

    expect(screen.getByRole("status")).toHaveTextContent("Could not refresh");

    await user.click(screen.getByRole("button", { name: "Dismiss" }));

    expect(screen.queryByRole("status")).toBeNull();
  });
});
