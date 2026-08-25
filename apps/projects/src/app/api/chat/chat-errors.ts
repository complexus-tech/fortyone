const RATE_LIMIT_ERROR_CODE = "rate_limit_exceeded";
const RATE_LIMIT_MESSAGE =
  "Maya hit a temporary AI rate limit. If a change was already confirmed as completed, do not repeat it. Wait a few seconds, then continue the chat.";
const GENERIC_STREAM_ERROR_MESSAGE =
  "Maya couldn't finish the response. Please try again in a moment.";

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

const findErrorRecord = (value: unknown): ErrorRecord | undefined => {
  const record = asRecord(value);
  if (!record) return undefined;

  return asRecord(record.error) ?? record;
};

export const getChatStreamErrorMessage = (error: unknown) => {
  const parsedError = parseErrorMessage(error);
  const errorRecord = findErrorRecord(parsedError);
  const code = errorRecord?.code;
  const message = errorRecord?.message;

  if (
    code === RATE_LIMIT_ERROR_CODE ||
    (typeof message === "string" && /rate limit/i.test(message)) ||
    (typeof parsedError === "string" && /rate limit/i.test(parsedError))
  ) {
    return RATE_LIMIT_MESSAGE;
  }

  return GENERIC_STREAM_ERROR_MESSAGE;
};
