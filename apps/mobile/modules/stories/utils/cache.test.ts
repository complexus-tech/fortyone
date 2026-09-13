import assert from "node:assert/strict";
import test from "node:test";
import { QueryClient } from "@tanstack/react-query";
import {
  optimisticallyUpdateStory,
  restoreStoryCache,
  updateStoryCache,
} from "./cache.ts";

const task = { id: "one", title: "Before", labels: ["old"] };
const other = { id: "two", title: "Unchanged" };

test("updates grouped, paginated, detail and nested story data without losing metadata", () => {
  const grouped = {
    groups: [{ key: "todo", stories: [task, other], totalCount: 2 }],
    meta: { totalGroups: 1 },
  };
  const infinite = {
    pages: [{ stories: [task, other], pagination: { hasMore: true } }],
    pageParams: [1],
  };
  assert.deepEqual(updateStoryCache(grouped, "one", { title: "After" }), {
    ...grouped,
    groups: [
      { ...grouped.groups[0], stories: [{ ...task, title: "After" }, other] },
    ],
  });
  assert.deepEqual(updateStoryCache(infinite, "one", { labels: ["new"] }), {
    ...infinite,
    pages: [
      { ...infinite.pages[0], stories: [{ ...task, labels: ["new"] }, other] },
    ],
  });
  assert.deepEqual(
    updateStoryCache({ ...other, subStories: [task] }, "one", {
      title: "After",
    }),
    {
      ...other,
      subStories: [{ ...task, title: "After" }],
    },
  );
  assert.equal(
    updateStoryCache(grouped, "missing", { title: "After" }),
    grouped,
  );
});

test("failed mutations restore every touched cache but preserve a newer update", async () => {
  const client = new QueryClient();
  const prefix = ["session", "a", "workspace", "stories"];
  const detail = [...prefix, "detail", "one"];
  const list = [...prefix, "list"];
  client.setQueryData(detail, task);
  client.setQueryData(list, { groups: [{ stories: [task, other] }] });
  const snapshots = await optimisticallyUpdateStory(client, prefix, "one", {
    title: "Optimistic",
  });
  assert.equal(snapshots.length, 2);
  client.setQueryData(detail, { ...task, title: "Newer change" });
  restoreStoryCache(client, snapshots);
  assert.deepEqual(client.getQueryData(detail), {
    ...task,
    title: "Newer change",
  });
  assert.deepEqual(client.getQueryData(list), {
    groups: [{ stories: [task, other] }],
  });
  client.clear();
});
