import assert from "node:assert/strict";
import { createRequire } from "node:module";
import test from "node:test";
import { DefaultChatTransport } from "ai";
import { createMayaCloudClient } from "./cloud-client.ts";
import { createMayaSessionGuard, type MayaAuthState } from "./session-scope.ts";

const scope = {
  userId: "user-one",
  workspace: "workspace-one",
  sessionEpoch: 2,
};
const credential = {
  userId: scope.userId,
  apiOrigin: "https://api.example.test",
  cookie: "fortyone_session=opaque-test-session",
};
const createFixture = (
  fetch: typeof globalThis.fetch,
  getSession = async () => credential,
) => {
  let state: MayaAuthState = {
    ...scope,
    isAuthenticated: true,
    isLoading: false,
  };
  const guard = createMayaSessionGuard(scope, () => state);
  const expired: string[] = [];
  const client = createMayaCloudClient({
    applicationURL: new URL("https://cloud.example.test"),
    apiOrigin: credential.apiOrigin,
    scope,
    assertCurrent: guard.assertCurrent,
    isCurrent: guard.isCurrent,
    getSession,
    expireSession: async (cookie) => {
      expired.push(cookie);
    },
    fetch,
  });
  return {
    client,
    expired,
    changeSession: () => {
      state = { ...state, sessionEpoch: 3 };
    },
  };
};

test("cloud credentials only reach exact allowlisted endpoints with explicit origin and no ambient cookies", async () => {
  const calls: RequestInit[] = [];
  const { client } = createFixture(async (_input, init) => {
    calls.push(init!);
    return new Response("ok");
  });
  const response = await client.fetch(client.url("/api/chat"), {
    method: "POST",
    body: "{}",
    headers: { Cookie: "wrong", Authorization: "wrong" },
  });
  assert.equal(await response.text(), "ok");
  assert.equal(calls[0].credentials, "omit");
  assert.equal(calls[0].redirect, "error");
  assert.equal(new Headers(calls[0].headers).get("cookie"), credential.cookie);
  assert.equal(
    new Headers(calls[0].headers).get("origin"),
    "https://cloud.example.test",
  );
  assert.equal(new Headers(calls[0].headers).get("authorization"), null);
  for (const path of [
    "https://attacker.test/api/chat",
    "https://cloud.example.test/api/chat?next=x",
    "https://cloud.example.test/api/other",
    "https://name@cloud.example.test/api/chat",
  ]) {
    await assert.rejects(
      client.fetch(path, { method: "POST" }),
      /cannot send your session/,
    );
  }
  assert.equal(calls.length, 1);
});

test("session changes during SecureStore lookup cannot send a request", async () => {
  let release!: (value: typeof credential) => void;
  const stored = new Promise<typeof credential>((resolve) => {
    release = resolve;
  });
  let calls = 0;
  const { client, changeSession } = createFixture(
    async () => {
      calls++;
      return new Response("ok");
    },
    () => stored,
  );
  const request = client.fetch(client.url("/api/chat"), { method: "POST" });
  changeSession();
  release(credential);
  await assert.rejects(request, /session changed/);
  assert.equal(calls, 0);
});

test("a late unauthorized response cannot clear the next session", async () => {
  let release!: (response: Response) => void;
  const response = new Promise<Response>((resolve) => {
    release = resolve;
  });
  const { client, changeSession, expired } = createFixture(
    async () => response,
  );
  const request = client.fetch(client.url("/api/chat"), { method: "POST" });
  await Promise.resolve();
  changeSession();
  release(new Response("Unauthorized", { status: 401 }));
  await assert.rejects(request, /session changed/);
  assert.deepEqual(expired, []);
});

test("stream body consumption remains scoped after response headers", async () => {
  let controller!: ReadableStreamDefaultController<Uint8Array>;
  const { client, changeSession } = createFixture(
    async () =>
      new Response(
        new ReadableStream({
          start(value) {
            controller = value;
          },
        }),
      ),
  );
  const response = await client.fetch(client.url("/api/chat"), {
    method: "POST",
  });
  const result = response.body!.getReader().read();
  changeSession();
  controller.enqueue(new TextEncoder().encode("private old-session text"));
  await assert.rejects(result, /session changed/);
});

