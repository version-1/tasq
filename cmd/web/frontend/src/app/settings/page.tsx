"use client";

import { useLayoutData } from "@/components/layout";
import { SettingsView } from "./_components/settings-view";

export default function SettingsPage() {
  const {
    summary,
    refreshIntervalMs,
    language,
    onRefreshIntervalChange,
    onLanguageChange,
  } = useLayoutData();

  return (
    <SettingsView
      refreshIntervalMs={refreshIntervalMs}
      language={language}
      generatedAt={summary.generatedAt}
      onRefreshIntervalChange={onRefreshIntervalChange}
      onLanguageChange={onLanguageChange}
    />
  );
}
