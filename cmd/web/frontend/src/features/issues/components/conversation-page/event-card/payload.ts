export type Translator = (key: string, options?: Record<string, unknown>) => string;

export type TokenUsageBucket = {
  totalTokens: number;
  inputTokens: number;
  cachedInputTokens: number;
  outputTokens: number;
  reasoningOutputTokens: number;
};

export type TokenUsageView = {
  total: TokenUsageBucket;
  last: TokenUsageBucket;
  modelContextWindow: number;
};

export type RateLimitWindow = {
  usedPercent: number;
  windowDurationMins: number;
  resetsAt: number;
};

export type RateLimitsView = {
  limitId: string;
  planType: string;
  primary: RateLimitWindow;
  secondary: RateLimitWindow;
  rateLimitReachedType: string;
};

export type ItemCompletedView = {
  type: string;
  command: string;
  aggregatedOutput: string;
  text: string;
  exitCode: number | null;
};

export type ApprovalRequestDetails = {
  reason: string;
  command: string;
};

export function tokenUsageView(payloadJSON: string): TokenUsageView | null {
  const payload = parsePayloadJSON(payloadJSON);
  if (!isRecord(payload) || !isRecord(payload.tokenUsage)) {
    return null;
  }
  const total = tokenUsageBucket(payload.tokenUsage.total);
  const last = tokenUsageBucket(payload.tokenUsage.last);
  const modelContextWindow = findNumberField(payload.tokenUsage, "modelContextWindow");
  if (!total || !last || modelContextWindow === null) {
    return null;
  }
  return { total, last, modelContextWindow };
}

export function rateLimitsView(payloadJSON: string): RateLimitsView | null {
  const payload = parsePayloadJSON(payloadJSON);
  if (!isRecord(payload) || !isRecord(payload.rateLimits)) {
    return null;
  }
  const primary = rateLimitWindow(payload.rateLimits.primary);
  const secondary = rateLimitWindow(payload.rateLimits.secondary);
  const limitId = findStringField(payload.rateLimits, "limitId");
  const planType = findStringField(payload.rateLimits, "planType");
  if (!primary || !secondary || limitId === "" || planType === "") {
    return null;
  }
  return {
    limitId,
    planType,
    primary,
    secondary,
    rateLimitReachedType: findStringField(payload.rateLimits, "rateLimitReachedType"),
  };
}

export function formatRateLimitWindow(window: RateLimitWindow, t: Translator): string {
  return t("issues.detailPage.rateLimitWindow", {
    percent: window.usedPercent,
    minutes: formatNumber(window.windowDurationMins),
    resetsAt: formatUnixSeconds(window.resetsAt),
  });
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat().format(value);
}

export function extractAggregatedOutput(payloadJSON: string): string {
  const payload = parsePayloadJSON(payloadJSON);
  return typeof payload === "string" ? payload : findAggregatedOutput(payload);
}

export function extractApprovalRequestDetails(payloadJSON: string): ApprovalRequestDetails {
  const payload = parsePayloadJSON(payloadJSON);
  if (typeof payload === "string") {
    return { reason: "", command: "" };
  }
  return {
    reason: findStringField(payload, "reason"),
    command: findStringField(payload, "command"),
  };
}

export function itemCompletedView(payloadJSON: string): ItemCompletedView {
  const payload = parsePayloadJSON(payloadJSON);
  const item = typeof payload === "string" ? null : findItemPayload(payload);
  const source = item ?? (isRecord(payload) ? payload : null);

  return {
    type: findStringField(source, "type") || "item/completed",
    command: findStringField(source, "command"),
    aggregatedOutput: source ? findAggregatedOutput(source) : "",
    text: findStringField(source, "text"),
    exitCode: findNumberField(source, "exitCode"),
  };
}

function tokenUsageBucket(value: unknown): TokenUsageBucket | null {
  if (!isRecord(value)) {
    return null;
  }
  const totalTokens = findNumberField(value, "totalTokens");
  const inputTokens = findNumberField(value, "inputTokens");
  const cachedInputTokens = findNumberField(value, "cachedInputTokens");
  const outputTokens = findNumberField(value, "outputTokens");
  const reasoningOutputTokens = findNumberField(value, "reasoningOutputTokens");
  if (
    totalTokens === null ||
    inputTokens === null ||
    cachedInputTokens === null ||
    outputTokens === null ||
    reasoningOutputTokens === null
  ) {
    return null;
  }
  return {
    totalTokens,
    inputTokens,
    cachedInputTokens,
    outputTokens,
    reasoningOutputTokens,
  };
}

function rateLimitWindow(value: unknown): RateLimitWindow | null {
  if (!isRecord(value)) {
    return null;
  }
  const usedPercent = findNumberField(value, "usedPercent");
  const windowDurationMins = findNumberField(value, "windowDurationMins");
  const resetsAt = findNumberField(value, "resetsAt");
  if (usedPercent === null || windowDurationMins === null || resetsAt === null) {
    return null;
  }
  return { usedPercent, windowDurationMins, resetsAt };
}

function formatUnixSeconds(value: number): string {
  const date = new Date(value * 1000);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function parsePayloadJSON(payloadJSON: string): unknown {
  if (payloadJSON.trim() === "") {
    return {};
  }
  try {
    return JSON.parse(payloadJSON) as unknown;
  } catch {
    return payloadJSON;
  }
}

function findAggregatedOutput(value: unknown): string {
  if (typeof value === "string") {
    return "";
  }
  if (!value || typeof value !== "object") {
    return "";
  }
  if ("aggregatedOutput" in value && typeof value.aggregatedOutput === "string") {
    return value.aggregatedOutput;
  }
  if ("aggregated_output" in value && typeof value.aggregated_output === "string") {
    return value.aggregated_output;
  }
  for (const child of Object.values(value)) {
    const output = findAggregatedOutput(child);
    if (output !== "") {
      return output;
    }
  }
  return "";
}

function findItemPayload(value: unknown): Record<string, unknown> | null {
  if (!isRecord(value)) {
    return null;
  }
  if (isRecord(value.item)) {
    return value.item;
  }
  if (typeof value.type === "string") {
    return value;
  }
  for (const child of Object.values(value)) {
    const item = findItemPayload(child);
    if (item) {
      return item;
    }
  }
  return null;
}

function findStringField(value: unknown, field: string): string {
  if (!isRecord(value)) {
    return "";
  }
  const direct = value[field];
  if (typeof direct === "string") {
    return direct;
  }
  for (const child of Object.values(value)) {
    const result = findStringField(child, field);
    if (result !== "") {
      return result;
    }
  }
  return "";
}

function findNumberField(value: unknown, field: string): number | null {
  if (!isRecord(value)) {
    return null;
  }
  const direct = value[field];
  if (typeof direct === "number" && Number.isFinite(direct)) {
    return direct;
  }
  for (const child of Object.values(value)) {
    const result = findNumberField(child, field);
    if (result !== null) {
      return result;
    }
  }
  return null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