test("failed transport calls are never automatically retried", async () => {
  let calls = 0;
  const { client } = createFixture(async () => {
    calls++;
    return new Response("Temporarily unavailable", { status: 503 });
  });
  await assert.rejects(
    client.fetch(client.url("/api/chat"), { method: "POST" }),
    /Temporarily unavailable/,
  );
  assert.equal(calls, 1);
});

test("AI SDK consumes SSE with React Native's actual non-streaming Response global", async () => {
  const require = createRequire(import.meta.url);
  const requireFromReactNative = createRequire(
    require.resolve("react-native/package.json"),
  );
  const { Response: ReactNativeResponse } = requireFromReactNative(
    "whatwg-fetch",
  ) as { Response: typeof Response };
  const StandardResponse = globalThis.Response;
  const chunks = [
    { type: "start", messageId: "assistant-one" },
    { type: "text-start", id: "text-one" },
    { type: "text-delta", id: "text-one", delta: "Hello 🌍" },
    { type: "text-end", id: "text-one" },
    { type: "finish", finishReason: "stop" },
  ];
  const upstream = new StandardResponse(
    chunks.map((chunk) => `data: ${JSON.stringify(chunk)}\n\n`).join("") +
      "data: [DONE]\n\n",
    { headers: { "Content-Type": "text/event-stream" } },
  );
  // This is the constructor RN installs and Expo intentionally leaves intact.
  assert.equal(new ReactNativeResponse(upstream.body).body, undefined);
  const { client } = createFixture(async () => upstream);
  globalThis.Response = ReactNativeResponse;
  try {
    const transport = new DefaultChatTransport({
      api: client.url("/api/chat"),
      fetch: client.fetch,
    });
    const stream = await transport.sendMessages({
      chatId: "abcdefghijklmnop",
      abortSignal: undefined,
      trigger: "submit-message",
      messageId: undefined,
      messages: [],
    });
    const reader = stream.getReader();
    const received = [];
    while (true) {
      const next = await reader.read();
      if (next.done) break;
      received.push(next.value);
    }
    assert.deepEqual(received, chunks);
  } finally {
    globalThis.Response = StandardResponse;
    client.dispose();
  }
});

test("native response metadata and JSON methods survive without eagerly locking its body", async () => {
  const upstream = new Response('{"messages":[]}', {
    status: 202,
    statusText: "Accepted",
    headers: { "Content-Type": "application/json", "X-Receipt": "received" },
  });
  Object.defineProperty(upstream, "url", {
    value: "https://cloud.example.test/api/chat/voice-history",
  });
  const { client } = createFixture(async () => upstream);
  const response = await client.fetch(client.url("/api/chat/voice-history"), {
    method: "POST",
  });
  assert.equal(response.url, upstream.url);
  assert.equal(response.status, 202);
  assert.equal(response.statusText, "Accepted");
  assert.equal(response.headers.get("X-Receipt"), "received");
  assert.equal(response.bodyUsed, false);
  assert.equal(upstream.body?.locked, false);
  const clone = response.clone();
  assert.deepEqual(await response.json(), { messages: [] });
  assert.equal(response.bodyUsed, true);
  assert.equal(await clone.text(), '{"messages":[]}');
  await assert.rejects(response.text(), /already used/);
});

test("native JSON completion is rejected if the session changed during consumption", async () => {
  let release!: (value: unknown) => void;
  const result = new Promise((resolve) => {
    release = resolve;
  });
  const upstream = new Response("{}");
  upstream.json = async function () {
    assert.equal(this, upstream);
    return result;
  };
  const { client, changeSession } = createFixture(async () => upstream);
  const response = await client.fetch(client.url("/api/chat/voice-history"), {
    method: "POST",
  });
  const pending = response.json();
  changeSession();
  release({ messages: ["old-session content"] });
  await assert.rejects(pending, /session changed/);
});

