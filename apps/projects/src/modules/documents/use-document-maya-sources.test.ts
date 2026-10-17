import { act, renderHook } from "@testing-library/react";
import type { DocumentComment } from "./types";
import type { DocumentMayaRun } from "./document-comment-maya";
import { useDocumentMayaSources } from "./use-document-maya-sources";

const comment = (id: string, age = 0): DocumentComment => ({
  id,
  body: "[@Maya](agent:maya) help",
  createdBy: "alice",
  authorName: "Alice",
  authorAvatar: null,
  createdAt: new Date(Date.now() - age).toISOString(),
  updatedAt: new Date().toISOString(),
});
beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date("2026-10-16T12:00:00Z"));
});
afterEach(() => jest.useRealTimers());

test("the grace placeholder expires without relying on an unrelated render", () => {
  const comments = [
    comment("saved", 5_000),
    comment("optimistic-new"),
    comment("old", 16_000),
  ];
  const { result } = renderHook(() => useDocumentMayaSources(comments, []));
  expect(result.current.map((item) => item.id)).toEqual([
    "saved",
    "optimistic-new",
  ]);
  act(() => {
    jest.advanceTimersByTime(10_000);
  });
  expect(result.current.map((item) => item.id)).toEqual(["optimistic-new"]);
});

test("a real run immediately replaces its grace placeholder and clears its timer", () => {
  const comments = [comment("saved")];
  const { result, rerender, unmount } = renderHook(
    ({ runs }: { runs: DocumentMayaRun[] }) =>
      useDocumentMayaSources(comments, runs),
    { initialProps: { runs: [] as DocumentMayaRun[] } },
  );
  expect(result.current).toHaveLength(1);
  rerender({
    runs: [
      {
        id: "run",
        sourceId: "saved",
        requesterId: "alice",
        status: "running",
        summary: "",
        updatedAt: "2026-10-16T12:00:00Z",
        version: 1,
      },
    ],
  });
  expect(result.current).toEqual([]);
  expect(jest.getTimerCount()).toBe(0);
  unmount();
});
