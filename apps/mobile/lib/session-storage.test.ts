import assert from "node:assert/strict";
import { test } from "node:test";
import { createSessionStorage, SESSION_KEY } from "./session-storage.ts";

const session = {
  apiOrigin: "https://api.fortyone.app",
  cookie: `fortyone_session=${"s".repeat(43)}`,
  userId: "user-a",
  workspace: "team-a",
  expiresAt: 5000,
};
const fakeStorage = (initial: Record<string, string> = {}) => {
  const entries = new Map(Object.entries(initial));
  return {
    entries,
    getItem: async (key: string) => entries.get(key) ?? null,
    setItem: async (key: string, value: string) => {
      entries.set(key, value);
    },
    removeItem: async (key: string) => {
      entries.delete(key);
    },
  };
};

test("restores persisted native session and does not trust legacy login flags", async () => {
  const adapter = fakeStorage({ hasSession: "true", workspace: "old-team" });
  const storage = createSessionStorage(adapter, () => 1000);
  assert.equal(await storage.get(), null);
  await storage.save(session);
  assert.deepEqual(
    await createSessionStorage(adapter, () => 1000).get(),
    session,
  );
});

test("serializes logout after pending persistence so a late write cannot restore credentials", async () => {
  const adapter = fakeStorage();
  let release: () => void = () => {};
  const barrier = new Promise<void>((resolve) => {
    release = resolve;
  });
  const storage = createSessionStorage(
    {
      ...adapter,
      setItem: async (key, value) => {
        await barrier;
        await adapter.setItem(key, value);
      },
    },
    () => 1000,
  );
  const writing = storage.save(session);
  const clearing = storage.clear();
  release();
  await Promise.all([writing, clearing]);
  assert.equal(await storage.get(), null);
  assert.equal(adapter.entries.has(SESSION_KEY), false);
});

test("expires locally and erases obsolete auth hints", async () => {
  const adapter = fakeStorage({
    [SESSION_KEY]: JSON.stringify(session),
    hasSession: "true",
    workspace: "old-team",
  });
  assert.equal(await createSessionStorage(adapter, () => 5000).get(), null);
  assert.equal(adapter.entries.size, 0);
});

test("rechecks a queued response after logout deletes its credential", async () => {
  const adapter = fakeStorage({ [SESSION_KEY]: JSON.stringify(session) });
  let current = true;
  let release: () => void = () => {};
  const barrier = new Promise<void>((resolve) => {
    release = resolve;
  });
  const storage = createSessionStorage(
    {
      ...adapter,
      removeItem: async (key) => {
        await barrier;
        await adapter.removeItem(key);
      },
    },
    () => 1000,
  );
  const clearing = storage.clear();
  // The response was accepted before logout finished, but reaches the storage
  // queue behind credential deletion and must not recreate the session.
  const writing = storage.save(session, () => current);
  current = false;
  release();
  await clearing;
  await assert.rejects(writing, /session changed/);
  assert.equal(adapter.entries.has(SESSION_KEY), false);
  assert.equal(await storage.get(), null);
});

test("erases a native write invalidated while the disk operation was in flight", async () => {
  const adapter = fakeStorage();
  let current = true;
  let release: () => void = () => {};
  let writingStarted: () => void = () => {};
  const started = new Promise<void>((resolve) => {
    writingStarted = resolve;
  });
  const barrier = new Promise<void>((resolve) => {
    release = resolve;
  });
  const storage = createSessionStorage(
    {
      ...adapter,
      setItem: async (key, value) => {
        writingStarted();
        await barrier;
        await adapter.setItem(key, value);
      },
    },
    () => 1000,
  );
  const writing = storage.save(session, () => current);
  await started;
  current = false;
  release();
  await assert.rejects(writing, /session changed/);
  assert.equal(adapter.entries.has(SESSION_KEY), false);
  assert.equal(await storage.get(), null);
});

test("rejects damaged session data without treating a flag as authentication", async () => {
  for (const raw of [
    "invalid-json",
    "null",
    JSON.stringify({ ...session, cookie: "malformed" }),
    JSON.stringify({ ...session, userId: null }),
  ]) {
    assert.equal(
      await createSessionStorage(
        fakeStorage({ [SESSION_KEY]: raw }),
        () => 1000,
      ).get(),
      null,
    );
  }
});

test("a superseded workspace save preserves the existing credential until its replacement runs", async () => {
  const adapter = fakeStorage({ [SESSION_KEY]: JSON.stringify(session) });
  let current = true;
  const storage = createSessionStorage(
    {
      ...adapter,
      setItem: async (key, value) => {
        await adapter.setItem(key, value);
        current = false;
      },
    },
    () => 1000,
  );
  await storage.get();
  await assert.rejects(
    storage.save({ ...session, workspace: "team-b" }, () => current),
    /session changed/,
  );
  assert.deepEqual(await storage.get(), session);
  assert.deepEqual(JSON.parse(adapter.entries.get(SESSION_KEY)!), session);
});
