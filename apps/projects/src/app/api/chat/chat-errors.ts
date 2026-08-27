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

const collectErrorDetails = (error: unknown) => {
  const codes = new Set<string>();
  const messages = new Set<string>();
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
    if (typeof record.code === "string") codes.add(record.code);
    if (typeof record.message === "string") messages.add(record.message);
    if (typeof record.status === "number") statuses.add(record.status);

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
  return { codes, statuses, text: Array.from(messages).join(" ") };
};

/**
 * Return operationally useful error metadata without serializing provider
 * request bodies, tool inputs, user content, or nested response payloads.
 */
export const getChatErrorDiagnostic = (error: unknown) => {
  const { codes, statuses } = collectErrorDetails(error);

  return {
    codes: Array.from(codes).slice(0, 5),
    errorType: error instanceof Error ? error.name : typeof error,
    statuses: Array.from(statuses).slice(0, 5),
  };
};

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
