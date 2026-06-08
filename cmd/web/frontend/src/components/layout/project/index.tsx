import type { ReactNode } from "react";
import { ShellLayout, useLayoutShellData } from "@/components/layout";
import type { HeaderPageLink } from "@/components/layout/header";

export function ProjectLayout({
  children,
  pages,
}: {
  children: ReactNode;
  pages: readonly HeaderPageLink[];
}) {
  const shellData = useLayoutShellData();

  return (
    <ShellLayout
      pages={pages}
      shellData={shellData}
      showViewNavigation={!shellData.isIssueDetailPage}
    >
      {children}
    </ShellLayout>
  );
}
