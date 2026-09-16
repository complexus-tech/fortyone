import assert from "node:assert/strict";
import { createRequire } from "node:module";
import test from "node:test";
import type { FileUIPart } from "ai";
import type { MayaMessage } from "../types";
import { createMayaChatRuntime } from "./chat-runtime.ts";
import { createMayaSessionGuard } from "./session-scope.ts";
import { createMayaCloudClient } from "./cloud-client.ts";
import { assertMayaRequestNotAborted } from "./abort.ts";

const scope = { userId: "user", workspace: "workspace", sessionEpoch: 1 };
const pdfFile: FileUIPart = {
  type: "file",
  filename: "brief.pdf",
  mediaType: "application/pdf",
  url: "data:application/pdf;base64,JVBERi0xLjcK",
};
const textResponse = () =>
  new Response(
    [
      { type: "start", messageId: "assistant-response" },
      { type: "text-start", id: "text" },
      { type: "text-delta", id: "text", delta: "Hello" },
      { type: "text-end", id: "text" },
      { type: "finish", finishReason: "stop" },
    ]
      .map((chunk) => `data: ${JSON.stringify(chunk)}\n\n`)
      .join("") + "data: [DONE]\n\n",
    {
      headers: {
        "Content-Type": "text/event-stream",
        "x-vercel-ai-ui-message-stream": "v1",
      },
    },
  );
const approvalMessage = (ids = ["approval-one"]): MayaMessage => ({
  id: "assistant-approval",
  role: "assistant",
  parts: ids.map((id) => ({
    type: "tool-createStory",
    toolCallId: `call-${id}`,
    state: "approval-requested",
    input: { title: "A task" },
    approval: { id },
  })),
});
const fixture = ({
  messages = [] as MayaMessage[],
  response = textResponse,
  history = async () => [] as MayaMessage[],
} = {}) => {
  let active = true;
  const guard = createMayaSessionGuard(scope, () => ({
    ...scope,
    isAuthenticated: active,
    isLoading: false,
  }));
  const requests: Record<string, unknown>[] = [];
  const cloud = createMayaCloudClient({
    applicationURL: new URL("https://cloud.example.test"),
    apiOrigin: "https://api.example.test",
    scope,
    assertCurrent: guard.assertCurrent,
    isCurrent: guard.isCurrent,
    getSession: async () => ({
      userId: scope.userId,
      apiOrigin: "https://api.example.test",
      cookie: "fortyone_session=test",
    }),
    expireSession: async () => undefined,
    fetch: async (_input, init) => {
      requests.push(JSON.parse(String(init?.body)));
      return response();
    },
  });
  const runtime = createMayaChatRuntime({
    id: "abcdefghijklmnop",
    workspace: scope.workspace,
    messages,
    cloud,
    guard,
    getHistory: history,
    onFinish: () => undefined,
  });
  return {
    runtime,
    requests,
    changeSession: () => {
      active = false;
    },
  };
};

test("text uses canonical UI-message SSE and overrides client-supplied scope", async () => {
  const { runtime, requests } = fixture();
  await runtime.send("Hi", {
    workspace: { slug: "wrong" },
    client: "wrong",
    timezone: "Africa/Harare",
  });
  assert.equal(requests.length, 1);
  assert.equal(requests[0].client, "mobile");
  assert.deepEqual(requests[0].workspace, { slug: scope.workspace });
  assert.equal(runtime.chat.messages.at(-1)?.parts[0].type, "text");
  assert.equal(runtime.chat.status, "ready");
});

test("an attachment-only user message sends its PDF once and remains visible in local history", async () => {
  const { runtime, requests } = fixture();
  await runtime.send("   ", {}, [pdfFile]);
  const uploaded = requests[0].messages as MayaMessage[];
  assert.deepEqual(uploaded[0].parts, [pdfFile]);
  assert.deepEqual(runtime.chat.messages[0].parts, [pdfFile]);
  assert.equal(runtime.chat.status, "ready");
  await runtime.send("What is the summary?", {});
  const followup = requests[1].messages as MayaMessage[];
  assert.equal(JSON.stringify(followup).includes(pdfFile.url), false);
  assert.equal(followup.at(-1)?.parts[0].type, "text");
});

