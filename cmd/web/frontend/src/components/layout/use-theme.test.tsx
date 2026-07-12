import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useTheme } from "./use-theme";

class MatchMediaMock {
  matches = false;
  private listeners = new Set<(event: MediaQueryListEvent) => void>();

  addEventListener(_type: string, listener: (event: MediaQueryListEvent) => void) {
    this.listeners.add(listener);
  }

  removeEventListener(_type: string, listener: (event: MediaQueryListEvent) => void) {
    this.listeners.delete(listener);
  }

  setMatches(matches: boolean) {
    this.matches = matches;
    for (const listener of this.listeners) {
      listener({ matches } as MediaQueryListEvent);
    }
  }
}

function ThemeProbe() {
  const { isDark, setIsDark } = useTheme();
  return (
    <button type="button" onClick={() => setIsDark(!isDark)}>
      {isDark ? "dark" : "light"}
    </button>
  );
}

describe("useTheme", () => {
  let mediaQuery: MatchMediaMock;

  beforeEach(() => {
    document.documentElement.removeAttribute("data-theme");
    mediaQuery = new MatchMediaMock();
    vi.stubGlobal("matchMedia", vi.fn(() => mediaQuery));
  });

  it("uses the saved theme in preference to the system setting", () => {
    window.localStorage.setItem("tasq.theme", "light");
    mediaQuery.matches = true;

    render(<ThemeProbe />);

    expect(screen.getByRole("button")).toHaveTextContent("light");
    expect(document.documentElement).toHaveAttribute("data-theme", "light");
  });

  it("follows changes to the system setting before a user choice", async () => {
    render(<ThemeProbe />);

    mediaQuery.setMatches(true);
    await waitFor(() => {
      expect(screen.getByRole("button")).toHaveTextContent("dark");
      expect(document.documentElement).toHaveAttribute("data-theme", "dark");
    });
  });

  it("persists a user choice and stops following the system setting", async () => {
    const user = userEvent.setup();
    render(<ThemeProbe />);

    await user.click(screen.getByRole("button"));
    expect(window.localStorage.getItem("tasq.theme")).toBe("dark");

    mediaQuery.setMatches(false);
    expect(screen.getByRole("button")).toHaveTextContent("dark");
  });
});
