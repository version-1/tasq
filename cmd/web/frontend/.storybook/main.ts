import type { StorybookConfig } from "@storybook/react-vite";
import { mergeConfig } from "vite";

const config: StorybookConfig = {
  framework: "@storybook/react-vite",
  stories: ["../src/**/*.stories.@(ts|tsx)"],
  addons: [],
  viteFinal: async (config) =>
    mergeConfig(config, {
      resolve: {
        alias: {
          "@": "/src",
        },
      },
    }),
};

export default config;
