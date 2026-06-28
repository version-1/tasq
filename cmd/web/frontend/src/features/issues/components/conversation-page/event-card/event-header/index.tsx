import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import { Markdown } from "@/components/ui/markdown";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "../index.module.css";
import { itemCompletedView, type Translator } from "../payload";

type EventHeaderProps = {
  bodyID: string;
  canFold?: boolean;
  event: OrchestratorConversationEvent;
  isBodyOpen: boolean;
  onToggleBody: () => void;
};

export function EventHeader({
  bodyID,
  canFold = true,
  event,
  isBodyOpen,
  onToggleBody,
}: EventHeaderProps) {
  const { t } = useTranslation();
  const foldLabel = isBodyOpen
    ? t("issues.detailPage.collapseConversationEvent")
    : t("issues.detailPage.expandConversationEvent");
  const title = headerTitle(event, t);

  return (
    <div className={styles.header}>
      <div className={styles.headerTop}>
        {title.label}
        <div className={styles.headerMeta}>
          <span className={styles.time}>{formatDateTime(event.at)}</span>
          {canFold ? (
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
          ) : null}
        </div>
      </div>
      {title.command ? (
        <Markdown
          className={styles.commandMarkdown}
          content={`\`\`\`shell\n${title.command}\n\`\`\``}
          emptyText=""
        />
      ) : null}
    </div>
  );
}

function headerTitle(
  event: OrchestratorConversationEvent,
  t: Translator,
): { label: ReactNode; command: string } {
  if (event.event === "item/commandExecution/requestApproval") {
    return {
      label: (
        <span className={[styles.type, styles.approvalBadge].join(" ")}>
          {eventLabel(event.event, t)}
        </span>
      ),
      command: "",
    };
  }

  if (event.event !== "item/completed") {
    return {
      label: <span className={styles.type}>{eventLabel(event.event, t)}</span>,
      command: "",
    };
  }

  const item = itemCompletedView(event.payload_json);
  return {
    label: <span className={styles.type}>{item.type}</span>,
    command: item.command,
  };
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
