"use client";

import { useTranslation } from "react-i18next";
import { PanelMessage } from "@/components/ui/pannel-message";
import { useLayoutData, useLayoutShellData } from "@/components/layout";
import { IssuesTableView } from "@/features/issues/components/table-view";

export default function IssuesTablePage() {
  const { t } = useTranslation();
  const { refreshIntervalMs } = useLayoutData();
  const { activeProject, isProjectIssueScope, projects } = useLayoutShellData();

  if (isProjectIssueScope && !activeProject) {
    return <PanelMessage title={t("projectSettings.projectNotFound")} />;
  }

  return (
    <IssuesTableView
      projectOptions={projects}
      projectID={activeProject?.id ?? null}
      refreshIntervalMs={refreshIntervalMs}
    />
  );
}