test("only a fresh user upload is sent when historical file attachments exist", async () => {
  const oldFile = {
    ...pdfFile,
    filename: "old.pdf",
    url: "data:application/pdf;base64,b2xk",
  };
  const { runtime, requests } = fixture({
    messages: [{ id: "old-user", role: "user", parts: [oldFile] }],
  });
  await runtime.send("Review the new file", {}, [pdfFile]);
  const payload = requests[0].messages as MayaMessage[];
  assert.equal(JSON.stringify(payload).includes(oldFile.url), false);
  assert.equal(JSON.stringify(payload).includes(pdfFile.url), true);
  assert.equal(payload[0].id, "old-user");
});

test("approval continuation never re-uploads the source attachment", async () => {
  const { runtime, requests } = fixture({
    messages: [
      { id: "user-file", role: "user", parts: [pdfFile] },
      approvalMessage(),
    ],
  });
  await runtime.approve("approval-one", true, {});
  assert.equal(requests.length, 1);
  assert.equal(
    JSON.stringify(requests[0].messages).includes(pdfFile.url),
    false,
  );
  const approval = (requests[0].messages as MayaMessage[]).at(-1)?.parts[0];
  assert.equal(
    approval && "state" in approval ? approval.state : undefined,
    "approval-responded",
  );
});

test("invalid attachment sources are rejected before context lookup or network submission", async () => {
  const { runtime, requests } = fixture();
  let contextReads = 0;
  const body = async () => {
    contextReads++;
    return {};
  };
  for (const url of [
    "file:///private/brief.pdf",
    "https://example.test/brief.pdf",
    "data:image/svg+xml;base64,PHN2Zz4=",
  ]) {
    await assert.rejects(runtime.send("Review", body, [{ ...pdfFile, url }]));
  }
  await assert.rejects(
    runtime.send(" ", body, []),
    /Write a message or attach/,
  );
  assert.equal(contextReads, 0);
  assert.equal(requests.length, 0);
  assert.equal(runtime.chat.messages.length, 0);
});

test("loading a pending approval never executes it and ordinary sends are blocked", async () => {
  const { runtime, requests } = fixture({ messages: [approvalMessage()] });
  await Promise.resolve();
  assert.equal(requests.length, 0);
  await assert.rejects(runtime.send("Another task", {}), /Confirm or cancel/);
  await assert.rejects(
    runtime.approve("unknown", true, {}),
    /no longer pending/,
  );
  assert.equal(requests.length, 0);
});

test("multiple decisions submit only once after every approval is answered", async () => {
  const { runtime, requests } = fixture({
    messages: [approvalMessage(["one", "two"])],
  });
  await runtime.approve("one", true, {});
  assert.equal(requests.length, 0);
  await runtime.approve("two", false, {});
  assert.equal(requests.length, 1);
  const message = (requests[0].messages as MayaMessage[])[0];
  assert.deepEqual(
    message.parts.map((part) => "approval" in part && part.approval?.approved),
    [true, false],
  );
});

test("failed approval execution reloads canonical history and never replays a completed mutation", async () => {
  const completed: MayaMessage = {
    id: "assistant-approval",
    role: "assistant",
    parts: [
      {
        type: "tool-createStory",
        toolCallId: "call-approval-one",
        state: "output-available",
        input: { title: "A task" },
        output: { success: true },
        approval: { id: "approval-one", approved: true },
      },
    ],
  };
  let loads = 0;
  const { runtime, requests } = fixture({
    messages: [approvalMessage()],
    response: () => new Response("Lost receipt", { status: 503 }),
    history: async () => {
      loads++;
      return [completed];
    },
  });
  await assert.rejects(
    runtime.approve("approval-one", true, {}),
    /Lost receipt/,
  );
  await assert.rejects(
    runtime.approve("approval-one", true, {}),
    /no longer pending/,
  );
  assert.equal(loads, 1);
  assert.equal(requests.length, 1);
});

