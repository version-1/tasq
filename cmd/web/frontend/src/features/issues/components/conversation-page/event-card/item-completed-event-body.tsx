import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/ui/markdown";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "./index.module.css";
import { itemCompletedView } from "./payload";

export function ItemCompletedEventBody({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();
  const item = itemCompletedView(event.payload_json);
  const content = [item.aggregatedOutput, item.text].filter(Boolean).join("\n\n");

  return (
    <div className={styles.itemBody}>
      <Markdown
        className={styles.markdown}
        content={content}
        emptyText={event.message || t("issues.detailPage.emptyConversationEvent")}
      />
      {item.exitCode !== null && item.exitCode !== 0 ? (
        <p className={styles.exitCode}>{t("issues.detailPage.exitCode", { code: item.exitCode })}</p>
      ) : null}
    </div>
  );
}
