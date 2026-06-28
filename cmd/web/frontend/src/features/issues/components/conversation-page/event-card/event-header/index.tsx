import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import { Markdown } from "@/components/ui/markdown";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "../index.module.css";
import { itemCompletedView, type Translator } from "../payload";

type EventHeaderProps = {
  bodyID: string;
  event: OrchestratorConversationEvent;
  isBodyOpen: boolean;
  onToggleBody: () => void;
};

export function EventHeader({
  bodyID,
  event,
  isBodyOpen,
  onToggleBody,
}: EventHeaderProps) {
  const { t } = useTranslation();
  const foldLabel = isBodyOpen
    ? t("issues.detailPage.collapseConversationEvent")
    : t("issues.detailPage.expandConversationEvent");

  return (
    <div className={styles.header}>
      <HeaderTitle event={event} t={t} />
      <div className={styles.headerMeta}>
        <span className={styles.time}>{formatDateTime(event.at)}</span>
        <button
          type="button"
          className={styles.foldButton}
          aria-controls={bodyID}
          aria-expanded={isBodyOpen}
          aria-label={foldLabel}
          title={foldLabel}
          onClick={onToggleBody}
        >
          <IconProxy className={styles.foldIcon} name="chevron-down" size={16} strokeWidth={2.3} />
        </button>
      </div>
    </div>
  );
}

function HeaderTitle({
  event,
  t,
}: {
  event: OrchestratorConversationEvent;
  t: Translator;
}) {
  if (event.event === "item/commandExecution/requestApproval") {
    return (
      <span className={[styles.type, styles.approvalBadge].join(" ")}>
        {eventLabel(event.event, t)}
      </span>
    );
  }

  if (event.event !== "item/completed") {
    return <span className={styles.type}>{eventLabel(event.event, t)}</span>;
  }

  const item = itemCompletedView(event.payload_json);
  if (item.command === "") {
    return <span className={styles.type}>{item.type}</span>;
  }

  return (
    <div className={styles.headerTitleStack}>
      <span className={styles.type}>{item.type}</span>
      <Markdown
        className={styles.commandMarkdown}
        content={`\`\`\`shell\n${item.command}\n\`\`\``}
        emptyText=""
      />
    </div>
  );
}

function eventLabel(event: OrchestratorConversationEvent["event"], t: Translator): string {
  if (event === "turn_completed") {
    return t("issues.detailPage.turnCompleted");
  }
  return t(`runStatuses.${event}`);
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
