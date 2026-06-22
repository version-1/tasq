import type { ReactNode } from "react";
import { ShellLayout, useLayoutShellData } from "@/components/layout";

export function ProjectLayout({
  children,
  showAddTaskButton = true,
}: {
  children: ReactNode;
  showAddTaskButton?: boolean;
}) {
  const shellData = useLayoutShellData();

  return (
    <ShellLayout
      shellData={shellData}
      showAddTaskButton={showAddTaskButton}
      showViewNavigation={!shellData.isIssueDetailPage}
    >
      {children}
    </ShellLayout>
  );
}
