import { useState } from "react";
import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import styles from "./index.module.css";

const projectAddCommand = "tq project add --key tasq .";

export function AddProjectDialog({
  onCancel,
}: {
  onCancel: () => void;
}) {
  const { t } = useTranslation();
  const [didCopy, setDidCopy] = useState(false);

  async function handleCopy() {
    await globalThis.navigator.clipboard?.writeText(projectAddCommand);
    setDidCopy(true);
  }

  return (
    <div className={styles.dialogBackdrop} role="presentation">
      <section
        aria-labelledby="add-project-title"
        aria-modal="true"
        className={styles.dialog}
        role="dialog"
      >
        <div className={styles.dialogContent}>
          <div className={styles.dialogHeader}>
            <div>
              <h2 id="add-project-title">{t("addProject.title")}</h2>
              <p>{t("addProject.description")}</p>
            </div>
            <button type="button" aria-label={t("addProject.close")} onClick={onCancel}>
              x
            </button>
          </div>

          <div className={styles.commandBlock}>
            <span>{t("addProject.commandLabel")}</span>
            <pre>
              <code>{projectAddCommand}</code>
            </pre>
            <button type="button" className={styles.copyButton} onClick={() => void handleCopy()}>
              <IconProxy name="copy" size={15} />
              {didCopy ? t("addProject.copied") : t("addProject.copy")}
            </button>
          </div>

          <p className={styles.helpText}>{t("addProject.helpText")}</p>

          <div className={styles.dialogActions}>
            <button type="button" onClick={onCancel}>
              {t("addProject.close")}
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}
