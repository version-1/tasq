import { useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import { dismissToast, getToastsSnapshot, subscribe, type Toast } from "@/lib/toast";
import styles from "./index.module.css";

export function ToastStack() {
  const { t } = useTranslation();
  const toasts = useSyncExternalStore(subscribe, getToastsSnapshot, getToastsSnapshot);

  if (toasts.length === 0) {
    return null;
  }

  return (
    <section className={styles.stack} aria-label={t("toast.region")} aria-live="polite">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} />
      ))}
    </section>
  );
}

function ToastItem({ toast }: { toast: Toast }) {
  const { t } = useTranslation();
  const titleKey = toast.type === "error" ? "toast.errorTitle" : "toast.successTitle";

  return (
    <article className={`${styles.toast} ${styles[toast.type]}`} role="status">
      <div className={styles.content}>
        <p className={styles.title}>{t(titleKey)}</p>
        <p className={styles.message}>{toast.message}</p>
      </div>
      <button
        type="button"
        className={styles.dismiss}
        aria-label={t("toast.dismiss")}
        title={t("toast.dismiss")}
        onClick={() => dismissToast(toast.id)}
      >
        <IconProxy name="x" size={16} strokeWidth={2} />
      </button>
    </article>
  );
}
