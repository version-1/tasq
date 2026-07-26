import { useId, useState } from "react";
import type { OrchestratorConversationEvent } from "@/lib/types";
import { ItemCompletedEventBody } from "./body/item-completed";
import { EventBodyPreview } from "./body/preview";
import { RateLimitsEventBody } from "./body/rate-limits";
import { StatusEventBody } from "./body/status";
import { TokenUsageEventBody } from "./body/token-usage";
import { EventHeader } from "./event-header";
import { itemCompletedView } from "./payload";
import styles from "./index.module.css";

export { ItemCompletedEventBody } from "./body/item-completed";
export { EventBodyPreview } from "./body/preview";
export { RateLimitsEventBody } from "./body/rate-limits";
export { StatusEventBody } from "./body/status";
export { TokenUsageEventBody } from "./body/token-usage";
export { EventHeader } from "./event-header";

export function EventCard({ event }: { event: OrchestratorConversationEvent }) {
  const bodyID = useId();
  const [isBodyOpen, setIsBodyOpen] = useState(() => isAgentMessage(event));
  return (
    <article className={styles.card}>
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

function isAgentMessage(event: OrchestratorConversationEvent): boolean {
  return event.event === "item/completed" && itemCompletedView(event.payload_json).type === "agentMessage";
}

function EventBody({ event }: { event: OrchestratorConversationEvent }) {
  if (event.event === "item/completed") {
    return <ItemCompletedEventBody event={event} />;
  }

  if (event.event === "thread/tokenUsage/updated") {
    return <TokenUsageEventBody event={event} />;
  }

  if (event.event === "account/rateLimits/updated") {
    return <RateLimitsEventBody event={event} />;
  }

  return <StatusEventBody event={event} />;
}
