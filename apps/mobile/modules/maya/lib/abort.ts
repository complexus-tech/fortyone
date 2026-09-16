type CancellationSignal = {
  readonly aborted: boolean;
  readonly reason?: unknown;
};

// React Native's AbortSignal does not implement throwIfAborted.
export const assertMayaRequestNotAborted = (signal: CancellationSignal) => {
  if (!signal.aborted) return;
  if (signal.reason !== undefined) throw signal.reason;
  const error = new Error("Maya request cancelled.");
  error.name = "AbortError";
  throw error;
};
