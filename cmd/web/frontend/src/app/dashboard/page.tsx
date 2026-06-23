"use client";

import { useLayoutData } from "@/components/layout";
import { DashboardView } from "@/features/dashboard/components/dashboard-view";

export default function DashboardPage() {
  const { summary, issues, refreshIntervalMs } = useLayoutData();

  return <DashboardView summary={summary} issues={issues} refreshIntervalMs={refreshIntervalMs} />;
}
