"use client";

import { useLayoutData } from "@/components/layout";
import { WorkspaceView } from "./_components/workspace-view";

export default function WorkspacePage() {
  const { summary, issues } = useLayoutData();

  return <WorkspaceView summary={summary} issues={issues} />;
}
