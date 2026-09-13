import assert from "node:assert/strict";
import test from "node:test";
import { dehydrate, onlineManager } from "@tanstack/react-query";
import { createSessionClient } from "./session-query-client.ts";
import { scopedQueryKey, queryPersistenceKey } from "./query-scope.ts";

const memoryStorage = () => {
  const values = new Map<string, string>();
  return {
    values,
    getString: (key: string) => values.get(key),
    set: (key: string, value: string) => values.set(key, value),
    remove: (key: string) => values.delete(key),
    getAllKeys: () => [...values.keys()],
  };
};

test("persisted task data is isolated by account and workspace", async () => {
  const storage = memoryStorage();
  const accountA = { userId: "a", workspace: "first" };
  const accountB = { userId: "b", workspace: "first" };
  const otherWorkspace = { userId: "a", workspace: "second" };
  const first = createSessionClient(accountA, storage);
  first.client.setQueryData(scopedQueryKey(accountA, "stories"), [
    { id: "private-a" },
  ]);
  await first.persister.persistClient({
    timestamp: Date.now(),
    buster: "test",
    clientState: dehydrate(first.client),
  });

  const second = createSessionClient(accountB, storage);
  const third = createSessionClient(otherWorkspace, storage);
  assert.equal(await second.persister.restoreClient(), undefined);
  assert.equal(await third.persister.restoreClient(), undefined);
  assert.notDeepEqual(
    scopedQueryKey(accountA, "stories"),
    scopedQueryKey(accountB, "stories"),
  );
  assert.notDeepEqual(
    scopedQueryKey(accountA, "stories"),
    scopedQueryKey(otherWorkspace, "stories"),
  );
  await Promise.all([first.reset(), second.reset(), third.reset()]);
});

test("logout cancels requests and late persistence cannot recreate the old account snapshot", async () => {
  const storage = memoryStorage();
  const scope = { userId: "a", workspace: "first" };
  const session = createSessionClient(scope, storage);
  const snapshot = {
    timestamp: Date.now(),
    buster: "test",
    clientState: dehydrate(session.client),
  };
  await session.persister.persistClient(snapshot);
  let aborted = false;
  const request = session.client
    .fetchQuery({
      queryKey: scopedQueryKey(scope, "stories"),
      queryFn: ({ signal }) =>
        new Promise((_resolve, reject) => {
          signal.addEventListener("abort", () => {
            aborted = true;
            reject(new Error("aborted"));
          });
        }),
    })
    .catch(() => undefined);

  await session.reset();
  await session.persister.persistClient(snapshot);
  await request;
  assert.equal(aborted, true);
  assert.equal(storage.values.has(queryPersistenceKey(scope)), false);
  assert.equal(session.client.getQueryCache().getAll().length, 0);
  assert.equal(await session.persister.restoreClient(), undefined);
});

test("offline and disposed sessions reject writes instead of creating a paused mutation queue", async () => {
  const session = createSessionClient(
    { userId: "a", workspace: "first" },
    memoryStorage(),
  );
  let writes = 0;
  const execute = () =>
    session.client
      .getMutationCache()
      .build(session.client, {
        mutationFn: async () => {
          writes++;
        },
      })
      .execute(undefined);
  onlineManager.setOnline(false);
  try {
    await assert.rejects(execute, /offline/);
    assert.equal(writes, 0);
    assert.equal(
      session.client.getMutationCache().getAll()[0].state.isPaused,
      false,
    );
  } finally {
    onlineManager.setOnline(true);
  }
  await session.reset();
  await assert.rejects(execute, /session changed/);
  assert.equal(writes, 0);
  session.client.clear();
});
