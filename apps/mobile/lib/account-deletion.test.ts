import assert from "node:assert/strict";
import test from "node:test";
import {
  AccountDeletionCleanupError,
  completeDeletedAccount,
  createAccountDeletionOperation,
  deleteConfirmedAccount,
  readAccountDeletionResult,
} from "./account-deletion";
import { clearLocalAccountData } from "./session-cleanup";

const original = {
  cookie: "fortyone_session=original",
  userId: "account-a",
  apiOrigin: "https://api.example.invalid",
  workspace: "workspace-a",
  expiresAt: Date.now() + 60_000,
};
const scope = { userId: original.userId, sessionEpoch: 1 };
const response = (status: number, body?: unknown) => ({
  status,
  json: async () => body,
});

test("only the documented 204 and 202 responses confirm account deletion", async () => {
  assert.equal(await readAccountDeletionResult(response(204)), "deleted");
  assert.equal(
    await readAccountDeletionResult(
      response(202, { data: { status: "cleanup_pending" } }),
    ),
    "cleanup_pending",
  );
  for (const res of [
    response(200),
    response(202, {}),
    response(202, { data: { status: "queued" } }),
  ]) {
    await assert.rejects(readAccountDeletionResult(res), /not confirmed/);
  }
});

test("failed or sole-admin-conflict requests retain the local session", async () => {
  for (const status of [401, 409, 500]) {
    let cleared = false;
    const failure = Object.assign(
      new Error("Assign another administrator in Design first."),
      { status },
    );
    await assert.rejects(
      deleteConfirmedAccount(scope, {
        isCurrent: () => true,
        getSession: async () => original,
        send: async () => {
          throw failure;
        },
        complete: async () => {
          cleared = true;
          return true;
        },
      }),
      (error) => error === failure,
    );
    assert.equal(cleared, false);
  }
});

test("a confirmation cannot delete the next account or workspace session", async () => {
  let current = true;
  let sent = false;
  await assert.rejects(
    deleteConfirmedAccount(scope, {
      isCurrent: () => current,
      getSession: async () => {
        current = false;
        return { ...original, userId: "account-b" };
      },
      send: async () => {
        sent = true;
        return response(204);
      },
      complete: async () => true,
    }),
    /session changed/,
  );
  assert.equal(sent, false);
});

test("accepted deletion uses the captured cookie and awaits local cleanup", async () => {
  const events: string[] = [];
  const result = await deleteConfirmedAccount(scope, {
    isCurrent: () => true,
    getSession: async () => original,
    send: async (session) => {
      assert.equal(session.cookie, original.cookie);
      events.push("accepted");
      return response(202, { data: { status: "cleanup_pending" } });
    },
    complete: async (cookie, result) => {
      assert.equal(cookie, original.cookie);
      assert.equal(result, "cleanup_pending");
      await Promise.resolve();
      events.push("cleaned");
      return true;
    },
  });
  events.push("finished");
  assert.deepEqual(events, ["accepted", "cleaned", "finished"]);
  assert.deepEqual(result, { result: "cleanup_pending", signedOut: true });
});

test("a late accepted response cannot clear another signed-in session", async () => {
  let session = original;
  let cleared = false;
  const result = await deleteConfirmedAccount(scope, {
    isCurrent: () => true,
    getSession: async () => session,
    send: async () => {
      session = {
        ...original,
        cookie: "fortyone_session=new",
        userId: "account-b",
      };
      return response(204);
    },
    complete: (cookie, result) =>
      completeDeletedAccount(cookie, result, {
        getVersion: () => 2,
        getSession: async () => session,
        clearLocalSession: async () => {
          cleared = true;
        },
        reportCleanupFailure: () =>
          assert.fail("must not clear replacement session"),
      }),
  });
  assert.equal(result.signedOut, false);
  assert.equal(cleared, false);
});

test("a session transition during credential verification prevents deletion cleanup", async () => {
  let version = 1;
  const completed = await completeDeletedAccount(original.cookie, "deleted", {
    getVersion: () => version,
    getSession: async () => {
      version++;
      return original;
    },
    clearLocalSession: async () =>
      assert.fail("must not clear during replacement"),
    reportCleanupFailure: () => assert.fail("no cleanup attempted"),
  });
  assert.equal(completed, false);
});

