import assert from "node:assert/strict";
import test from "node:test";
import type { RichTextValue } from "../../components/rich-text/content";
import type { FormState } from "./form-state";
import { createDraftRepository } from "../../components/rich-text/draft-storage";
import { initialState } from "./form-state";
import {
  createStoryFinalizer,
  CreatedStoryDraftCleanupError,
} from "./submission";

type Options = Parameters<ReturnType<typeof createStoryFinalizer>["submit"]>[0];

const description: RichTextValue = {
  html: "<p>The last typed sentence.</p>",
  text: "The last typed sentence.",
  mentions: ["member-1"],
};
const draft = (overrides: Partial<FormState> = {}): FormState => ({
  ...initialState,
  title: "Create a task",
  teamId: "team-1",
  idempotencyKey: "durable-retry-key",
  ...overrides,
});
const options = (overrides: Partial<Options> = {}): Options => ({
  getDraft: () => draft(),
  description,
  availableTeamIds: ["team-1", "team-2"],
  createIdempotencyKey: () => "generated-retry-key",
  persistDraft: async () => {},
  createStory: async () => "created-task",
  clearDraft: async () => {},
  assertActive: () => {},
  ...overrides,
});
const deferred = () => {
  let resolve: () => void = () => {};
  const promise = new Promise<void>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};

test("submission reads current native fields and the acknowledged DOM description without a React render", async () => {
  const finalizer = createStoryFinalizer(() => {});
  let latest = draft({ title: "Old rendered title" });
  let submitted: FormState | undefined;
  let persisted: FormState | undefined;
  const request = options({
    getDraft: () => latest,
    persistDraft: async (snapshot) => {
      persisted = snapshot;
    },
    createStory: async (snapshot) => {
      assert.equal(snapshot, persisted);
      submitted = snapshot;
      return "created-task";
    },
  });
  latest = draft({
    title: "  Latest native title  ",
    teamId: "team-2",
    labelIds: ["latest-label"],
  });
  await finalizer.submit(request);
  assert.equal(submitted?.title, "Latest native title");
  assert.equal(submitted?.teamId, "team-2");
  assert.deepEqual(submitted?.labelIds, ["latest-label"]);
  assert.deepEqual(submitted?.description, description);
  assert.notEqual(submitted?.description.mentions, description.mentions);
  assert.equal(submitted?.idempotencyKey, "durable-retry-key");
});

test("the submission lock covers durable persistence, POST and cleanup, coalescing repeated presses", async () => {
  const persistStarted = deferred();
  const persistDone = deferred();
  const cleanupStarted = deferred();
  const cleanupDone = deferred();
  let creates = 0;
  let lateWrites = 0;
  const finalizer = createStoryFinalizer(() => {});
  const request = options({
    persistDraft: async () => {
      persistStarted.resolve();
      await persistDone.promise;
    },
    createStory: async () => {
      creates += 1;
      return "created-task";
    },
    clearDraft: async () => {
      cleanupStarted.resolve();
      await cleanupDone.promise;
    },
  });
  const pending = finalizer.submit(request);
  assert.equal(finalizer.submit(request), pending);
  assert.equal(finalizer.canEdit(), false);
  await persistStarted.promise;
  assert.equal(creates, 0);
  await assert.rejects(
    finalizer.persistEdit(async () => {
      lateWrites += 1;
    }),
    /can no longer be changed/,
  );
  persistDone.resolve();
  await cleanupStarted.promise;
  assert.equal(creates, 1);
  assert.equal(finalizer.canEdit(), false);
  await assert.rejects(
    finalizer.persistEdit(async () => {
      lateWrites += 1;
    }),
  );
  cleanupDone.resolve();
  assert.equal(await pending, "created-task");
  await assert.rejects(
    finalizer.persistEdit(async () => {
      lateWrites += 1;
    }),
  );
  assert.equal(lateWrites, 0);
  assert.equal(finalizer.canEdit(), false);
});

test("cleanup failure retains the server result and retry does not create another task", async () => {
  const states: { isSubmitting: boolean; isFinalized: boolean }[] = [];
  const finalizer = createStoryFinalizer((state) => states.push(state));
  let creates = 0;
  let clears = 0;
  const request = options({
    createStory: async () => {
      creates += 1;
      return "already-created";
    },
    clearDraft: async () => {
      clears += 1;
      if (clears === 1) throw new Error("Storage unavailable");
    },
  });
  await assert.rejects(
    finalizer.submit(request),
    CreatedStoryDraftCleanupError,
  );
  assert.deepEqual(states.at(-1), { isSubmitting: false, isFinalized: true });
  assert.equal(finalizer.canEdit(), false);
  assert.equal(
    await finalizer.submit({
      ...request,
      getDraft: () => {
        throw new Error("Completed content must not be resubmitted");
      },
      persistDraft: async () => {
        throw new Error("Completed draft must not be rewritten");
      },
    }),
    "already-created",
  );
  assert.equal(creates, 1);
  assert.equal(clears, 2);
  assert.equal(await finalizer.submit(request), "already-created");
  assert.equal(creates, 1);
  assert.equal(clears, 2);
});

