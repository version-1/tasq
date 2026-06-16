import { CircleAlert, CircleCheck, X } from "lucide-react";
import { useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";
import { toastStore, type Toast } from "@/lib/toast";
import styles from "./index.module.css";

export function ToastStack() {
  const { t } = useTranslation();
  const toasts = useSyncExternalStore(
    toastStore.subscribe,
    toastStore.getSnapshot,
    toastStore.getSnapshot,
  );

  if (toasts.length === 0) {
    return null;
  }

  return (
    <section className={styles.stack} aria-label={t("toast.region")} aria-live="polite">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onDismiss={toastStore.dismiss} />
      ))}
    </section>
  );
}

function ToastItem({
  toast,
  onDismiss,
}: {
  toast: Toast;
  onDismiss: (id: string) => void;
}) {
  const { t } = useTranslation();
  const titleKey = toast.type === "error" ? "toast.error.title" : "toast.success.title";
  const StatusIcon = toast.type === "error" ? CircleAlert : CircleCheck;

  return (
    <article className={`${styles.toast} ${styles[toast.type]}`} role="status">
      <span className={styles.iconBubble} aria-hidden="true">
        <StatusIcon className={styles.statusIcon} size={26} strokeWidth={2.5} />
      </span>
      <div className={styles.content}>
        <p className={styles.title}>{t(titleKey)}</p>
        <p className={styles.message}>{toast.message}</p>
      </div>
      <button
        type="button"
        className={styles.dismiss}
        aria-label={t("toast.dismiss")}
        title={t("toast.dismiss")}
        onClick={() => onDismiss(toast.id)}
      >
        <X aria-hidden="true" size={22} strokeWidth={2} />
      </button>
    </article>
  );
}
