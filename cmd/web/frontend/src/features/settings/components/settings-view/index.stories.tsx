import type { Meta, StoryObj } from "@storybook/react-vite";
import { noop } from "@/stories/fixtures";
import { SettingsView } from "./index";

const meta = {
  title: "Features/Settings/SettingsView",
  component: SettingsView,
  args: {
    refreshIntervalMs: 3000,
    language: "ja",
    generatedAt: "2026-06-16T00:00:00.000Z",
    onRefreshIntervalChange: noop,
    onLanguageChange: noop,
  },
} satisfies Meta<typeof SettingsView>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};
