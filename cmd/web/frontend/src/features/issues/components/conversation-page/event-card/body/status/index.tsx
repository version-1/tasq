import { useTranslation } from "react-i18next";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "../../index.module.css";

export function StatusEventBody({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();

  return (
    <p className={styles.statusMessage}>
      {event.message || t(`runStatuses.${event.event}`)}
    </p>
  );
}
