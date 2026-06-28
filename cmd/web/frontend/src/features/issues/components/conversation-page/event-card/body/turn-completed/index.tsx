import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/ui/markdown";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "../../index.module.css";
import { extractAggregatedOutput } from "../../payload";

export function TurnCompletedEventBody({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();

  return (
    <Markdown
      className={styles.markdown}
      content={extractAggregatedOutput(event.payload_json)}
      emptyText={event.message || t("issues.detailPage.emptyConversationEvent")}
    />
  );
}
