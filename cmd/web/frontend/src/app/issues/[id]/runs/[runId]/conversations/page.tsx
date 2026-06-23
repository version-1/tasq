import { Suspense } from "react";
import { ConversationPage } from "@/features/issues/components/conversation-page";

export default function ConversationRoute() {
  return (
    <Suspense fallback={null}>
      <ConversationPage />
    </Suspense>
  );
}
