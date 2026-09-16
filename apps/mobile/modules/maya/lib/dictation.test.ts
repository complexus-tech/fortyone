import assert from "node:assert/strict";
import test from "node:test";
import { createMayaDictation, isCurrentDictationEvent } from "./dictation.ts";

test("delayed native completions cannot finish a newer recording", () => {
  assert.equal(
    isCurrentDictationEvent({ url: "old.m4a" }, "new.m4a", false),
    false,
  );
  assert.equal(isCurrentDictationEvent({ url: null }, "new.m4a", true), false);
  assert.equal(
    isCurrentDictationEvent({ url: "new.m4a" }, "new.m4a", true),
    false,
  );
  assert.equal(
    isCurrentDictationEvent({ url: "new.m4a" }, "new.m4a", false),
    true,
  );
  assert.equal(isCurrentDictationEvent({ url: null }, "new.m4a", false), true);
});

const deferred = <T>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};
const fixture = (
  overrides: Partial<Parameters<typeof createMayaDictation>[0]> = {},
) => {
  const events: string[] = [];
  const texts: string[] = [];
  const controller = createMayaDictation({
    getScopeKey: () => "chat-one",
    isCurrent: () => true,
    permission: async () => {
      events.push("permission");
      return true;
    },
    prepare: async () => {
      events.push("prepare");
    },
    record: () => {
      events.push("record");
    },
    stop: async () => {
      events.push("stop");
      return "file:///owned/recording.m4a";
    },
    release: () => {
      events.push("release");
    },
    transcribe: async () => {
      events.push("transcribe");
      return " Dictated words ";
    },
    onText: (text) => {
      events.push("text");
      texts.push(text);
    },
    ...overrides,
  });
  return { controller, events, texts };
};

test("permission is requested only on explicit start; transcript arrives after microphone and file cleanup", async () => {
  const { controller, events, texts } = fixture();
  assert.deepEqual(events, []);
  await controller.start();
  assert.equal(controller.getSnapshot().status, "recording");
  await controller.finish();
  assert.deepEqual(events, [
    "permission",
    "prepare",
    "record",
    "stop",
    "transcribe",
    "release",
    "text",
  ]);
  assert.deepEqual(texts, ["Dictated words"]);
  assert.equal(controller.getSnapshot().status, "idle");
});

test("cancel during permission cannot start the microphone after permission resolves", async () => {
  const permission = deferred<boolean>();
  const { controller, events, texts } = fixture({
    permission: () => permission.promise,
  });
  const start = controller.start();
  const cancel = controller.cancel();
  permission.resolve(true);
  await Promise.all([start, cancel]);
  assert.equal(events.includes("record"), false);
  assert.equal(events.includes("prepare"), false);
  assert.deepEqual(texts, []);
  assert.equal(controller.getSnapshot().status, "idle");
});

test("cancel waits for pending native prepare before final mic cleanup and another start", async () => {
  const prepare = deferred<void>();
  const { controller, events } = fixture({ prepare: () => prepare.promise });
  const start = controller.start();
  await Promise.resolve();
  const cancel = controller.cancel();
  await assert.rejects(controller.start(), /current dictation/);
  assert.equal(events.includes("stop"), false);
  prepare.resolve();
  await Promise.all([start, cancel]);
  assert.equal(events.includes("record"), false);
  assert.equal(events.filter((event) => event === "stop").length, 1);
  assert.equal(controller.getSnapshot().status, "idle");
});

test("cancelling transcription cleans promptly and a late reply cannot enter a new recording", async () => {
  const transcript = deferred<string>();
  const entered = deferred<void>();
  let signal: AbortSignal | undefined;
  const { controller, texts, events } = fixture({
    transcribe: async (_uri, abort) => {
      signal = abort;
      entered.resolve();
      return transcript.promise;
    },
  });
  await controller.start();
  const finish = controller.finish();
  await entered.promise;
  await controller.cancel();
  assert.equal(signal?.aborted, true);
  assert.equal(controller.getSnapshot().status, "idle");
  await controller.start();
  transcript.resolve("Old conversation text");
  await finish;
  assert.equal(controller.getSnapshot().status, "recording");
  assert.deepEqual(texts, []);
  assert.equal(events.filter((event) => event === "release").length, 1);
  await controller.cancel();
});

