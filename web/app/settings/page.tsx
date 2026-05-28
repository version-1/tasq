"use client";

import { useLayoutData } from "@/components/layout";
import { SettingsView } from "./_components/settings-view";

export default function SettingsPage() {
  const { summary, refreshIntervalMs, onRefreshIntervalChange } = useLayoutData();

  return (
    <SettingsView
      refreshIntervalMs={refreshIntervalMs}
      generatedAt={summary.generatedAt}
      onRefreshIntervalChange={onRefreshIntervalChange}
    />
  );
}
