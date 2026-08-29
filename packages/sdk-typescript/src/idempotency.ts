const MINIMUM_KEY_BYTES = 16;
const MAXIMUM_KEY_BYTES = 255;

// Generate this once per logical mutation and persist it until the operation's
// retry lifecycle is complete. A fresh key represents a new operation.
export const createIdempotencyKey = (
  cryptoProvider: Pick<Crypto, "getRandomValues"> = globalThis.crypto,
): string => {
  if (!cryptoProvider?.getRandomValues) {
    throw new Error("A cryptographically secure random source is required");
  }
  const bytes = cryptoProvider.getRandomValues(new Uint8Array(32));
  return Array.from(bytes, (value) => value.toString(16).padStart(2, "0")).join(
    "",
  );
};

export const validateIdempotencyKey = (value: string): void => {
  const encoded = new TextEncoder().encode(value);
  if (
    encoded.length < MINIMUM_KEY_BYTES ||
    encoded.length > MAXIMUM_KEY_BYTES ||
    !/^[!-~]+$/u.test(value)
  ) {
    throw new Error(
      "FortyOne idempotency key must contain 16 to 255 visible ASCII characters",
    );
  }
};
