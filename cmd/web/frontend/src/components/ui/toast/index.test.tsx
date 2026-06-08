import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { addToast, dismissToast, getToastsSnapshot } from "@/lib/toast";
import { ToastStack } from "./index";

function clearToasts() {
  getToastsSnapshot().forEach((toast) => dismissToast(toast.id));
}

describe("ToastStack", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    clearToasts();
  });

  afterEach(() => {
    clearToasts();
    vi.useRealTimers();
  });

  it("renders success and error toasts in a top-right stack", () => {
    addToast({ type: "success", message: "Issue saved" });
    addToast({ type: "error", message: "Failed to save issue" });

    render(<ToastStack />);

    expect(screen.getByRole("region", { name: "Notifications" })).toBeTruthy();
    expect(screen.getByText("Success")).toBeTruthy();
    expect(screen.getByText("Issue saved")).toBeTruthy();
    expect(screen.getByText("Error")).toBeTruthy();
    expect(screen.getByText("Failed to save issue")).toBeTruthy();
  });

  it("dismisses a toast from the close button", () => {
    addToast({ type: "success", message: "Issue saved" });
    render(<ToastStack />);

    fireEvent.click(screen.getByRole("button", { name: "Dismiss notification" }));

    expect(screen.queryByText("Issue saved")).toBeNull();
  });

  it("removes a toast automatically when the store expires it", () => {
    addToast({ type: "success", message: "Issue saved" });
    render(<ToastStack />);

    act(() => {
      vi.advanceTimersByTime(3_000);
    });

    expect(screen.queryByText("Issue saved")).toBeNull();
  });
});
