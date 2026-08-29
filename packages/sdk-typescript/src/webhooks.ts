const SIGNING_SECRET_PREFIX = "whsec_";
const SIGNATURE_VERSION = "v1";
const SIGNATURE_BYTES = 32;
const MAX_BODY_BYTES = 256 * 1024;
const DEFAULT_TOLERANCE_SECONDS = 5 * 60;
const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/iu;

export type WebhookVerificationErrorCode =
  | "invalid_secret"
  | "invalid_headers"
  | "invalid_timestamp"
  | "stale_timestamp"
  | "invalid_signature"
  | "invalid_body";

export class WebhookVerificationError extends Error {
  readonly code: WebhookVerificationErrorCode;

  constructor(code: WebhookVerificationErrorCode, message: string) {
    super(message);
    this.name = "WebhookVerificationError";
    this.code = code;
  }
}

export interface VerifyWebhookInput {
  secret: string;
  body: Uint8Array;
  webhookId: string;
  webhookTimestamp: string;
  webhookSignature: string;
  now?: Date;
  toleranceSeconds?: number;
}

export interface VerifiedWebhook {
  id: string;
  timestamp: Date;
}

export const verifyWebhook = async (
  input: VerifyWebhookInput,
): Promise<VerifiedWebhook> => {
  if (!UUID_PATTERN.test(input.webhookId)) {
    throw new WebhookVerificationError(
      "invalid_headers",
      "Webhook-Id is missing or malformed",
    );
  }
  if (input.body.length === 0 || input.body.length > MAX_BODY_BYTES) {
    throw new WebhookVerificationError(
      "invalid_body",
      "Webhook body is empty or exceeds the supported limit",
    );
  }
  const timestampSeconds = parseTimestamp(input.webhookTimestamp);
  const timestamp = new Date(timestampSeconds * 1_000);
  const now = input.now ?? new Date();
  const toleranceSeconds =
    input.toleranceSeconds ?? DEFAULT_TOLERANCE_SECONDS;
  if (
    !Number.isFinite(toleranceSeconds) ||
    toleranceSeconds < 1 ||
    toleranceSeconds > 3_600
  ) {
    throw new WebhookVerificationError(
      "invalid_timestamp",
      "Webhook timestamp tolerance must be from 1 through 3600 seconds",
    );
  }
  if (
    Math.abs(Math.floor(now.getTime() / 1_000) - timestampSeconds) >
    toleranceSeconds
  ) {
    throw new WebhookVerificationError(
      "stale_timestamp",
      "Webhook timestamp is outside the accepted replay window",
    );
  }

  const key = decodeSigningSecret(input.secret);
  const signedContent = concatenate(
    new TextEncoder().encode(
      `${input.webhookId}.${input.webhookTimestamp}.`,
    ),
    input.body,
  );
  let cryptoKey: CryptoKey;
  try {
    cryptoKey = await globalThis.crypto.subtle.importKey(
      "raw",
      key,
      { name: "HMAC", hash: "SHA-256" },
      false,
      ["verify"],
    );
  } finally {
    key.fill(0);
  }
  for (const candidate of input.webhookSignature.trim().split(/\s+/u)) {
    const [version, encoded, extra] = candidate.split(",");
    if (version !== SIGNATURE_VERSION || !encoded || extra !== undefined) {
      continue;
    }
    const signature = decodeBase64(encoded);
    if (
      signature?.length === SIGNATURE_BYTES &&
      (await globalThis.crypto.subtle.verify(
        "HMAC",
        cryptoKey,
        signature,
        signedContent,
      ))
    ) {
      return { id: input.webhookId, timestamp };
    }
  }
  throw new WebhookVerificationError(
    "invalid_signature",
    "Webhook signature is invalid",
  );
};

const parseTimestamp = (value: string) => {
  if (!/^\d{1,12}$/u.test(value)) {
    throw new WebhookVerificationError(
      "invalid_timestamp",
      "Webhook-Timestamp is missing or malformed",
    );
  }
  const seconds = Number(value);
  if (!Number.isSafeInteger(seconds) || seconds <= 0) {
    throw new WebhookVerificationError(
      "invalid_timestamp",
      "Webhook-Timestamp is missing or malformed",
    );
  }
  return seconds;
};

const decodeSigningSecret = (secret: string) => {
  if (!secret.startsWith(SIGNING_SECRET_PREFIX)) {
    throw new WebhookVerificationError(
      "invalid_secret",
      "Webhook signing secret is malformed",
    );
  }
  const decoded = decodeBase64(secret.slice(SIGNING_SECRET_PREFIX.length));
  if (decoded?.length !== SIGNATURE_BYTES) {
    throw new WebhookVerificationError(
      "invalid_secret",
      "Webhook signing secret is malformed",
    );
  }
  return decoded;
};

const decodeBase64 = (value: string) => {
  if (
    value.length === 0 ||
    value.length % 4 !== 0 ||
    !/^[A-Za-z0-9+/]+={0,2}$/u.test(value)
  ) {
    return undefined;
  }
  try {
    const decoded = Uint8Array.from(globalThis.atob(value), (character) =>
      character.charCodeAt(0),
    );
    const canonical = globalThis.btoa(
      Array.from(decoded, (byte) => String.fromCharCode(byte)).join(""),
    );
    return canonical === value ? decoded : undefined;
  } catch {
    return undefined;
  }
};

const concatenate = (prefix: Uint8Array, body: Uint8Array) => {
  const result = new Uint8Array(prefix.length + body.length);
  result.set(prefix);
  result.set(body, prefix.length);
  return result;
};
