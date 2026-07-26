import { useTranslation } from "react-i18next";
import { IssueFilterOptions } from "@/features/issues/components/filter-options";
import { ConversationEvents } from "@/features/issues/components/conversation-page";
import { itemCompletedView } from "@/features/issues/components/conversation-page/event-card/payload";
import type {
  OrchestratorConversation,
  OrchestratorConversationEvent,
  OrchestratorIssueRun,
} from "@/lib/types";
import styles from "./index.module.css";

export const defaultConversationMessageTypes = [
  "running",
  "succeeded",
  "failed",
  "item/completed:agentMessage",
  "item/completed:commandExecution",
] as const;

const itemCompletedPrefix = "item/completed:";

type ConversationTabProps = {
  conversation: OrchestratorConversation | null;
  error: string;
  isLoading: boolean;
  messageTypes: string[];
  onMessageTypesChange: (types: string[]) => void;
  onRunChange: (runID: string) => void;
  runs: OrchestratorIssueRun[];
  runsError: string;
  runsLoading: boolean;
  selectedRunID: string;
};

export function ConversationTab({
  conversation,
  error,
  isLoading,
  messageTypes,
  onMessageTypesChange,
  onRunChange,
  runs,
  runsError,
  runsLoading,
  selectedRunID,
}: ConversationTabProps) {
  const { t } = useTranslation();
  const messageTypeOptions = eventTypeOptions(conversation?.events ?? [], t);
  const visibleEvents = filterEvents(conversation?.events ?? [], messageTypes);

  return (
    <section className={styles.section} aria-labelledby="issue-conversation">
      <div className={styles.sectionHeader}>
        <div>
          <h3 id="issue-conversation">{t("issues.detailPage.conversation")}</h3>
          <p>{t("issues.detailPage.conversationDescription")}</p>
        </div>
      </div>
      <div className={styles.toolbar}>
        <label className={styles.runSelectLabel}>
          <span>{t("issues.detailPage.run")}</span>
          <select
            value={selectedRunID}
            disabled={runsLoading || runs.length === 0}
            onChange={(event) => onRunChange(event.target.value)}
          >
            {runs.length === 0 ? (
              <option value="">{t("issues.detailPage.noRuns")}</option>
            ) : null}
            {runs.map((run) => (
              <option key={run.run_id} value={run.run_id}>
                {t("issues.detailPage.runOption", {
                  attempt: run.attempt,
                  runID: run.run_id,
                  status: t(`runStatuses.${run.status}`),
                })}
              </option>
            ))}
          </select>
        </label>
        <IssueFilterOptions
          allLabel={t("issues.detailPage.allMessageTypes")}
          applyLabel={t("issues.table.apply")}
          cancelLabel={t("issues.table.cancel")}
          clearLabel={t("issues.table.clearAll")}
          label={t("issues.detailPage.messageTypes")}
          onChange={(values) => onMessageTypesChange(values.map(String))}
          options={messageTypeOptions}
          selectedCountLabel={(count) => t("issues.table.selectedCount", { count })}
          selectedValues={messageTypes}
        />
      </div>
      {runsLoading ? <p className={styles.muted}>{t("layout.loading")}</p> : null}
      {runsError ? <p className={styles.error}>{runsError}</p> : null}
      {isLoading ? <p className={styles.muted}>{t("layout.loading")}</p> : null}
      {error ? <p className={styles.error}>{error}</p> : null}
      {!runsLoading && !runsError && runs.length === 0 ? (
        <p className={styles.muted}>{t("issues.detailPage.noRuns")}</p>
      ) : null}
      {conversation && !isLoading && !error ? (
        visibleEvents.length === 0 && conversation.events.length > 0 ? (
          <p className={styles.muted}>{t("issues.detailPage.noMatchingConversationEvents")}</p>
        ) : (
          <ConversationEvents events={visibleEvents} />
        )
      ) : null}
    </section>
  );
}

function filterEvents(
  events: OrchestratorConversationEvent[],
  selectedTypes: string[],
): OrchestratorConversationEvent[] {
  if (selectedTypes.length === 0) {
    return events;
  }
  return events.filter((event) => {
    if (event.event !== "item/completed") {
      return selectedTypes.includes(event.event);
    }

    const itemType = itemCompletedView(event.payload_json).type;
    return selectedTypes.includes("item/completed")
      || selectedTypes.includes(`${itemCompletedPrefix}${itemType}`);
  });
}

function eventTypeOptions(
  events: OrchestratorConversationEvent[],
  t: (key: string) => string,
) {
  const itemTypes = Array.from(new Set(
    events
      .filter((event) => event.event === "item/completed")
      .map((event) => itemCompletedView(event.payload_json).type),
  ));

  return [
    { label: t("issues.detailPage.messageTypeRunning"), value: "running" },
    { label: t("issues.detailPage.messageTypeSucceeded"), value: "succeeded" },
    { label: t("issues.detailPage.messageTypeFailed"), value: "failed" },
    {
      label: t("issues.detailPage.messageTypeItemCompleted"),
      children: itemTypes.map((itemType) => ({
        label: itemType,
        value: `${itemCompletedPrefix}${itemType}`,
      })),
    },
    { label: t("issues.detailPage.messageTypeTokenUsageUpdated"), value: "thread/tokenUsage/updated" },
    { label: t("issues.detailPage.messageTypeRateLimitsUpdated"), value: "account/rateLimits/updated" },
  ];
}
