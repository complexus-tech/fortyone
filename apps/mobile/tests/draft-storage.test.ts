import assert from "node:assert/strict";
import { test } from "node:test";
import {
  createDraftRepository,
  getDraftKey,
  type DraftStorage,
} from "../components/rich-text/draft-storage";

test("drafts are separated by account, workspace, and document without key collisions", () => {
  const keys = [
    getDraftKey("user-a", "workspace-a", "story:1"),
    getDraftKey("user-b", "workspace-a", "story:1"),
    getDraftKey("user-a", "workspace-b", "story:1"),
    getDraftKey("user-a", "workspace-a", "story:2"),
    getDraftKey("user:a", "b", "c"),
    getDraftKey("user", "a:b", "c"),
  ];
  assert.equal(new Set(keys).size, keys.length);
});

test("a delayed autosave cannot resurrect a draft after successful submission clears it", async () => {
  const values = new Map<string, string>();
  let release: () => void = () => {};
  const delay = new Promise<void>((resolve) => {
    release = resolve;
  });
  const storage: DraftStorage = {
    getItem: async (key) => values.get(key) ?? null,
    setItem: async (key, value) => {
      await delay;
      values.set(key, value);
    },
    removeItem: async (key) => {
      values.delete(key);
    },
  };
  const repository = createDraftRepository(storage);
  const writing = repository.write("draft", { text: "Last keystroke" });
  const removing = repository.remove("draft");
  release();
  await Promise.all([writing, removing]);
  assert.equal(values.has("draft"), false);
});

test("storage failures are surfaced and do not block a subsequent cleanup", async () => {
  let removed = false;
  const repository = createDraftRepository({
    getItem: async () => null,
    setItem: async () => {
      throw new Error("Disk full");
    },
    removeItem: async () => {
      removed = true;
    },
  });
  await assert.rejects(repository.write("draft", "text"), /Disk full/);
  await repository.remove("draft");
  assert.equal(removed, true);
});

test("invalid restored data is rejected instead of silently replacing a saved draft", async () => {
  const repository = createDraftRepository({
    getItem: async () => '{"html":42}',
    setItem: async () => {},
    removeItem: async () => {},
  });
  await assert.rejects(
    repository.read(
      "draft",
      (value): value is string => typeof value === "string",
    ),
    /cannot be opened/,
  );
});