test("a changed chat invalidates a provider reply even without an explicit cancel", async () => {
  let chat = "chat-one";
  const transcript = deferred<string>();
  const entered = deferred<void>();
  const { controller, texts, events } = fixture({
    getScopeKey: () => chat,
    transcribe: async () => {
      entered.resolve();
      return transcript.promise;
    },
  });
  await controller.start();
  const finish = controller.finish();
  await entered.promise;
  chat = "chat-two";
  transcript.resolve("Stale text");
  await assert.rejects(finish, /cancelled/);
  assert.deepEqual(texts, []);
  assert.equal(events.includes("release"), true);
});

test("permission denial and provider failure surface errors without publishing text", async () => {
  const denied = fixture({ permission: async () => false });
  await assert.rejects(denied.controller.start(), /microphone access/);
  assert.equal(denied.events.includes("record"), false);
  assert.equal(denied.controller.getSnapshot().status, "idle");
  const failed = fixture({
    transcribe: async () => {
      throw new Error("Service unavailable");
    },
  });
  await failed.controller.start();
  await assert.rejects(failed.controller.finish(), /Service unavailable/);
  assert.deepEqual(failed.texts, []);
  assert.equal(failed.events.includes("release"), true);
  assert.equal(failed.controller.getSnapshot().error, "Service unavailable");
});

test("the 60-second boundary finishes exactly once", async () => {
  const { controller, events, texts } = fixture();
  await controller.start();
  await controller.progress(59);
  assert.equal(controller.getSnapshot().seconds, 59);
  await Promise.all([
    controller.progress(60),
    controller.progress(61),
    controller.finish(),
  ]);
  assert.equal(events.filter((event) => event === "transcribe").length, 1);
  assert.equal(texts.length, 1);
});

test("native cutoff transcribes without a UI timer tick and waits for audio cleanup", async () => {
  const audioCleanup = deferred<string | null>();
  const uri = "file:///owned/recording.m4a";
  let stopCalls = 0;
  const { controller, events, texts } = fixture({
    stop: () => {
      stopCalls += 1;
      return audioCleanup.promise;
    },
  });
  await controller.start();

  // Native forDuration stops capture independently of JS. Its completion can
  // arrive before the progress interval ever runs (for example, a busy JS thread).
  assert.equal(controller.getSnapshot().seconds, 0);
  assert.equal(isCurrentDictationEvent({ url: uri }, uri, false), true);
  const completed = controller.finish();
  assert.equal(controller.getSnapshot().status, "transcribing");
  assert.equal(events.includes("transcribe"), false);
  assert.deepEqual(texts, []);

  // A delayed timer or manual tap must not submit the same recording again.
  await Promise.all([controller.progress(75), controller.finish()]);
  assert.equal(stopCalls, 1);
  audioCleanup.resolve(uri);
  await completed;
  assert.equal(events.filter((event) => event === "transcribe").length, 1);
  assert.deepEqual(texts, ["Dictated words"]);
  assert.equal(controller.getSnapshot().status, "idle");
});

test("a delayed UI cutoff remains a fallback when native completion is delayed", async () => {
  const transcript = deferred<string>();
  const { controller, events, texts } = fixture({
    transcribe: () => {
      events.push("transcribe");
      return transcript.promise;
    },
  });
  await controller.start();
  await controller.progress(59.99);
  assert.equal(controller.getSnapshot().status, "recording");

  const completed = controller.progress(63);
  assert.equal(controller.getSnapshot().seconds, 60);
  assert.equal(controller.getSnapshot().status, "transcribing");
  // Completion and repeated progress events cannot start another transcription.
  await Promise.all([controller.finish(), controller.progress(64)]);
  transcript.resolve("Completed at the recording limit");
  await completed;
  assert.equal(events.filter((event) => event === "stop").length, 1);
  assert.equal(events.filter((event) => event === "transcribe").length, 1);
  assert.deepEqual(texts, ["Completed at the recording limit"]);
});

test("a cleanup failure keeps microphone ownership locked until explicit cancellation succeeds", async () => {
  let fails = true;
  const { controller } = fixture({
    stop: async () => {
      if (fails) throw new Error("Microphone cleanup failed");
      return "file:///owned/recording.m4a";
    },
  });
  await controller.start();
  await assert.rejects(controller.cancel(), /cleanup failed/);
  await assert.rejects(controller.start(), /current dictation/);
  fails = false;
  await controller.cancel();
  assert.equal(controller.getSnapshot().status, "idle");
});
