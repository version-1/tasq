import type { ReactNode } from "react";
import { ShellLayout, useLayoutShellData } from "@/components/layout";

export function ProjectLayout({ children }: { children: ReactNode }) {
  const shellData = useLayoutShellData();

  return (
    <ShellLayout
      shellData={shellData}
      showViewNavigation={!shellData.isIssueDetailPage}
    >
      {children}
    </ShellLayout>
  );
}
