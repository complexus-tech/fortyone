import assert from "node:assert/strict";
import test from "node:test";
import {
  MayaVoiceController,
  type VoiceDependencies,
  type VoiceLease,
} from "./voice-controller.ts";
import type { VoiceEvent, VoiceToolResult } from "./voice-protocol.ts";

const deferred = <T>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
};
const flush = () => new Promise<void>((resolve) => setImmediate(resolve));

function fixture(overrides: Partial<VoiceDependencies> = {}) {
  let current = true;
  let callbacks!: Parameters<VoiceDependencies["openTransport"]>[2];
  const sent: unknown[] = [];
  const tools: Parameters<VoiceDependencies["tool"]>[1][] = [];
  const ended: string[] = [];
  let closes = 0;
  let invalidations = 0;
  const mute: boolean[] = [];
  const lease = {
    sessionId: "session-1",
    clientSecret: "provider-secret",
    maxSessionSeconds: 300,
  };
  const deps: VoiceDependencies = {
    captureIdentity: async () => ({
      userId: "user-1",
      workspace: "workspace-1",
      sessionEpoch: 1,
      cookie: "private-cookie",
    }),
    isCurrent: () => current,
    startSession: async () => lease,
    endSession: async (_identity, id) => {
      ended.push(id);
    },
    tool: async (_identity, call) => {
      tools.push(call);
      return call.arguments.confirmed
        ? { success: true }
        : {
            requiresConfirmation: true,
            confirmationToken: "human-only-token",
            confirmation: { title: "Review me" },
          };
    },
    openTransport: async (_lease, _signal, handlers) => {
      callbacks = handlers;
      callbacks.onOpen();
      return {
        send: (value) => sent.push(value),
        close: () => {
          closes++;
        },
        mute: (value) => {
          mute.push(value);
        },
      };
    },
    invalidate: () => {
      invalidations++;
    },
    clientAction: async () => true,
    ...overrides,
  };
  const controller = new MayaVoiceController(deps);
  return {
    controller,
    tools,
    sent,
    ended,
    mute,
    get closes() {
      return closes;
    },
    get invalidations() {
      return invalidations;
    },
    event: (event: VoiceEvent) => callbacks.onEvent(event),
    changeIdentity: () => {
      current = false;
      controller.checkIdentity();
    },
  };
}
function callEvent(
  callId = "call-1",
  name = "create_task",
  args = { title: "Review me", confirmed: true, confirmationToken: "invented" },
): VoiceEvent {
  return {
    type: "response.done",
    response: {
      output: [
        {
          type: "function_call",
          call_id: callId,
          name,
          arguments: JSON.stringify(args),
        },
      ],
    },
  };
}

test("the opening response preserves the server's identity and session instructions", async (t) => {
  const f = fixture();
  t.after(f.controller.stop);

  await f.controller.start();

  assert.equal(f.controller.getSnapshot().status, "connected");
  assert.deepEqual(f.sent, [{ type: "response.create" }]);
});

test("only the human approval method can send a write confirmation", async (t) => {
  const f = fixture();
  t.after(f.controller.stop);
  await f.controller.start();
  f.event(callEvent());
  await flush();
  assert.equal(f.tools.length, 1);
  assert.deepEqual(f.tools[0].arguments, {
    title: "Review me",
    confirmed: false,
  });
  assert.equal(f.controller.getSnapshot().pendingAction?.title, "Create task");
  assert.doesNotMatch(
    JSON.stringify([f.sent, f.controller.getSnapshot()]),
    /human-only-token|invented|private-cookie|provider-secret/,
  );
  await f.controller.approveAction();
  assert.deepEqual(f.tools[1].arguments, {
    title: "Review me",
    confirmed: true,
    confirmationToken: "human-only-token",
  });
  assert.equal(f.tools[1].callId, "call-1:approved");
  assert.equal(f.invalidations, 1);
  assert.equal(f.controller.getSnapshot().pendingAction, null);
  assert.doesNotMatch(JSON.stringify(f.sent), /human-only-token/);
});

