import assert from "node:assert/strict";
import test from "node:test";
import type { SearchResponse } from "./types";
import { getNextSearchPage, mergeSearchPages } from "./pagination";

const page = (overrides: Partial<SearchResponse> = {}): SearchResponse => ({
  stories: [],
  objectives: [],
  totalStories: 0,
  totalObjectives: 0,
  page: 1,
  pageSize: 20,
  totalPages: 0,
  ...overrides,
});

test("search reaches later pages and stops on a final or unexpectedly empty page", () => {
  const result = page({
    page: 1,
    totalPages: 3,
    stories: [{ id: "first" } as SearchResponse["stories"][number]],
  });
  assert.equal(getNextSearchPage(result), 2);
  assert.equal(getNextSearchPage({ ...result, page: 3 }), undefined);
  assert.equal(getNextSearchPage({ ...result, stories: [] }), undefined);
});

test("changing result positions across pages never duplicates a rendered row", () => {
  const story = {
    id: "a",
    title: "Original",
  } as SearchResponse["stories"][number];
  const merged = mergeSearchPages([
    page({ stories: [story], totalStories: 21, totalPages: 2 }),
    page({
      page: 2,
      stories: [
        { ...story, title: "Updated" },
        { ...story, id: "b" },
      ],
    }),
  ]);
  assert.deepEqual(
    merged?.stories.map(({ id, title }) => ({ id, title })),
    [
      { id: "a", title: "Updated" },
      { id: "b", title: "Original" },
    ],
  );
  assert.equal(merged?.totalStories, 21);
  assert.equal(mergeSearchPages(undefined), undefined);
});
