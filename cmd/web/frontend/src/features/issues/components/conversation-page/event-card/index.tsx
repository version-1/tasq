import { useId, useState } from "react";
import type { OrchestratorConversationEvent } from "@/lib/types";
import { ApprovalRequestEventBody } from "./body/approval-request";
import { ItemCompletedEventBody } from "./body/item-completed";
import { EventBodyPreview } from "./body/preview";
import { RateLimitsEventBody } from "./body/rate-limits";
import { StatusEventBody } from "./body/status";
import { TokenUsageEventBody } from "./body/token-usage";
import { TurnCompletedEventBody } from "./body/turn-completed";
import { EventHeader } from "./event-header";
import styles from "./index.module.css";

export { ApprovalRequestEventBody } from "./body/approval-request";
export { ItemCompletedEventBody } from "./body/item-completed";
export { EventBodyPreview } from "./body/preview";
export { RateLimitsEventBody } from "./body/rate-limits";
export { StatusEventBody } from "./body/status";
export { TokenUsageEventBody } from "./body/token-usage";
export { TurnCompletedEventBody } from "./body/turn-completed";
export { EventHeader } from "./event-header";

export function EventCard({ event }: { event: OrchestratorConversationEvent }) {
  const bodyID = useId();
  const [isBodyOpen, setIsBodyOpen] = useState(false);

  return (
    <article
      className={[
        styles.card,
        event.event === "item/commandExecution/requestApproval" ? styles.approval : "",
      ].join(" ")}
    >
      <EventHeader
        bodyID={bodyID}
        event={event}
        isBodyOpen={isBodyOpen}
        onToggleBody={() => setIsBodyOpen((current) => !current)}
      />
      {isBodyOpen ? (
        <div id={bodyID} className={styles.body}>
          <EventBody event={event} />
        </div>
      ) : (
        <div id={bodyID} className={styles.body}>
          <EventBodyPreview event={event} />
        </div>
      )}
    </article>
  );
}

function EventBody({ event }: { event: OrchestratorConversationEvent }) {
  if (event.event === "turn_completed") {
    return <TurnCompletedEventBody event={event} />;
  }

  if (event.event === "item/completed") {
    return <ItemCompletedEventBody event={event} />;
  }

  if (event.event === "item/commandExecution/requestApproval") {
    return <ApprovalRequestEventBody event={event} />;
  }

  if (event.event === "thread/tokenUsage/updated") {
    return <TokenUsageEventBody event={event} />;
  }

  if (event.event === "account/rateLimits/updated") {
    return <RateLimitsEventBody event={event} />;
  }

  return <StatusEventBody event={event} />;
}
