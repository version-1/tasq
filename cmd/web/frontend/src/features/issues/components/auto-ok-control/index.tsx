import { useTranslation } from "react-i18next";
import { Switch } from "@/components/ui/switch";
import type { Issue } from "@/lib/types";
import type { AutoOKChangeHandler } from "../issues-view/types";
import { useAutoOKToggle } from "@/features/issues/hooks/use-auto-ok-toggle";
import styles from "./index.module.css";

export function AutoOKControl({ issue, disabled = false, onChange }: {
  issue: Issue;
  disabled?: boolean;
  onChange?: AutoOKChangeHandler;
}) {
  const { t } = useTranslation();
  const { isSaving, hasError, toggle } = useAutoOKToggle(issue.id, onChange);
  const reason = issue.autoOkStopReason === "limit"
    ? t("issues.autoOK.limit")
    : issue.autoOkStopReason === "error" ? t("issues.autoOK.error") : null;

  return (
    <div className={styles.control} aria-busy={isSaving}>
      <label className={styles.row}>
        <span>{t("issues.autoOK.label")}</span>
        <span className={styles.count}>{issue.autoOkCount}/5</span>
        <Switch
          aria-label={t("issues.autoOK.toggle", { id: issue.id })}
          checked={issue.autoOk}
          disabled={disabled || isSaving || !onChange}
          onChange={(enabled) => { void toggle(enabled); }}
        />
      </label>
      {reason ? <span className={styles.reason}>{reason}</span> : null}
      {hasError ? <span className={styles.error} role="alert">{t("issues.autoOK.saveError")}</span> : null}
    </div>
  );
}
