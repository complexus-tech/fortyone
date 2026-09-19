import assert from "node:assert/strict";
import test from "node:test";
import { QueryClient } from "@tanstack/react-query";
import { optimisticallyPatchEntity, restoreEntityCache } from "./cache.ts";

const objective = { id: "objective-one", name: "Before", health: "on_track" };
const other = { id: "objective-two", name: "Unchanged" };

test("patches matching entities across detail, list, and infinite cache shapes", async () => {
  const client = new QueryClient();
  const prefix = ["session", "user", "workspace", "objectives"];
  const detailKey = [...prefix, "detail", objective.id];
  const listKey = [...prefix, "list"];
  const infiniteKey = [...prefix, "activity"];
  client.setQueryData(detailKey, objective);
  client.setQueryData(listKey, { objectives: [objective, other], total: 2 });
  client.setQueryData(infiniteKey, {
    pages: [{ objectives: [objective], pagination: { hasMore: false } }],
  });

  const snapshots = await optimisticallyPatchEntity(
    client,
    [prefix],
    objective.id,
    { name: "After" },
  );

  assert.equal(snapshots.length, 3);
  assert.deepEqual(client.getQueryData(detailKey), {
    ...objective,
    name: "After",
  });
  assert.deepEqual(client.getQueryData(listKey), {
    objectives: [{ ...objective, name: "After" }, other],
    total: 2,
  });
  assert.equal(objective.name, "Before");
  client.clear();
});

test("rollback restores only snapshots that have not received a newer update", async () => {
  const client = new QueryClient();
  const prefix = ["session", "user", "workspace", "intake"];
  const detailKey = [...prefix, "detail", "request-one"];
  const listKey = [...prefix, "list"];
  const request = { id: "request-one", status: "pending" };
  client.setQueryData(detailKey, request);
  client.setQueryData(listKey, { requests: [request] });

  const snapshots = await optimisticallyPatchEntity(
    client,
    [prefix],
    request.id,
    { status: "accepted" },
  );
  client.setQueryData(detailKey, { ...request, status: "declined" });
  restoreEntityCache(client, snapshots);

  assert.deepEqual(client.getQueryData(detailKey), {
    ...request,
    status: "declined",
  });
  assert.deepEqual(client.getQueryData(listKey), { requests: [request] });
  client.clear();
});
