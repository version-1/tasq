import type { ReactNode } from "react";
import { ShellLayout, useLayoutShellData } from "@/components/layout";

export function DefaultLayout({ children }: { children: ReactNode }) {
  const shellData = useLayoutShellData();

  return (
    <ShellLayout shellData={shellData}>
      {children}
    </ShellLayout>
  );
}
