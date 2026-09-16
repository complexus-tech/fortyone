import type { MayaSessionScope } from "./session-scope";
import { asRecord } from "./chat-protocol";

export type MayaCloudPath =
  | "/api/chat"
  | "/api/chat/voice-history"
  | "/api/transcribe";
type StoredCredential = { cookie: string; userId: string; apiOrigin: string };
type CloudClientOptions = {
  applicationURL: URL;
  apiOrigin: string;
  scope: MayaSessionScope;
  assertCurrent: () => void;
  isCurrent: () => boolean;
  getSession: () => Promise<StoredCredential | null>;
  expireSession: (cookie: string) => Promise<void>;
  fetch: typeof globalThis.fetch;
};

export class MayaRequestError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "MayaRequestError";
  }
}

const CLOUD_PATHS = new Set<string>([
  "/api/chat",
  "/api/chat/voice-history",
  "/api/transcribe",
]);
const REQUEST_TIMEOUT_MS = 280_000;
const TRANSCRIPTION_TIMEOUT_MS = 60_000;
const BODY_METHODS = new Set<PropertyKey>([
  "arrayBuffer",
  "blob",
  "bytes",
  "formData",
  "json",
  "text",
]);

/** Expo streams, but React Native's global Response constructor does not. */
const scopeResponse = (
  source: Response,
  lifecycle: {
    assertCurrent: () => void;
    retain: () => () => void;
    abort: (reason?: unknown) => void;
  },
): Response => {
  const complete = lifecycle.retain();
  let consumed = false;
  let reader: ReadableStreamDefaultReader<Uint8Array> | undefined;
  const assertUnused = () => {
    if (consumed || source.bodyUsed || body.locked)
      throw new TypeError("Maya response body is already used.");
  };
  const body = new ReadableStream<Uint8Array>(
    {
      async pull(destination) {
        try {
          lifecycle.assertCurrent();
          if (!reader) {
            if (consumed || source.bodyUsed)
              throw new TypeError("Maya response body is already used.");
            consumed = true;
            reader = source.body!.getReader();
          }
          const result = await reader.read();
          lifecycle.assertCurrent();
          if (result.done) {
            complete();
            destination.close();
          } else destination.enqueue(result.value);
        } catch (error) {
          complete();
          lifecycle.abort(error);
          void reader?.cancel().catch(() => undefined);
          destination.error(error);
        }
      },
      async cancel(reason) {
        consumed = true;
        complete();
        lifecycle.abort(reason);
        if (reader) await reader.cancel(reason);
        else await source.body?.cancel(reason);
      },
    },
    // Do not lock the native body merely by exposing it: json()/text() must
    // remain usable until the caller actually chooses stream consumption.
    { highWaterMark: 0 },
  );
  return new Proxy(source, {
    get(target, property) {
      if (property === "body") return body;
      if (property === "bodyUsed") return consumed || target.bodyUsed;
      if (property === "clone")
        return () => {
          lifecycle.assertCurrent();
          assertUnused();
          return scopeResponse(target.clone(), lifecycle);
        };
      // Native body methods need their original receiver and implementation.
      // In particular, Expo json() delegates to its native text() internally.
      const value = Reflect.get(target, property, target);
      if (typeof value !== "function") return value;
      if (!BODY_METHODS.has(property)) return value.bind(target);
      return async (...args: unknown[]) => {
        assertUnused();
        consumed = true;
        try {
          lifecycle.assertCurrent();
          const result: unknown = await value.apply(target, args);
          lifecycle.assertCurrent();
          return result;
        } catch (error) {
          lifecycle.abort(error);
          throw error;
        } finally {
          complete();
        }
      };
    },
  });
};

export const assertMayaCloudURL = (input: string, applicationURL: URL) => {
  const target = new URL(input);
  if (
    target.origin !== applicationURL.origin ||
    target.username ||
    target.password ||
    target.search ||
    target.hash ||
    !CLOUD_PATHS.has(target.pathname)
  )
    throw new Error("Maya cannot send your session to this address.");
  return target;
};

