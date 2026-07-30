import type { Preview } from "@storybook/react-vite";
import { QueryClientProvider } from "@tanstack/react-query";
import { type PropsWithChildren, useEffect } from "react";
import { MemoryRouter } from "react-router-dom";
import "../src/app/globals.css";
import "../src/lib/i18n";
import { appQueryClient } from "../src/lib/query-client";

const StorybookTheme = {
  dark: "dark",
  light: "light",
} as const;

type StorybookTheme = (typeof StorybookTheme)[keyof typeof StorybookTheme];

type StorybookThemeProps = PropsWithChildren<{
  theme: StorybookTheme;
}>;

function StorybookThemeProvider({ children, theme }: StorybookThemeProps) {
  useEffect(() => {
    const root = document.documentElement;
    const previousTheme = root.getAttribute("data-theme");

    if (theme === StorybookTheme.dark) {
      root.setAttribute("data-theme", theme);
    } else {
      root.removeAttribute("data-theme");
    }

    return () => {
      if (previousTheme === null) {
        root.removeAttribute("data-theme");
      } else {
        root.setAttribute("data-theme", previousTheme);
      }
    };
  }, [theme]);

  return children;
}

const preview: Preview = {
  globalTypes: {
    theme: {
      defaultValue: StorybookTheme.light,
      description: "Preview color theme",
      toolbar: {
        items: [
          { title: "Light", value: StorybookTheme.light },
          { title: "Dark", value: StorybookTheme.dark },
        ],
        title: "Theme",
      },
    },
  },
  decorators: [
    (Story, context) => (
      <StorybookThemeProvider
        theme={
          context.globals.theme === StorybookTheme.dark
            ? StorybookTheme.dark
            : StorybookTheme.light
        }
      >
        <QueryClientProvider client={appQueryClient}>
          <MemoryRouter initialEntries={["/issues"]}>
            <Story />
          </MemoryRouter>
        </QueryClientProvider>
      </StorybookThemeProvider>
    ),
  ],
  parameters: {
    controls: {
      expanded: true,
    },
  },
};

export default preview;
