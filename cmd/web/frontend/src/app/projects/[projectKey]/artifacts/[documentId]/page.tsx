import { useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useLayoutShellData } from "@/components/layout";
import { DocumentDetailView } from "@/features/projects/components/document-detail-view";

export default function ProjectDocumentPage() {
  const { t } = useTranslation();
  const { projectKey = "", documentId = "" } = useParams();
  const { activeProject, loadState } = useLayoutShellData();
  const projectError = loadState.kind === "error"
    ? loadState.message
    : loadState.kind === "ready" && !activeProject
      ? t("projectSettings.projectNotFoundDetail", { key: projectKey })
      : "";

  return (
    <DocumentDetailView
      key={JSON.stringify([projectKey, activeProject?.id, documentId])}
      projectID={activeProject?.id}
      projectKey={projectKey}
      documentID={documentId}
      projectError={projectError}
    />
  );
}