test("immediate cancellation fences work queued before credentials or network begin", async () => {
  const { runtime, requests } = fixture();
  const send = runtime.send("Hi", {});
  const rejection = assert.rejects(send, /cancelled/);
  await runtime.stop();
  await rejection;
  assert.equal(requests.length, 0);
  await runtime.send("A fresh message", {});
  assert.equal(requests.length, 1);
});

test("leaving Maya while context loads aborts it and prevents a late chat POST", async () => {
  const { runtime, requests } = fixture();
  let started!: () => void;
  const loading = new Promise<void>((resolve) => {
    started = resolve;
  });
  let aborted = false;
  const send = runtime.send(
    "Hello",
    (signal) =>
      new Promise((_, reject) => {
        signal.addEventListener(
          "abort",
          () => {
            aborted = true;
            reject(new Error("Context cancelled"));
          },
          { once: true },
        );
        started();
      }),
  );
  const rejection = assert.rejects(send, /Context cancelled/);
  await loading;
  await runtime.stop();
  await rejection;
  assert.equal(aborted, true);
  assert.equal(requests.length, 0);
  assert.equal(runtime.chat.messages.length, 0);
});

test("legacy workspace context survives request preparation without changing its captured slug", async () => {
  const { runtime, requests } = fixture();
  await runtime.send("Hello", async () => ({
    workspace: {
      id: "workspace-id",
      name: "A workspace",
      userRole: "member",
      slug: "wrong",
    },
    terminology: { stories: "tasks" },
    memories: [],
    username: "person",
  }));
  assert.deepEqual(requests[0].workspace, {
    id: "workspace-id",
    name: "A workspace",
    userRole: "member",
    slug: scope.workspace,
  });
  assert.deepEqual(requests[0].terminology, { stories: "tasks" });
  assert.deepEqual(requests[0].memories, []);
});

test("failed context loading leaves a pending approval available for an explicit retry", async () => {
  const { runtime, requests } = fixture({ messages: [approvalMessage()] });
  await assert.rejects(
    runtime.approve("approval-one", true, async () => {
      throw new Error("Settings unavailable");
    }),
    /Settings unavailable/,
  );
  assert.equal(requests.length, 0);
  await runtime.approve("approval-one", true, {});
  assert.equal(requests.length, 1);
});

test("first-request failure before reservation can recover a missing new chat", async () => {
  let available = false;
  const { runtime, requests } = fixture({
    response: () =>
      available ? textResponse() : new Response("Unavailable", { status: 503 }),
    history: async () => {
      throw Object.assign(new Error("Missing"), { status: 404 });
    },
  });
  await assert.rejects(runtime.send("Keep my draft", {}), /Unavailable/);
  available = true;
  await runtime.send("Keep my draft", {});
  assert.equal(requests.length, 2);
  assert.equal((requests[1].messages as MayaMessage[]).length, 1);
});

test("recovery never recreates a previously persisted but deleted conversation", async () => {
  const { runtime, requests } = fixture({
    messages: [
      {
        id: "user-first",
        role: "user",
        parts: [{ type: "text", text: "Existing history" }],
      },
    ],
    response: () => new Response("Unavailable", { status: 503 }),
    history: async () => {
      throw Object.assign(new Error("Conversation deleted"), { status: 404 });
    },
  });
  await assert.rejects(runtime.send("Next", {}), /Unavailable/);
  await assert.rejects(runtime.send("Next", {}), /Conversation deleted/);
  assert.equal(requests.length, 1);
});

test("a disposed controller clears private messages and cannot send under another session", async () => {
  const { runtime, requests, changeSession } = fixture({
    messages: [approvalMessage()],
  });
  changeSession();
  runtime.dispose();
  assert.deepEqual(runtime.chat.messages, []);
  assert.throws(
    () => runtime.approve("approval-one", true, {}),
    /session changed/,
  );
  assert.equal(requests.length, 0);
});

