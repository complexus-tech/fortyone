import assert from "node:assert/strict";
import test from "node:test";
import { createEditorFlushController } from "./flush-protocol";
import type { RichTextValue } from "./content";

const snapshot: RichTextValue = {
  html: "<p>Latest edit</p>",
  text: "Latest edit",
  mentions: [],
};

test("fire-and-forget DOM command cannot create until durable snapshot acknowledgement", async () => {
  const controller = createEditorFlushController();
  const events: string[] = [];
  let requestId = 0;
  let finishPersist!: () => void;
  const persisted = new Promise<void>((resolve) => {
    finishPersist = resolve;
  });
  const created = controller
    .request((id) => {
      requestId = id;
      events.push("command sent");
      void persisted.then(() => {
        events.push("draft persisted");
        controller.acknowledge(id, snapshot, null);
      });
    })
    .then((value) => {
      assert.deepEqual(value, snapshot);
      events.push("create");
    });
  await Promise.resolve();
  assert.deepEqual(events, ["command sent"]);
  controller.acknowledge(requestId + 1, snapshot, null);
  await Promise.resolve();
  assert.deepEqual(events, ["command sent"]);
  finishPersist();
  await created;
  assert.deepEqual(events, ["command sent", "draft persisted", "create"]);
});

test("failed persistence rejects flush and does not create or close", async () => {
  const controller = createEditorFlushController();
  let requestId = 0;
  let continued = false;
  const action = controller
    .request((id) => {
      requestId = id;
    })
    .then(() => {
      continued = true;
    });
  controller.acknowledge(requestId, null, "Secure storage is unavailable");
  await assert.rejects(action, /Secure storage/);
  assert.equal(continued, false);
});

test("concurrent requests share one command and stale acknowledgement cannot settle a retry", async () => {
  const controller = createEditorFlushController();
  const requests: number[] = [];
  const send = (id: number) => {
    requests.push(id);
  };
  const first = controller.request(send);
  assert.equal(controller.request(send), first);
  controller.acknowledge(requests[0], null, "Try again");
  await assert.rejects(first, /Try again/);
  let resolved = false;
  const retry = controller.request(send).then(() => {
    resolved = true;
  });
  controller.acknowledge(requests[0], snapshot, null);
  await Promise.resolve();
  assert.equal(resolved, false);
  controller.acknowledge(requests[1], snapshot, null);
  await retry;
  assert.equal(requests.length, 2);
});

test("unready, crashed, and disposed editor requests fail closed", async () => {
  const controller = createEditorFlushController();
  await assert.rejects(
    controller.request(() => {
      throw new Error("Still loading");
    }),
    /Still loading/,
  );
  const interrupted = controller.request(() => undefined);
  controller.fail(new Error("Editor crashed"));
  await assert.rejects(interrupted, /Editor crashed/);
  const pending = controller.request(() => undefined);
  controller.dispose();
  await assert.rejects(pending, /closed before saving/);
  await assert.rejects(
    controller.request(() => undefined),
    /closed/,
  );
});

test("missing acknowledgement times out and malformed snapshots are rejected", async () => {
  const controller = createEditorFlushController(10);
  await assert.rejects(
    controller.request(() => undefined),
    /did not finish saving/,
  );
  let requestId = 0;
  const invalid = controller.request((id) => {
    requestId = id;
  });
  controller.acknowledge(requestId, { html: "<p>missing text</p>" }, null);
  await assert.rejects(invalid, /invalid draft/);
});
