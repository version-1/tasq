"use client";

import { useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useLayoutShellData } from "@/components/layout";
import { PanelMessage } from "@/components/ui/pannel-message";
import { DocumentArtifactsView } from "@/features/projects/components/document-artifacts-view";

export default function ProjectArtifactsPage() {
  const { t } = useTranslation();
  const { projectKey } = useParams();
  const { activeProject, loadState } = useLayoutShellData();
  if (loadState.kind !== "ready") return <PanelMessage title={t("projectArtifacts.loading")} />;
  if (!activeProject) return <PanelMessage title={t("projectSettings.projectNotFound")} detail={t("projectSettings.projectNotFoundDetail", { key: projectKey ?? "" })} />;
  return <DocumentArtifactsView key={activeProject.id} projectID={activeProject.id} projectKey={activeProject.key} />;
}
