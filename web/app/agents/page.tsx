"use client";

import { useLayoutData } from "@/components/layout";
import { AgentsView } from "./_components/agents-view";

export default function AgentsPage() {
  const { summary } = useLayoutData();

  return <AgentsView runs={summary.runs} />;
}