const responseError = async (response: Response) => {
  const text = await response.text();
  let message = text;
  try {
    const body = asRecord(JSON.parse(text));
    const error = asRecord(body.error);
    message =
      typeof body.error === "string"
        ? body.error
        : typeof error.message === "string"
          ? error.message
          : typeof body.message === "string"
            ? body.message
            : "";
  } catch {
    // Chat errors use plain text; HTML from a gateway is not useful in the app.
    if (text.includes("<")) message = "";
  }
  return new MayaRequestError(
    message.trim().slice(0, 500) ||
      "Maya could not complete this request. Try opening the conversation again.",
    response.status,
  );
};

/** Dedicated cloud transport: no redirects, ambient cookie jar, retries, or arbitrary URLs. */
export const createMayaCloudClient = (options: CloudClientOptions) => {
  const requests = new Set<AbortController>();
  let disposed = false;
  const assertCurrent = () => {
    if (disposed)
      throw new Error("This Maya conversation is no longer active.");
    options.assertCurrent();
  };

  const fetch: typeof globalThis.fetch = async (input, init = {}) => {
    const inputURL =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.href
          : input.url;
    const target = assertMayaCloudURL(inputURL, options.applicationURL);
    if (init.method?.toUpperCase() !== "POST")
      throw new Error("Unsupported Maya request.");
    const multipart = target.pathname === "/api/transcribe";
    if (multipart && !(init.body instanceof FormData))
      throw new Error("Maya transcription requires an audio upload.");
    assertCurrent();
    const session = await options.getSession();
    assertCurrent();
    if (
      !session ||
      session.userId !== options.scope.userId ||
      session.apiOrigin !== options.apiOrigin
    ) {
      throw new MayaRequestError("Sign in again before using Maya.", 401);
    }
    const controller = new AbortController();
    const onAbort = () => controller.abort(init.signal?.reason);
    if (init.signal?.aborted) onAbort();
    else init.signal?.addEventListener("abort", onAbort, { once: true });
    requests.add(controller);
    const timeout = setTimeout(
      () =>
        controller.abort(
          new Error(
            "Maya took too long to respond. Open the conversation to check its latest state.",
          ),
        ),
      multipart ? TRANSCRIPTION_TIMEOUT_MS : REQUEST_TIMEOUT_MS,
    );
    const cleanup = () => {
      clearTimeout(timeout);
      requests.delete(controller);
      init.signal?.removeEventListener("abort", onAbort);
      controller.signal.removeEventListener("abort", cleanup);
    };
    controller.signal.addEventListener("abort", cleanup, { once: true });
    const assertResponseCurrent = () => {
      assertCurrent();
      if (controller.signal.aborted)
        throw controller.signal.reason ?? new Error("Maya request cancelled.");
    };

    try {
      assertResponseCurrent();
      const response = await options.fetch(target.href, {
        method: "POST",
        body: init.body,
        signal: controller.signal,
        credentials: "omit",
        redirect: "error",
        headers: {
          // Expo supplies the native multipart boundary for audio uploads.
          ...(multipart ? {} : { "Content-Type": "application/json" }),
          Accept:
            target.pathname === "/api/chat"
              ? "text/event-stream"
              : "application/json",
          Origin: options.applicationURL.origin,
          Cookie: session.cookie,
        },
      });
      assertResponseCurrent();
      if (!response.ok) {
        if (response.status === 401 && options.isCurrent())
          await options.expireSession(session.cookie);
        throw await responseError(response);
      }
      if (!response.body) throw new Error("Maya returned an empty response.");
      // Keep the abort controller alive through body consumption, not just headers.
      let bodies = 0;
      return scopeResponse(response, {
        assertCurrent: assertResponseCurrent,
        abort: (reason) => {
          cleanup();
          controller.abort(reason);
        },
        retain: () => {
          bodies++;
          let completed = false;
          return () => {
            if (completed) return;
            completed = true;
            if (--bodies === 0) cleanup();
          };
        },
      });
    } catch (error) {
      cleanup();
      controller.abort();
      throw error;
    }
  };

  return {
    fetch,
    url: (path: MayaCloudPath) => new URL(path, options.applicationURL).href,
    abort: () => {
      for (const request of requests) request.abort();
    },
    dispose: () => {
      disposed = true;
      for (const request of requests) request.abort();
    },
  };
};
