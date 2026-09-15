import assert from "node:assert/strict";
import test from "node:test";
import { createDraftRepository } from "../../../components/rich-text/draft-storage";
import { createCommentFinalizer } from "./comment-submission";

test("a known successful send freezes writes and retries only failed cleanup", async () => {
  let creates = 0;
  let clears = 0;
  let finalized = 0;
  let lateWrites = 0;
  const finalizer = createCommentFinalizer(() => finalized++);
  const create = async () => {
    creates++;
  };
  const clear = async () => {
    clears++;
    if (clears < 3) throw new Error("Disk unavailable");
  };

  await assert.rejects(finalizer.send(create, clear), /comment was sent/);
  assert.equal(finalized, 1);
  await assert.rejects(
    finalizer.persistDraft(async () => {
      lateWrites++;
    }),
    /no longer available/,
  );

  let dismissed = false;
  await assert.rejects(
    finalizer.close(clear).then(() => {
      dismissed = true;
    }),
    /comment was sent/,
  );
  assert.equal(dismissed, false);
  await finalizer.send(create, clear);
  await finalizer.close(clear);
  assert.equal(creates, 1);
  assert.equal(clears, 3);
  assert.equal(lateWrites, 0);
});

test("closing an unsent comment preserves its draft and fences late bridge writes", async () => {
  const finalizer = createCommentFinalizer(() => {});
  let writes = 0;
  let clears = 0;
  await finalizer.persistDraft(async () => {
    writes++;
  });
  await finalizer.close(async () => {
    clears++;
  });
  await assert.rejects(
    finalizer.persistDraft(async () => {
      writes++;
    }),
  );
  assert.equal(writes, 1);
  assert.equal(clears, 0);
});

test("failed discard preserves editing; successful discard rejects later writes", async () => {
  const finalizer = createCommentFinalizer(() => {});
  await assert.rejects(
    finalizer.discard(async () => {
      throw new Error("Cannot remove draft");
    }),
    /Cannot remove draft/,
  );
  await finalizer.persistDraft(async () => {});
  await finalizer.discard(async () => {});
  await assert.rejects(finalizer.persistDraft(async () => {}));
});

test("a failed create is not treated as a known success", async () => {
  let finalized = false;
  let clears = 0;
  const finalizer = createCommentFinalizer(() => {
    finalized = true;
  });
  await assert.rejects(
    finalizer.send(
      async () => {
        throw new Error("Request failed");
      },
      async () => {
        clears++;
      },
    ),
    /Request failed/,
  );
  await finalizer.persistDraft(async () => {});
  assert.equal(finalized, false);
  assert.equal(clears, 0);
});

test("concurrent Send and Close share finalization and queued writes cannot resurrect a cleared draft", async () => {
  let releaseWrite!: () => void;
  const delay = new Promise<void>((resolve) => {
    releaseWrite = resolve;
  });
  const stored = new Map<string, string>();
  const repository = createDraftRepository({
    getItem: async (key) => stored.get(key) ?? null,
    setItem: async (key, value) => {
      await delay;
      stored.set(key, value);
    },
    removeItem: async (key) => {
      stored.delete(key);
    },
  });
  const finalizer = createCommentFinalizer(() => {});
  const write = finalizer.persistDraft(() =>
    repository.write("comment", "draft"),
  );
  let creates = 0;
  const create = async () => {
    creates++;
  };
  const clear = () => repository.remove("comment");
  const sending = finalizer.send(create, clear);
  assert.equal(finalizer.send(create, clear), sending);
  assert.equal(finalizer.close(clear), sending);
  await assert.rejects(
    finalizer.persistDraft(() => repository.write("comment", "late draft")),
  );
  releaseWrite();
  await Promise.all([write, sending]);
  assert.equal(creates, 1);
  assert.equal(stored.has("comment"), false);
});
