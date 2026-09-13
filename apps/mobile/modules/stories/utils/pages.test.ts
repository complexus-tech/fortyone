import assert from "node:assert/strict";
import test from "node:test";
import type { Story } from "../types";
import { mergeStoryPages } from "./pages";

test("overlapping pages render each task once without mutating cached pages", () => {
  const current = { id: "a", title: "Current first page" } as Story;
  const second = { id: "b", title: "Next task" } as Story;
  const firstPage = Object.freeze([current]);
  const nextPage = Object.freeze([
    { ...current, title: "Older overlap" },
    second,
  ]);
  const result = mergeStoryPages([firstPage, nextPage]);
  assert.deepEqual(result, [current, second]);
  assert.equal(result[0], current);
  assert.equal(firstPage.length, 1);
  assert.equal(nextPage.length, 2);
});
