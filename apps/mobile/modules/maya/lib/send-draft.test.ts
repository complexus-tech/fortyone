import assert from "node:assert/strict";
import test from "node:test";
import type { FileUIPart } from "ai";
import { sendMayaDraft } from "./send-draft.ts";

const deferred = <T>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
};
const file: FileUIPart = {
  type: "file",
  mediaType: "application/pdf",
  filename: "brief.pdf",
  url: "data:application/pdf;base64,JVBERg==",
};
const fixture = (
  overrides: Partial<Parameters<typeof sendMayaDraft>[0]> = {},
) => {
  let current = true;
  const calls: string[] = [];
  const options = {
    text: "Review this",
    prepare: async () => {
      calls.push("prepare");
      return [file];
    },
    flushVoiceHistory: async () => {
      calls.push("flush");
    },
    send: async (text: string, files: FileUIPart[]) => {
      assert.equal(text, "Review this");
      assert.deepEqual(files, [file]);
      calls.push("send");
    },
    isCurrent: () => current,
    onSent: () => {
      calls.push("clear");
    },
    ...overrides,
  };
  return {
    options,
    calls,
    cancel: () => {
      current = false;
    },
  };
};

test("stop or back while attachments prepare prevents history flush and submission", async () => {
  const pending = deferred<FileUIPart[]>();
  const { options, calls, cancel } = fixture({
    prepare: () => pending.promise,
  });
  const result = sendMayaDraft(options);
  cancel();
  pending.resolve([file]);
  assert.equal(await result, false);
  assert.deepEqual(calls, []);
});

test("a conversation change during history flush prevents late submission", async () => {
  const pending = deferred<void>();
  const started = deferred<void>();
  const { options, calls, cancel } = fixture({
    flushVoiceHistory: () => {
      started.resolve();
      return pending.promise;
    },
  });
  const result = sendMayaDraft(options);
  await started.promise;
  cancel();
  pending.resolve();
  assert.equal(await result, false);
  assert.deepEqual(calls, ["prepare"]);
});

test("stale successful completion cannot clear a replacement draft", async () => {
  const pending = deferred<void>();
  const started = deferred<void>();
  const { options, calls, cancel } = fixture({
    send: () => {
      started.resolve();
      return pending.promise;
    },
  });
  const result = sendMayaDraft(options);
  await started.promise;
  cancel();
  pending.resolve();
  assert.equal(await result, false);
  assert.deepEqual(calls, ["prepare", "flush"]);
});

test("current send errors preserve the staged draft and remain actionable", async () => {
  const error = new Error("Connection unavailable");
  const { options, calls } = fixture({
    send: async () => {
      throw error;
    },
  });
  await assert.rejects(
    sendMayaDraft(options),
    (cause: unknown) => cause === error,
  );
  assert.deepEqual(calls, ["prepare", "flush"]);
});

test("a late preparation error after cancellation does not replace the new UI error", async () => {
  const pending = deferred<FileUIPart[]>();
  const { options, calls, cancel } = fixture({
    prepare: () => pending.promise,
  });
  const result = sendMayaDraft(options);
  cancel();
  pending.reject(new Error("Old file removed"));
  assert.equal(await result, false);
  assert.deepEqual(calls, []);
});

test("only a completed current send clears the draft after preparation and history", async () => {
  const { options, calls } = fixture();
  assert.equal(await sendMayaDraft(options), true);
  assert.deepEqual(calls, ["prepare", "flush", "send", "clear"]);
});
