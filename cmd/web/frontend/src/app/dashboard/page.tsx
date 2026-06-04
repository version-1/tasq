"use client";

import { useLayoutData } from "@/components/layout";
import { DashboardView } from "./_components/dashboard-view";

export default function DashboardPage() {
  const { summary, issues } = useLayoutData();

  return <DashboardView summary={summary} issues={issues} />;
}