test("abort after headers rejects late native completion even if the upstream ignores abort", async () => {
  let release!: (value: unknown) => void;
  const result = new Promise((resolve) => {
    release = resolve;
  });
  const upstream = new Response("{}");
  upstream.json = async () => result;
  const { client } = createFixture(async () => upstream);
  const response = await client.fetch(client.url("/api/chat/voice-history"), {
    method: "POST",
  });
  const pending = response.json();
  client.abort();
  release({ messages: [] });
  await assert.rejects(pending, /abort|cancelled/i);
});

test("Strict Mode reattachment retains a guard; final release retires it", async () => {
  const state = { ...scope, isAuthenticated: true, isLoading: false };
  const guard = createMayaSessionGuard(scope, () => state);
  const releaseFirst = guard.retain();
  releaseFirst();
  const releaseSecond = guard.retain();
  await Promise.resolve();
  assert.equal(guard.isCurrent(), true);
  releaseSecond();
  await Promise.resolve();
  assert.equal(guard.isCurrent(), false);
});

test("transcription preserves multipart data and leaves its boundary to Expo", async () => {
  let outgoing: RequestInit | undefined;
  const form = new FormData();
  form.append(
    "audio",
    new Blob(["audio"], { type: "audio/m4a" }),
    "recording.m4a",
  );
  const { client } = createFixture(async (_input, init) => {
    outgoing = init;
    return new Response(JSON.stringify({ text: "A dictated message" }));
  });
  const response = await client.fetch(client.url("/api/transcribe"), {
    method: "POST",
    body: form,
    headers: {
      "Content-Type": "wrong-boundary",
      Cookie: "wrong",
      Authorization: "wrong",
    },
  });
  assert.deepEqual(await response.json(), { text: "A dictated message" });
  assert.equal(outgoing?.body, form);
  assert.equal(outgoing?.credentials, "omit");
  assert.equal(outgoing?.redirect, "error");
  const headers = new Headers(outgoing?.headers);
  assert.equal(headers.get("content-type"), null);
  assert.equal(headers.get("cookie"), credential.cookie);
  assert.equal(headers.get("origin"), "https://cloud.example.test");
  assert.equal(headers.get("authorization"), null);
});

test("transcription rejects arbitrary destinations and non-multipart payloads before upload", async () => {
  let calls = 0;
  const { client } = createFixture(async () => {
    calls++;
    return new Response("{}");
  });
  await assert.rejects(
    client.fetch("https://attacker.test/api/transcribe", {
      method: "POST",
      body: new FormData(),
    }),
    /cannot send your session/,
  );
  await assert.rejects(
    client.fetch(client.url("/api/transcribe"), { method: "POST", body: "{}" }),
    /requires an audio upload/,
  );
  assert.equal(calls, 0);
});

test("transcription preserves server string errors and never retries the upload", async () => {
  let calls = 0;
  const { client } = createFixture(async () => {
    calls++;
    return new Response(JSON.stringify({ error: "Audio file too large" }), {
      status: 413,
    });
  });
  await assert.rejects(
    client.fetch(client.url("/api/transcribe"), {
      method: "POST",
      body: new FormData(),
    }),
    /Audio file too large/,
  );
  assert.equal(calls, 1);
});

test("a cancelled transcription cannot return a late native result", async () => {
  let release!: (value: unknown) => void;
  const result = new Promise((resolve) => {
    release = resolve;
  });
  const upstream = new Response("{}");
  upstream.json = async () => result;
  const { client } = createFixture(async () => upstream);
  const controller = new AbortController();
  const response = await client.fetch(client.url("/api/transcribe"), {
    method: "POST",
    body: new FormData(),
    signal: controller.signal,
  });
  const pending = response.json();
  controller.abort();
  release({ text: "Old dictation" });
  await assert.rejects(pending, /abort|cancelled/i);
});

test("a transcription upload cannot complete into a different account session", async () => {
  let release!: (response: Response) => void;
  const upstream = new Promise<Response>((resolve) => {
    release = resolve;
  });
  const { client, changeSession } = createFixture(async () => upstream);
  const request = client.fetch(client.url("/api/transcribe"), {
    method: "POST",
    body: new FormData(),
  });
  await Promise.resolve();
  changeSession();
  release(
    new Response(JSON.stringify({ text: "Previous account's dictation" })),
  );
  await assert.rejects(request, /session changed/);
});