test("a failed durable write prevents POST and cleanup, and leaves editing available", async () => {
  const finalizer = createStoryFinalizer(() => {});
  let creates = 0;
  let clears = 0;
  await assert.rejects(
    finalizer.submit(
      options({
        persistDraft: async () => {
          throw new Error("Disk full");
        },
        createStory: async () => {
          creates += 1;
          return "task";
        },
        clearDraft: async () => {
          clears += 1;
        },
      }),
    ),
    /Disk full/,
  );
  assert.equal(creates, 0);
  assert.equal(clears, 0);
  assert.equal(finalizer.canEdit(), true);
  let edited = false;
  await finalizer.persistEdit(async () => {
    edited = true;
  });
  assert.equal(edited, true);
});

test("a retry after a failed create reuses the idempotency key persisted before the first request", async () => {
  const finalizer = createStoryFinalizer(() => {});
  let latest = draft({ idempotencyKey: "" });
  const keys: string[] = [];
  const request = options({
    getDraft: () => latest,
    persistDraft: async (snapshot) => {
      latest = snapshot;
    },
    createStory: async (snapshot) => {
      assert.equal(latest.idempotencyKey, "generated-retry-key");
      keys.push(snapshot.idempotencyKey);
      if (keys.length === 1) throw new Error("Connection interrupted");
      return "idempotent-task";
    },
  });
  await assert.rejects(finalizer.submit(request), /Connection interrupted/);
  assert.equal(await finalizer.submit(request), "idempotent-task");
  assert.deepEqual(keys, ["generated-retry-key", "generated-retry-key"]);
});

test("invalid title or unavailable team is rejected before writing or creating", async () => {
  for (const value of [
    draft({ title: "  " }),
    draft({ teamId: "removed-team" }),
  ]) {
    const finalizer = createStoryFinalizer(() => {});
    let sideEffects = 0;
    await assert.rejects(
      finalizer.submit(
        options({
          getDraft: () => value,
          persistDraft: async () => {
            sideEffects += 1;
          },
          createStory: async () => {
            sideEffects += 1;
            return "unexpected";
          },
        }),
      ),
    );
    assert.equal(sideEffects, 0);
    assert.equal(finalizer.canEdit(), true);
  }
});

test("a session change during creation prevents clearing or navigating an account's newer draft", async () => {
  const finalizer = createStoryFinalizer(() => {});
  let active = true;
  let clears = 0;
  const request = options({
    assertActive: () => {
      if (!active) throw new Error("Session changed");
    },
    createStory: async () => {
      active = false;
      return "old-session-task";
    },
    clearDraft: async () => {
      clears += 1;
    },
  });
  await assert.rejects(finalizer.submit(request), /Session changed/);
  await assert.rejects(finalizer.submit(request), /Session changed/);
  assert.equal(clears, 0);
  assert.equal(finalizer.canEdit(), false);
});

test("queued autosaves finish before final persistence and late bridge callbacks cannot resurrect a cleared draft", async () => {
  const store = new Map<string, string>();
  const releaseOldWrite = deferred();
  let oldWrite = true;
  const repository = createDraftRepository({
    getItem: async (key) => store.get(key) ?? null,
    setItem: async (key, value) => {
      if (oldWrite) {
        oldWrite = false;
        await releaseOldWrite.promise;
      }
      store.set(key, value);
    },
    removeItem: async (key) => {
      store.delete(key);
    },
  });
  const finalizer = createStoryFinalizer(() => {});
  const autosave = finalizer.persistEdit(() =>
    repository.write("draft", draft({ title: "Older autosave" })),
  );
  const pending = finalizer.submit(
    options({
      persistDraft: (snapshot) => repository.write("draft", snapshot),
      createStory: async (snapshot) => {
        assert.deepEqual(JSON.parse(store.get("draft") ?? "null"), snapshot);
        return "created-task";
      },
      clearDraft: () => repository.remove("draft"),
    }),
  );
  releaseOldWrite.resolve();
  await autosave;
  await pending;
  await assert.rejects(
    finalizer.persistEdit(() =>
      repository.write("draft", draft({ title: "Late bridge callback" })),
    ),
  );
  assert.equal(store.has("draft"), false);
});