test("an old native build fails before creating a paid provider session", async () => {
  const f = fixture({
    prepareTransport: () => {
      throw new Error("Rebuild required");
    },
    startSession: async () => {
      throw new Error("must not create a session");
    },
  });
  await f.controller.start();
  assert.equal(f.controller.getSnapshot().status, "idle");
  assert.equal(f.controller.getSnapshot().error, "Rebuild required");
  assert.deepEqual(f.ended, []);
});

test("cancellation and duplicate provider events never execute a mutation", async (t) => {
  const f = fixture();
  t.after(f.controller.stop);
  await f.controller.start();
  f.event(callEvent());
  f.event(callEvent());
  await flush();
  f.controller.cancelAction();
  await f.controller.approveAction();
  assert.equal(f.tools.length, 1);
  assert.equal(f.tools[0].arguments.confirmed, false);
  assert.match(JSON.stringify(f.sent), /cancelled/);
});

test("parallel proposals cannot replace the action being prepared", async (t) => {
  const result = deferred<VoiceToolResult>();
  let calls = 0;
  const f = fixture({
    tool: async () => {
      calls++;
      return result.promise;
    },
  });
  t.after(f.controller.stop);
  await f.controller.start();
  f.event(callEvent("first"));
  f.event(callEvent("second"));
  result.resolve({ requiresConfirmation: true, confirmationToken: "secret" });
  await flush();
  assert.equal(calls, 1);
  assert.equal(f.controller.getSnapshot().pendingAction?.id, "first");
});

test("late tool results cannot populate a new account or send over a new connection", async (t) => {
  const result = deferred<VoiceToolResult>();
  const f = fixture({ tool: async () => result.promise });
  t.after(f.controller.stop);
  await f.controller.start();
  f.event(callEvent());
  f.changeIdentity();
  const sent = f.sent.length;
  result.resolve({ requiresConfirmation: true, confirmationToken: "secret" });
  await flush();
  assert.equal(f.sent.length, sent);
  assert.equal(f.controller.getSnapshot().pendingAction, null);
  assert.equal(f.controller.getSnapshot().status, "idle");
  assert.equal(f.closes, 1);
});

test("a lease returned after stopping is ended using the captured identity", async () => {
  const result = deferred<VoiceLease>();
  const f = fixture({ startSession: async () => result.promise });
  const starting = f.controller.start();
  await flush();
  f.controller.stop();
  result.resolve({
    sessionId: "late-session",
    clientSecret: "secret",
    maxSessionSeconds: 300,
  });
  await starting;
  assert.deepEqual(f.ended, ["late-session"]);
  assert.equal(f.controller.getSnapshot().status, "idle");
});

test("late transport resolution is closed after cancellation", async () => {
  const opened =
    deferred<Awaited<ReturnType<VoiceDependencies["openTransport"]>>>();
  let closed = 0;
  const f = fixture({ openTransport: async () => opened.promise });
  const starting = f.controller.start();
  await flush();
  f.controller.stop();
  opened.resolve({
    send: () => assert.fail("cannot send after stop"),
    close: () => {
      closed++;
    },
    mute: () => undefined,
  });
  await starting;
  assert.equal(closed, 1);
  assert.deepEqual(f.ended, ["session-1"]);
});

test("lease deadline and idle deadline close the connection and end accounting", async (t) => {
  t.mock.timers.enable({ apis: ["setTimeout", "Date"], now: 0 });
  const f = fixture({
    startSession: async () => ({
      sessionId: "short",
      clientSecret: "secret",
      maxSessionSeconds: 2,
    }),
  });
  await f.controller.start();
  assert.equal(f.controller.getSnapshot().remainingSeconds, 2);
  t.mock.timers.tick(2000);
  assert.equal(f.controller.getSnapshot().status, "idle");
  assert.equal(f.closes, 1);
  assert.deepEqual(f.ended, ["short"]);
  const idle = fixture();
  await idle.controller.start();
  t.mock.timers.tick(60_000);
  assert.equal(idle.controller.getSnapshot().status, "idle");
  assert.equal(idle.closes, 1);
});

