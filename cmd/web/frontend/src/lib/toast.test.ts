import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Toast, ToastInput, ToastListener } from "@/lib/toast";

type ToastModule = typeof import("@/lib/toast");

async function loadToastModule(): Promise<ToastModule> {
  vi.resetModules();
  return import("@/lib/toast");
}

function collectToastChanges(module: ToastModule): {
  changes: Array<readonly Toast[]>;
  unsubscribe: () => void;
} {
  const changes: Array<readonly Toast[]> = [];
  const listener: ToastListener = (toasts) => {
    changes.push(toasts);
  };

  return {
    changes,
    unsubscribe: module.subscribe(listener),
  };
}

describe("toast store", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("adds a toast and notifies subscribers", async () => {
    const toastStore = await loadToastModule();
    const { changes, unsubscribe } = collectToastChanges(toastStore);
    const input: ToastInput = { type: "success", message: "Saved" };

    const toast = toastStore.addToast(input);

    expect(toast).toEqual({ id: "toast-1", ...input });
    expect(changes).toEqual([
      [],
      [{ id: "toast-1", type: "success", message: "Saved" }],
    ]);

    unsubscribe();
  });

  it("creates unique toast ids", async () => {
    const toastStore = await loadToastModule();

    const firstToast = toastStore.addToast({ type: "success", message: "Saved" });
    const secondToast = toastStore.addToast({ type: "error", message: "Failed" });

    expect(firstToast.id).toBe("toast-1");
    expect(secondToast.id).toBe("toast-2");
    expect(firstToast.id).not.toBe(secondToast.id);
  });

  it("stops notifying a listener after unsubscribe", async () => {
    const toastStore = await loadToastModule();
    const listener = vi.fn<ToastListener>();
    const unsubscribe = toastStore.subscribe(listener);

    unsubscribe();
    toastStore.addToast({ type: "success", message: "Saved" });

    expect(listener).toHaveBeenCalledTimes(1);
    expect(listener).toHaveBeenLastCalledWith([]);
  });

  it("automatically removes success toasts after 3 seconds", async () => {
    const toastStore = await loadToastModule();
    const { changes, unsubscribe } = collectToastChanges(toastStore);

    toastStore.addToast({ type: "success", message: "Saved" });

    vi.advanceTimersByTime(2_999);
    expect(changes.at(-1)).toEqual([{ id: "toast-1", type: "success", message: "Saved" }]);

    vi.advanceTimersByTime(1);
    expect(changes.at(-1)).toEqual([]);

    unsubscribe();
  });

  it("automatically removes error toasts after 5 seconds", async () => {
    const toastStore = await loadToastModule();
    const { changes, unsubscribe } = collectToastChanges(toastStore);

    toastStore.addToast({ type: "error", message: "Failed" });

    vi.advanceTimersByTime(4_999);
    expect(changes.at(-1)).toEqual([{ id: "toast-1", type: "error", message: "Failed" }]);

    vi.advanceTimersByTime(1);
    expect(changes.at(-1)).toEqual([]);

    unsubscribe();
  });
});
