type LogFields = Record<string, unknown>;

const SENSITIVE_KEYS = new Set([
  "authorization",
  "botToken",
  "cookie",
  "encryptionKey",
  "FORTYONE_BOT_TOKEN",
  "OPENAI_API_KEY",
  "SLACK_BOT_TOKEN",
  "SLACK_CLIENT_SECRET",
  "SLACK_SIGNING_SECRET",
  "token",
]);

const redact = (value: unknown): unknown => {
  if (!value || typeof value !== "object") {
    return value;
  }
  if (Array.isArray(value)) {
    return value.map(redact);
  }
  return Object.fromEntries(
    Object.entries(value as Record<string, unknown>).map(([key, entry]) => [
      key,
      SENSITIVE_KEYS.has(key) ? "[redacted]" : redact(entry),
    ]),
  );
};

export const errorMessage = (error: unknown) =>
  error instanceof Error ? error.message : String(error);

export const logBotError = (message: string, fields: LogFields = {}) => {
  const safeFields = redact(fields) as LogFields;

  // eslint-disable-next-line no-console -- Bot runtime logs JSON diagnostics to stderr.
  console.error(
    JSON.stringify({
      level: "error",
      message,
      ...safeFields,
    }),
  );
};