test("active voice stops at five minutes even with a longer server lease and delayed countdown ticks", async (t) => {
  t.mock.timers.enable({ apis: ["setTimeout", "Date"], now: 0 });
  const f = fixture({
    startSession: async () => ({
      sessionId: "long-lease",
      clientSecret: "secret",
      maxSessionSeconds: 3600,
    }),
  });
  t.after(f.controller.stop);
  await f.controller.start();
  assert.equal(f.controller.getSnapshot().remainingSeconds, 300);

  for (let elapsed = 30_000; elapsed <= 270_000; elapsed += 30_000) {
    t.mock.timers.tick(30_000);
    f.event({ type: "input_audio_buffer.speech_started" });
    assert.equal(f.controller.getSnapshot().status, "connected");
  }
  t.mock.timers.tick(29_999);
  assert.equal(f.controller.getSnapshot().status, "connected");
  assert.equal(f.controller.getSnapshot().remainingSeconds, 1);
  t.mock.timers.tick(1);
  assert.equal(f.controller.getSnapshot().status, "idle");
  assert.equal(f.closes, 1);
  assert.deepEqual(f.ended, ["long-lease"]);
  t.mock.timers.tick(60_000);
  assert.equal(f.closes, 1);
});

test("transcripts replace partial text and finalize once; muting changes native tracks", async (t) => {
  const f = fixture();
  t.after(f.controller.stop);
  const finalized: unknown[] = [];
  await f.controller.start({
    onTranscriptFinalized: (message) => {
      finalized.push(message);
    },
  });
  f.event({
    type: "conversation.item.input_audio_transcription.delta",
    item_id: "user",
    delta: "Hel",
  });
  const completed = {
    type: "conversation.item.input_audio_transcription.completed",
    item_id: "user",
    transcript: "Hello",
  };
  f.event(completed);
  f.event(completed);
  assert.equal(f.controller.getSnapshot().transcript[0].text, "Hello");
  assert.equal(f.controller.getSnapshot().transcript[0].order, 0);
  assert.equal(finalized.length, 1);
  f.controller.toggleMute();
  assert.equal(f.controller.getSnapshot().isMuted, true);
  f.controller.toggleMute();
  assert.equal(f.controller.getSnapshot().isMuted, false);
  assert.deepEqual(f.mute, [false, true, false]);
});

test("final callbacks retain conversation order when assistant transcription completes first", async (t) => {
  const f = fixture();
  t.after(f.controller.stop);
  const finalized: { id: string; order?: number }[] = [];
  await f.controller.start({
    onTranscriptFinalized: (message) => {
      finalized.push(message);
    },
  });
  for (const id of ["user", "assistant"])
    f.event({ type: "conversation.item.added", item: { id } });
  f.event({
    type: "response.output_audio_transcript.done",
    item_id: "assistant",
    transcript: "Your tasks are ready.",
  });
  f.event({
    type: "conversation.item.input_audio_transcription.completed",
    item_id: "user",
    transcript: "What should I work on?",
  });
  assert.deepEqual(
    finalized.map((message) => message.order),
    [1, 0],
  );
  assert.deepEqual(
    f.controller.getSnapshot().transcript.map((message) => message.role),
    ["user", "assistant"],
  );
  assert.ok(
    finalized.every((message) => message.id.startsWith("voice-session-1-")),
  );
});

test("lost approval replies retry the same exact idempotency key", async (t) => {
  const calls: Parameters<VoiceDependencies["tool"]>[1][] = [];
  const f = fixture({
    tool: async (_identity, input) => {
      calls.push(input);
      if (!input.arguments.confirmed)
        return { requiresConfirmation: true, confirmationToken: "token" };
      if (calls.length === 2) throw new Error("Network unavailable");
      return { success: true };
    },
  });
  t.after(f.controller.stop);
  await f.controller.start();
  f.event(callEvent());
  await flush();
  await f.controller.approveAction();
  await f.controller.approveAction();
  assert.deepEqual(calls[1], calls[2]);
  // Refresh after the uncertain result too, so checking the task reads current data.
  assert.equal(f.invalidations, 2);
});

test("completed transcripts are erased on an account switch even after stopping", async () => {
  const f = fixture();
  await f.controller.start();
  f.event({
    type: "conversation.item.input_audio_transcription.completed",
    item_id: "user",
    transcript: "Private work",
  });
  f.controller.stop();
  assert.equal(f.controller.getSnapshot().transcript.length, 1);
  f.changeIdentity();
  assert.equal(f.controller.getSnapshot().transcript.length, 0);
});
