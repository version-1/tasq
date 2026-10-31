import { useTranslation } from "react-i18next";
import { ContextMenuHelp, ContextMenuItem } from "@/components/ui/context-menu";
import { IconProxy } from "@/components/ui/icon-proxy";
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
      <ContextMenuItem
        checked={issue.autoOk}
        disabled={disabled || isSaving || !onChange}
        icon={<IconProxy name="circle-play" />}
        label={t("issues.autoOK.toggle", { id: issue.id })}
        onSelect={() => { void toggle(!issue.autoOk); }}
        trailingIcon={
          <span className={styles.count}>
            {issue.autoOk ? "ON" : "OFF"} · {issue.autoOkCount}/5
            {issue.autoOk ? <IconProxy name="check" /> : null}
          </span>
        }
      >
        {t("issues.autoOK.label")}
      </ContextMenuItem>
      {reason ? <ContextMenuHelp>{reason}</ContextMenuHelp> : null}
      {hasError ? <span className={styles.error} role="alert">{t("issues.autoOK.saveError")}</span> : null}
    </div>
  );
}
