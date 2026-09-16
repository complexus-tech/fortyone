import assert from "node:assert/strict";
import test from "node:test";
import {
  createVoiceHistoryQueue,
  type VoiceHistoryMessage,
} from "./voice-history-queue.ts";

const prefix = "voice-20e867da-1266-45b8-978c-7ab900594b7a-";
const user: VoiceHistoryMessage = {
  id: `${prefix}user`,
  role: "user",
  text: "My work",
  order: 0,
};
const assistant: VoiceHistoryMessage = {
  id: `${prefix}assistant`,
  role: "assistant",
  text: "Two tasks",
  order: 1,
};
const wire = ({ id, role, text }: VoiceHistoryMessage) => ({ id, role, text });

test("assistant tail includes its exact previously saved user anchor", async () => {
  const batches: VoiceHistoryMessage[][] = [];
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async (batch) => {
      batches.push(batch);
    },
  });
  queue.add(user);
  await queue.flush(false);
  queue.add(assistant);
  await queue.flush(false);
  assert.deepEqual(batches, [[wire(user)], [wire(user), wire(assistant)]]);
  assert.equal(queue.hasPending, false);
});

test("out-of-order finalization waits for the user and respects provider turn order", async () => {
  const batches: VoiceHistoryMessage[][] = [];
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async (batch) => {
      batches.push(batch);
    },
  });
  queue.add(assistant);
  await queue.flush(false);
  assert.equal(batches.length, 0);
  queue.add(user);
  await queue.flush(false);
  assert.deepEqual(batches, [[wire(user), wire(assistant)]]);
});

test("an uncertain write keeps exact IDs and requires explicit retry", async () => {
  const batches: VoiceHistoryMessage[][] = [];
  let unavailable = true;
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async (batch) => {
      batches.push(batch);
      if (unavailable) throw new Error("Connection lost");
    },
  });
  queue.add(user);
  queue.add(assistant);
  await assert.rejects(queue.flush(false), /Connection lost/);
  assert.equal(queue.hasPending, true);
  await assert.rejects(queue.flush(false), /Connection lost/);
  assert.equal(batches.length, 1);
  unavailable = false;
  await queue.flush(true);
  assert.deepEqual(batches[1], batches[0]);
  assert.equal(queue.hasPending, false);
});

test("new voice session greeting never attaches to a previous session user", async () => {
  const batches: VoiceHistoryMessage[][] = [];
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async (batch) => {
      batches.push(batch);
    },
  });
  queue.add(user);
  queue.add(assistant);
  await queue.flush();
  queue.add({
    ...assistant,
    id: "voice-843a4bbd-b15b-491f-a03a-14d5f49c3cf6-greeting",
    order: 0,
  });
  await queue.flush(false);
  assert.equal(batches.length, 1);
});

test("messages arriving during a save are preserved for the next anchored batch", async () => {
  const batches: VoiceHistoryMessage[][] = [];
  let release!: () => void;
  const waiting = new Promise<void>((resolve) => {
    release = resolve;
  });
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async (batch) => {
      batches.push(batch);
      if (batches.length === 1) await waiting;
    },
  });
  queue.add(user);
  const saving = queue.flush(false);
  await Promise.resolve();
  queue.add(assistant);
  release();
  await saving;
  assert.deepEqual(batches, [[wire(user)], [wire(user), wire(assistant)]]);
});

test("finalized IDs cannot silently change and retired sessions cannot save", async () => {
  let active = true;
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => {
      if (!active) throw new Error("Session changed");
    },
    persist: async () => undefined,
  });
  queue.add(user);
  queue.add(user);
  assert.throws(
    () => queue.add({ ...user, text: "Changed" }),
    /changed after it was finalized/,
  );
  active = false;
  assert.throws(() => queue.flush(), /Session changed/);
});

test("ending a greeting-only call does not write or attach it to the next call", async () => {
  const batches: VoiceHistoryMessage[][] = [];
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async (batch) => {
      batches.push(batch);
    },
  });
  queue.add(assistant);
  await queue.flush();
  assert.equal(queue.hasPending, false);
  assert.equal(batches.length, 0);
  const nextUser = {
    ...user,
    id: "voice-843a4bbd-b15b-491f-a03a-14d5f49c3cf6-user",
  };
  queue.add(nextUser);
  await queue.flush();
  assert.deepEqual(batches, [[wire(nextUser)]]);
});

test("history chunks keep a user's multiple assistant followups together", async () => {
  const batches: VoiceHistoryMessage[][] = [];
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async (batch) => {
      batches.push(batch);
    },
  });
  queue.add(user);
  for (let index = 1; index <= 38; index++) {
    queue.add({
      ...assistant,
      id: `${prefix}assistant-${index}`,
      order: index,
    });
  }
  const nextUser = { ...user, id: `${prefix}user-2`, order: 39 };
  const nextAssistant = {
    ...assistant,
    id: `${prefix}assistant-39`,
    order: 40,
  };
  queue.add(nextUser);
  queue.add(nextAssistant);
  await queue.flush();
  assert.deepEqual(
    batches.map((batch) => batch.length),
    [39, 2],
  );
  assert.deepEqual(batches[1], [wire(nextUser), wire(nextAssistant)]);
  assert.equal(queue.hasPending, false);
});

test("one oversized voice turn stays pending without saving an incomplete group", async () => {
  let writes = 0;
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async () => {
      writes++;
    },
  });
  queue.add(user);
  for (let index = 1; index <= 40; index++) {
    queue.add({
      ...assistant,
      id: `${prefix}assistant-${index}`,
      order: index,
    });
  }
  await assert.rejects(queue.flush(), /exceeds the conversation history limit/);
  assert.equal(writes, 0);
  assert.equal(queue.hasPending, true);
});

test("concurrent flush callers do not automatically retry an uncertain save", async () => {
  let writes = 0;
  const queue = createVoiceHistoryQueue({
    assertCurrent: () => undefined,
    persist: async () => {
      writes++;
      throw new Error("Connection lost");
    },
  });
  queue.add(user);
  const results = await Promise.allSettled([queue.flush(), queue.flush()]);
  assert.equal(
    results.every((result) => result.status === "rejected"),
    true,
  );
  assert.equal(writes, 1);
  assert.equal(queue.hasPending, true);
});
