import { Suspense } from "react";
import { ConversationPage } from "./_components/conversation-page";

export default function ConversationRoute() {
  return (
    <Suspense fallback={null}>
      <ConversationPage />
    </Suspense>
  );
}
