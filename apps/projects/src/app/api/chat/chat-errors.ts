const RATE_LIMIT_ERROR_CODE = "rate_limit_exceeded";
const RATE_LIMIT_MESSAGE =
  "Maya hit a temporary AI rate limit. If a change was already confirmed as completed, do not repeat it. Wait a few seconds, then continue the chat.";
const TIMEOUT_MESSAGE =
  "Maya's response took too long. Any completed change is still saved; ask Maya to verify the result before repeating it.";
const AUTHENTICATION_MESSAGE =
  "Your Maya session is no longer authorized. Refresh the page, sign in again if needed, then continue the chat.";
const GENERIC_STREAM_ERROR_MESSAGE =
  "Maya couldn't finish the response. Please try again in a moment.";
const REQUEST_TOO_LARGE_MESSAGE =
  "That message is too large for Maya. Remove an attachment or split the request into smaller parts, then try again.";

type ErrorRecord = Record<string, unknown>;

const MAX_DIAGNOSTIC_MESSAGE_LENGTH = 500;
const OPERATIONAL_ERROR_NAMES = new Set([
  "AI_APICallError",
  "AI_LoadAPIKeyError",
  "AI_RetryError",
  "APICallError",
  "ApiContractError",
  "ApiError",
  "HTTPError",
  "LoadAPIKeyError",
  "RetryError",
  "TimeoutError",
]);

const asRecord = (value: unknown): ErrorRecord | undefined =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as ErrorRecord)
    : undefined;

const parseErrorMessage = (error: unknown): unknown => {
  const message = error instanceof Error ? error.message : error;
  if (typeof message !== "string") return message;

  try {
    return JSON.parse(message) as unknown;
  } catch {
    return message;
  }
};

const sanitizeDiagnosticMessage = (message: string) =>
  message
    .replace(/data:[^;,\s]+;base64,[A-Za-z0-9+/=]+/g, "[inline data redacted]")
    .replace(/\b(?:sk|rk|pk|sess)-[A-Za-z0-9_-]{12,}\b/g, "[credential redacted]")
    .replace(
      /\b(?<field>authorization|api[-_ ]?key|password|secret|token)(?<separator>\s*[:=]\s*)(?<value>[^,\s]+)/gi,
      "$<field>$<separator>[redacted]",
    )
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, MAX_DIAGNOSTIC_MESSAGE_LENGTH);

const getErrorName = (value: unknown, record: ErrorRecord) => {
  if (value instanceof Error) return value.name;
  return typeof record.name === "string" ? record.name : undefined;
};

const collectApiErrorEnvelope = (
  value: unknown,
  codes: Set<string>,
  requestIds: Set<string>,
) => {
  const envelope = asRecord(value);
  const detail = asRecord(envelope?.error);
  if (!detail) return;

  if (typeof detail.code === "string") codes.add(detail.code);
  if (typeof detail.request_id === "string") requestIds.add(detail.request_id);
  if (typeof detail.requestId === "string") requestIds.add(detail.requestId);
};

const collectErrorDetails = (error: unknown) => {
  const codes = new Set<string>();
  const messages = new Set<string>();
  const operationalMessages = new Set<string>();
  const requestIds = new Set<string>();
  const retryable = new Set<boolean>();
  const statuses = new Set<number>();
  const visited = new Set<unknown>();

  const visit = (value: unknown, depth = 0) => {
    if (depth > 5 || value === null || value === undefined) return;
    if (typeof value === "string") {
      messages.add(value);
      const parsed = parseErrorMessage(value);
      if (parsed !== value) visit(parsed, depth + 1);
      return;
    }
    if (value instanceof Error) {
      messages.add(value.message);
    }
    if (visited.has(value)) return;
    visited.add(value);

    const record = asRecord(value);
    if (!record) return;
    const errorName = getErrorName(value, record);
    if (typeof record.code === "string") codes.add(record.code);
    if (typeof record.message === "string") messages.add(record.message);
    if (typeof record.status === "number") statuses.add(record.status);
    if (typeof record.statusCode === "number") statuses.add(record.statusCode);
    if (typeof record.isRetryable === "boolean") {
      retryable.add(record.isRetryable);
    }
    if (
      errorName &&
      OPERATIONAL_ERROR_NAMES.has(errorName) &&
      typeof record.message === "string"
    ) {
      const sanitizedMessage = sanitizeDiagnosticMessage(record.message);
      if (sanitizedMessage) operationalMessages.add(sanitizedMessage);
    }

    collectApiErrorEnvelope(record.data, codes, requestIds);

    const response = asRecord(record.response);
    if (typeof response?.status === "number") statuses.add(response.status);

    for (const key of ["error", "cause", "lastError", "errors"]) {
      const nested = record[key];
      if (Array.isArray(nested)) {
        for (const item of nested) {
          visit(item, depth + 1);
        }
      } else {
        visit(nested, depth + 1);
      }
    }
  };

  visit(error);
  return {
    codes,
    operationalMessages,
    requestIds,
    retryable,
    statuses,
    text: Array.from(messages).join(" "),
  };
};

/**
 * Return operationally useful error metadata without serializing provider
 * request bodies, tool inputs, user content, or nested response payloads.
 */
export const getChatErrorDiagnostic = (error: unknown) => {
  const {
    codes,
    operationalMessages,
    requestIds,
    retryable,
    statuses,
  } = collectErrorDetails(error);
  const messages = Array.from(operationalMessages).slice(0, 3);
  const providerRequestIds = Array.from(requestIds).slice(0, 3);
  const retryability = Array.from(retryable);

  return {
    codes: Array.from(codes).slice(0, 5),
    errorType: error instanceof Error ? error.name : typeof error,
    ...(messages.length > 0 ? { messages } : {}),
    ...(providerRequestIds.length > 0
      ? { requestIds: providerRequestIds }
      : {}),
    ...(retryability.length > 0 ? { retryable: retryability } : {}),
    statuses: Array.from(statuses).slice(0, 5),
  };
};

export const formatChatErrorDiagnostic = (error: unknown) =>
  JSON.stringify(getChatErrorDiagnostic(error));

export const getChatStreamErrorMessage = (error: unknown) => {
  const { codes, text } = collectErrorDetails(error);

  if (
    codes.has(RATE_LIMIT_ERROR_CODE) ||
    /(?:rate limit|too many requests|\b429\b)/i.test(text)
  ) {
    return RATE_LIMIT_MESSAGE;
  }

  if (/(?:timeout|timed out|deadline exceeded|request aborted)/i.test(text)) {
    return TIMEOUT_MESSAGE;
  }

  if (
    codes.has("unauthorized") ||
    /(?:unauthorized|authentication required|\b401\b)/i.test(text)
  ) {
    return AUTHENTICATION_MESSAGE;
  }

  if (
    codes.has("request_too_large") ||
    /(?:request|payload|message).{0,20}too large|\b413\b/i.test(text)
  ) {
    return REQUEST_TOO_LARGE_MESSAGE;
  }

  return GENERIC_STREAM_ERROR_MESSAGE;
};
