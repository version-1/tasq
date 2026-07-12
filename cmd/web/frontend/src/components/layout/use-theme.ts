import { useEffect, useState } from "react";

const themeStorageKey = "tasq.theme";

type Theme = "dark" | "light";

function storedTheme(): Theme | null {
  const stored = window.localStorage.getItem(themeStorageKey);
  return stored === "dark" || stored === "light" ? stored : null;
}

function systemTheme(): Theme {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function useTheme() {
  const [preference, setPreference] = useState<Theme | null>(storedTheme);
  const [systemPreference, setSystemPreference] = useState<Theme>(systemTheme);
  const theme = preference ?? systemPreference;

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  useEffect(() => {
    if (preference !== null) return;

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const updateSystemPreference = (event: MediaQueryListEvent) => {
      setSystemPreference(event.matches ? "dark" : "light");
    };

    mediaQuery.addEventListener("change", updateSystemPreference);
    return () => mediaQuery.removeEventListener("change", updateSystemPreference);
  }, [preference]);

  function setIsDark(isDark: boolean) {
    const nextTheme = isDark ? "dark" : "light";
    setPreference(nextTheme);
    window.localStorage.setItem(themeStorageKey, nextTheme);
  }

  return {
    isDark: theme === "dark",
    setIsDark,
  };
}