test("local cleanup failure does not misreport accepted account deletion", async () => {
  let message = "";
  const completed = await completeDeletedAccount(
    original.cookie,
    "cleanup_pending",
    {
      getVersion: () => 1,
      getSession: async () => original,
      clearLocalSession: async () => {
        throw new Error("Storage unavailable");
      },
      reportCleanupFailure: (value) => {
        message = value;
      },
    },
  );
  assert.equal(completed, true);
  assert.match(message, /Your account has been deleted/);
  assert.match(message, /connected-service data is continuing/);
  assert.match(message, /could not be removed from this device/);
});

test("accepted deletion survives credential-read failure and retries only local cleanup", async () => {
  for (const res of [
    response(204),
    response(202, { data: { status: "cleanup_pending" } }),
  ]) {
    let sends = 0;
    let clears = 0;
    let readable = false;
    const operation = createAccountDeletionOperation({
      isCurrent: () => true,
      getSession: async () => original,
      send: async () => {
        sends++;
        return res;
      },
      complete: (cookie, result) =>
        completeDeletedAccount(cookie, result, {
          getVersion: () => 1,
          getSession: async () => {
            if (!readable) throw new Error("SecureStore unavailable");
            return original;
          },
          clearLocalSession: async () => {
            clears++;
          },
          reportCleanupFailure: () => assert.fail("cleanup succeeds"),
        }),
    });

    await assert.rejects(operation(scope), (error) => {
      assert.ok(error instanceof AccountDeletionCleanupError);
      assert.match(error.message, /account has been/);
      assert.match(error.message, /Finish signing out/);
      assert.doesNotMatch(error.message, /fortyone_session/);
      return true;
    });
    assert.equal(clears, 0, "unverified credentials must not be erased");
    readable = true;
    assert.equal((await operation(scope)).signedOut, true);
    assert.equal(clears, 1);
    assert.equal(sends, 1, "cleanup retry must not repeat DELETE");
  }
});

test("an accepted deletion checkpoint cannot be retried for another session", async () => {
  let sends = 0;
  let completions = 0;
  const operation = createAccountDeletionOperation({
    isCurrent: () => true,
    getSession: async () => original,
    send: async () => {
      sends++;
      return response(204);
    },
    complete: async () => {
      completions++;
      throw new Error("SecureStore unavailable");
    },
  });
  await assert.rejects(operation(scope), AccountDeletionCleanupError);
  await assert.rejects(
    operation({ userId: "account-b", sessionEpoch: 2 }),
    /session changed/,
  );
  await assert.rejects(
    operation({ ...scope, sessionEpoch: 2 }),
    /session changed/,
  );
  assert.equal(sends, 1);
  assert.equal(completions, 1);
});

test("local sign-out cannot expose a new sign-in until pending draft cleanup finishes", async () => {
  let release = () => {};
  const pendingDrafts = new Promise<void>((resolve) => {
    release = resolve;
  });
  const events: string[] = [];
  const cleanup = clearLocalAccountData({
    resetCache: async () => {
      events.push("cache");
    },
    clearCredentials: async () => {
      events.push("credential");
    },
    clearDrafts: async () => {
      events.push("drafts");
      await pendingDrafts;
    },
    finish: () => {
      events.push("signed-out");
    },
  });
  await new Promise<void>((resolve) => setImmediate(resolve));
  assert.deepEqual(events, ["cache", "credential", "drafts"]);
  release();
  await cleanup;
  assert.equal(events.at(-1), "signed-out");
});

test("cleanup attempts credentials and drafts even when cache cleanup fails", async () => {
  const events: string[] = [];
  await assert.rejects(
    clearLocalAccountData({
      resetCache: async () => {
        throw new Error("Cache failed");
      },
      clearCredentials: async () => {
        events.push("credential");
      },
      clearDrafts: async () => {
        events.push("drafts");
      },
      finish: () => {
        events.push("signed-out");
      },
    }),
    /Cache failed/,
  );
  assert.deepEqual(events, ["credential", "drafts", "signed-out"]);
});
