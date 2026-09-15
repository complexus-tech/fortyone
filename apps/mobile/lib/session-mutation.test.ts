import assert from "node:assert/strict";
import test from "node:test";
import { MutationObserver, onlineManager } from "@tanstack/react-query";
import { createSessionClient } from "./session-query-client.ts";
import { withSessionMutationGuard } from "./session-mutation.ts";

function deferred() {
  let resolve = () => {};
  const promise = new Promise<void>((complete) => {
    resolve = complete;
  });
  return { promise, resolve };
}

function sessionClient(
  userId = "first-account",
  workspace = "first-workspace",
) {
  return createSessionClient(
    { userId, workspace },
    {
      getString: () => undefined,
      set: () => undefined,
      remove: () => undefined,
      getAllKeys: () => [],
    },
  );
}

test("reset during optimistic work cannot send a write in the next account or workspace", async () => {
  const first = sessionClient();
  const entered = deferred();
  const release = deferred();
  const writes: string[] = [];
  let currentAccount = "first-account";
  const callbacks: string[] = [];
  const observer = new MutationObserver(
    first.client,
    withSessionMutationGuard(first, {
      onMutate: async () => {
        entered.resolve();
        await release.promise;
        return "optimistic snapshot";
      },
      mutationFn: async () => {
        writes.push(currentAccount);
      },
      onError: () => {
        callbacks.push("old-session rollback");
      },
      onSettled: () => {
        callbacks.push("old-session refetch");
      },
    }),
  );
  const pending = observer.mutate(undefined);
  const rejection = assert.rejects(pending, /session changed/);
  await entered.promise;
  await first.reset();
  first.activate(); // A disposed provider cannot be revived by a late effect.
  currentAccount = "second-account";
  const second = sessionClient(currentAccount, "second-workspace");
  release.resolve();
  await rejection;
  assert.deepEqual(writes, []);
  assert.deepEqual(callbacks, []);
  await second.reset();
});

test("observer option updates retain the final admission guard", async () => {
  const session = sessionClient();
  const entered = deferred();
  const release = deferred();
  let writes = 0;
  const observer = new MutationObserver(
    session.client,
    withSessionMutationGuard(session, {
      onMutate: async () => {
        entered.resolve();
        await release.promise;
      },
      mutationFn: async () => {
        writes++;
      },
    }),
  );
  const rejection = assert.rejects(
    observer.mutate(undefined),
    /session changed/,
  );
  await entered.promise;
  observer.setOptions(
    withSessionMutationGuard(session, {
      mutationFn: async () => {
        writes++;
      },
    }),
  );
  await session.reset();
  release.resolve();
  await rejection;
  assert.equal(writes, 0);
});

test("going offline during optimistic work rejects before sending and still rolls back the active session", async () => {
  const session = sessionClient();
  const entered = deferred();
  const release = deferred();
  let writes = 0;
  let rolledBack = false;
  const observer = new MutationObserver(
    session.client,
    withSessionMutationGuard(session, {
      onMutate: async () => {
        entered.resolve();
        await release.promise;
        return "snapshot";
      },
      mutationFn: async () => {
        writes++;
      },
      onError: (_error, _variables, snapshot) => {
        rolledBack = snapshot === "snapshot";
      },
    }),
  );
  const rejection = assert.rejects(observer.mutate(undefined), /offline/);
  await entered.promise;
  onlineManager.setOnline(false);
  try {
    release.resolve();
    await rejection;
    assert.equal(writes, 0);
    assert.equal(rolledBack, true);
  } finally {
    onlineManager.setOnline(true);
    await session.reset();
  }
});

test("an admitted mutation preserves variables, result and optimistic context", async () => {
  const session = sessionClient();
  let received: unknown;
  const observer = new MutationObserver(
    session.client,
    withSessionMutationGuard(session, {
      mutationFn: async (value: number) => value * 2,
      onMutate: (value: number) => ({ previous: value - 1 }),
      onSuccess: (result, variables, snapshot, context) => {
        received = {
          result,
          variables,
          snapshot,
          sameClient: context.client === session.client,
        };
      },
    }),
  );
  assert.equal(await observer.mutate(4), 8);
  assert.deepEqual(received, {
    result: 8,
    variables: 4,
    snapshot: { previous: 3 },
    sameClient: true,
  });
  await session.reset();
});

test("an already-sent write may finish but cannot run callbacks in a disposed session", async () => {
  const session = sessionClient();
  const entered = deferred();
  const release = deferred();
  let callbacks = 0;
  const observer = new MutationObserver(
    session.client,
    withSessionMutationGuard(session, {
      mutationFn: async () => {
        entered.resolve();
        await release.promise;
        return "server-result";
      },
      onSuccess: () => {
        callbacks++;
      },
      onSettled: () => {
        callbacks++;
      },
    }),
  );
  const result = observer.mutate(undefined);
  await entered.promise;
  await session.reset();
  release.resolve();
  assert.equal(await result, "server-result");
  assert.equal(callbacks, 0);
});
