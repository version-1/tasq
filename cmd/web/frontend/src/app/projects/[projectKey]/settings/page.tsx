"use client";

import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { PanelMessage } from "@/components/ui/pannel-message";
import { useLayoutShellData } from "@/components/layout";
import { fetchProjectWorkflow } from "@/lib/api";
import type { ProjectWorkflow } from "@/lib/types";
import { WorkflowSettingsView } from "./_components/workflow-settings-view";

type WorkflowLoadState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "ready"; workflow: ProjectWorkflow }
  | { kind: "empty" }
  | { kind: "error"; message: string };

export default function ProjectSettingsPage() {
  const { t } = useTranslation();
  const { projectKey } = useParams();
  const shellData = useLayoutShellData();
  const [workflowState, setWorkflowState] = useState<WorkflowLoadState>({ kind: "idle" });
  const project = shellData.activeProject;
  const projectID = project?.id ?? null;
  const decodedProjectKey = decodeProjectKey(projectKey);

  useEffect(() => {
    if (shellData.loadState.kind !== "ready" || projectID === null) {
      setWorkflowState({ kind: "idle" });
      return;
    }

    let ignore = false;
    setWorkflowState({ kind: "loading" });

    void fetchProjectWorkflow(projectID, { silent: true })
      .then((workflow) => {
        if (!ignore) {
          setWorkflowState({ kind: "ready", workflow });
        }
      })
      .catch((error: unknown) => {
        if (ignore) return;
        if (isNotFoundError(error)) {
          setWorkflowState({ kind: "empty" });
          return;
        }
        setWorkflowState({
          kind: "error",
          message: error instanceof Error ? error.message : t("projectSettings.failedToLoadWorkflow"),
        });
      });

    return () => {
      ignore = true;
    };
  }, [projectID, shellData.loadState.kind, t]);

  if (shellData.loadState.kind === "ready" && !project) {
    return (
      <PanelMessage
        title={t("projectSettings.projectNotFound")}
        detail={t("projectSettings.projectNotFoundDetail", { key: decodedProjectKey })}
      />
    );
  }

  if (workflowState.kind === "loading" || workflowState.kind === "idle") {
    return <PanelMessage title={t("projectSettings.loadingWorkflow")} />;
  }

  if (workflowState.kind === "error") {
    return (
      <PanelMessage
        title={t("projectSettings.failedToLoadWorkflow")}
        detail={workflowState.message}
      />
    );
  }

  if (workflowState.kind === "empty") {
    return (
      <PanelMessage
        title={t("projectSettings.workflowUnavailable")}
        detail={t("projectSettings.workflowUnavailableDetail")}
      />
    );
  }

  return <WorkflowSettingsView workflow={workflowState.workflow} />;
}

function isNotFoundError(error: unknown): boolean {
  return typeof error === "object" && error !== null && "code" in error
    ? String(error.code).endsWith("not_found")
    : false;
}

function decodeProjectKey(projectKey: string | undefined): string {
  if (!projectKey) {
    return "";
  }
  try {
    return decodeURIComponent(projectKey);
  } catch {
    return projectKey;
  }
}
