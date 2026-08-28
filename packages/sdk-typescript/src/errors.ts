import type { components } from "./generated/schema.js";

export type ErrorField = NonNullable<
  components["schemas"]["Error"]["fields"]
>[number];
export type ErrorResponse = components["schemas"]["ErrorResponse"];

export class FortyOneApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId: string | undefined;
  readonly fields: readonly ErrorField[];
  readonly retryAfterSeconds: number | undefined;

  constructor(input: {
    status: number;
    code: string;
    message: string;
    requestId?: string;
    fields?: readonly ErrorField[];
    retryAfterSeconds?: number;
  }) {
    super(input.message);
    this.name = "FortyOneApiError";
    this.status = input.status;
    this.code = input.code;
    this.requestId = input.requestId;
    this.fields = Object.freeze(
      (input.fields ?? []).map((field) => Object.freeze({ ...field })),
    );
    this.retryAfterSeconds = input.retryAfterSeconds;
  }
}

export const apiErrorFromResponse = (
  response: Response,
  payload: unknown,
): FortyOneApiError => {
  const envelope = isErrorResponse(payload) ? payload.error : undefined;
  const requestId =
    envelope?.requestId ?? response.headers.get("X-Request-ID") ?? undefined;
  const retryAfterSeconds = parseRetryAfterSeconds(
    response.headers.get("Retry-After"),
  );
  return new FortyOneApiError({
    status: response.status,
    code: envelope?.code ?? "unexpected_response",
    message:
      envelope?.message ??
      `FortyOne API request failed with HTTP ${response.status}`,
    ...(requestId ? { requestId } : {}),
    ...(envelope?.fields ? { fields: envelope.fields } : {}),
    ...(retryAfterSeconds === undefined ? {} : { retryAfterSeconds }),
  });
};

const isErrorResponse = (value: unknown): value is ErrorResponse => {
  if (!value || typeof value !== "object" || !("error" in value)) {
    return false;
  }
  const error = value.error;
  return (
    !!error &&
    typeof error === "object" &&
    "code" in error &&
    typeof error.code === "string" &&
    "message" in error &&
    typeof error.message === "string" &&
    "requestId" in error &&
    typeof error.requestId === "string"
  );
};

const parseRetryAfterSeconds = (value: string | null) => {
  if (!value || !/^\d+$/u.test(value.trim())) {
    return undefined;
  }
  const seconds = Number(value.trim());
  return Number.isSafeInteger(seconds) && seconds > 0 ? seconds : undefined;
};
