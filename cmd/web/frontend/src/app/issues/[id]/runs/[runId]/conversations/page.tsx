"use client";

import { Navigate, useParams, useSearchParams } from "react-router-dom";

export default function ConversationRoute() {
  const { id, runId } = useParams();
  const [searchParams] = useSearchParams();
  const params = new URLSearchParams(searchParams);
  params.set("tab", "conversation");
  if (runId) {
    params.set("runId", runId);
  }

  return <Navigate to={`/issues/${id ?? ""}?${params.toString()}`} replace />;
}