test("Strict Mode retain/release does not destroy an immediately reattached conversation", async () => {
  const { runtime } = fixture({ messages: [approvalMessage()] });
  runtime.retain()();
  const release = runtime.retain();
  await Promise.resolve();
  assert.equal(runtime.chat.messages.length, 1);
  release();
  await Promise.resolve();
  assert.equal(runtime.chat.messages.length, 0);
});

test(
  "the installed RN globals support the complete SDK send path with fragmented Expo-style UTF-8 streaming",
  { concurrency: false, timeout: 5_000 },
  async () => {
    const require = createRequire(import.meta.url);
    const requireFromReactNative = createRequire(
      require.resolve("react-native/package.json"),
    );
    const nativeAbort = requireFromReactNative(
      "abort-controller/dist/abort-controller",
    ) as Pick<typeof globalThis, "AbortController" | "AbortSignal">;
    const nativeFetch = requireFromReactNative("whatwg-fetch") as Pick<
      typeof globalThis,
      "Response" | "Request" | "Headers"
    >;
    const globals = {
      AbortController: nativeAbort.AbortController,
      AbortSignal: nativeAbort.AbortSignal,
      Response: nativeFetch.Response,
      Request: nativeFetch.Request,
      Headers: nativeFetch.Headers,
    };
    const descriptors = Object.fromEntries(
      Object.keys(globals).map((key) => [
        key,
        Object.getOwnPropertyDescriptor(globalThis, key)!,
      ]),
    );
    const NodeResponse = globalThis.Response;
    const text = "Café 🌍 — ready";
    const bytes = new TextEncoder().encode(
      [
        { type: "start", messageId: "native-assistant" },
        { type: "text-start", id: "native-text" },
        { type: "text-delta", id: "native-text", delta: text },
        { type: "text-end", id: "native-text" },
        { type: "finish", finishReason: "stop" },
      ]
        .map((chunk) => `data: ${JSON.stringify(chunk)}\n\n`)
        .join("") + "data: [DONE]\n\n",
    );
    let offset = 0;
    // Expo returns a native streaming response even though RN's global Response
    // constructor only implements buffered fetch. Capture the upstream first.
    const upstream = new NodeResponse(
      new ReadableStream<Uint8Array>({
        pull(controller) {
          if (offset === bytes.length) {
            controller.close();
            return;
          }
          // One-byte chunks necessarily split both the SSE frames and UTF-8 glyphs.
          controller.enqueue(bytes.subarray(offset, ++offset));
        },
      }),
      {
        headers: {
          "Content-Type": "text/event-stream",
          "x-vercel-ai-ui-message-stream": "v1",
        },
      },
    );
    let runtime: ReturnType<typeof createMayaChatRuntime> | undefined;
    try {
      Object.assign(globalThis, globals);
      assert.equal(new Response("buffered").body, undefined);
      const setup = fixture({ response: () => upstream });
      runtime = setup.runtime;
      await runtime.send("Hello", async (signal) => {
        assert.equal("throwIfAborted" in signal, false);
        assertMayaRequestNotAborted(signal);
        await Promise.resolve();
        assertMayaRequestNotAborted(signal);
        return {
          timezone: "Africa/Harare",
          memories: [],
          terminology: { stories: "tasks" },
        };
      });
      assert.equal(setup.requests.length, 1);
      assert.equal(runtime.chat.status, "ready");
      assert.equal(runtime.chat.error, undefined);
      const assistant = runtime.chat.messages.at(-1);
      assert.equal(assistant?.id, "native-assistant");
      assert.deepEqual(
        assistant?.parts
          .filter((part) => part.type === "text")
          .map((part) => ({ text: part.text, state: part.state })),
        [{ text, state: "done" }],
      );
    } finally {
      runtime?.dispose();
      Object.defineProperties(globalThis, descriptors);
    }
  },
);
