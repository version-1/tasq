"use client";

import { Copy } from "lucide-react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { OrchestratorIssueRun } from "@/lib/types";
import { toast } from "@/lib/toast";
import styles from "./index.module.css";

type RunRowProps = {
  issueID: number;
  run: OrchestratorIssueRun;
};

export function RunRow({ issueID, run }: RunRowProps) {
  const { t } = useTranslation();
  const threadID = run.thread_id?.trim() || null;

  async function handleCopyThreadID() {
    if (!threadID) {
      return;
    }

    try {
      await navigator.clipboard.writeText(threadID);
      toast.success({ message: t("toast.success.threadIDCopied") });
    } catch {
      toast.error({ message: t("toast.error.clipboardUnavailable") });
    }
  }

  return (
    <div className={styles.runRow}>
      <Link
        className={styles.runLink}
        to={`/issues/${issueID}/conversations?runId=${encodeURIComponent(run.run_id)}`}
      >
        <span className={styles.runIdentity}>
          <span className={styles.runID}>{run.run_id}</span>
          <span className={styles.threadID}>
            <span className={styles.threadLabel}>{t("issues.detailPage.threadID")}</span>{" "}
            {threadID ?? t("issues.detailPage.noThreadID")}
          </span>
        </span>
        <span className={styles.runMeta}>
          {t(`runStatuses.${run.status}`)} · {t("issues.attempt")} {run.attempt}
        </span>
        <span className={styles.runTime}>{formatDateTime(run.updated_at)}</span>
      </Link>
      <button
        aria-label={t("issues.detailPage.copyThreadID")}
        className={styles.copyButton}
        disabled={!threadID}
        onClick={handleCopyThreadID}
        title={t("issues.detailPage.copyThreadID")}
        type="button"
      >
        <Copy aria-hidden="true" size={16} strokeWidth={2} />
      </button>
    </div>
  );
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
