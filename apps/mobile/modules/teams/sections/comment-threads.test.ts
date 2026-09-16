import assert from "node:assert/strict";
import test from "node:test";
import { threadedFeedbackComments } from "./comment-threads";

test("feedback replies stay with their parent in chronological order without changing the cache", () => {
  const comments = [
    { id: "first", createdAt: "2026-09-01" },
    { id: "reply-new", createdAt: "2026-09-04", parentId: "first" },
    { id: "second", createdAt: "2026-09-02" },
    { id: "reply-old", createdAt: "2026-09-03", parentId: "first" },
  ].map((comment) => ({ ...comment, authorName: "Rudo", body: comment.id }));
  assert.deepEqual(
    threadedFeedbackComments(comments).map((comment) => comment.id),
    ["second", "first", "reply-old", "reply-new"],
  );
  assert.deepEqual(
    comments.map((comment) => comment.id),
    ["first", "reply-new", "second", "reply-old"],
  );
});

test("a reply remains visible if its parent is unavailable", () => {
  const comment = {
    id: "reply",
    parentId: "unavailable",
    authorName: "Rudo",
    body: "Update",
    createdAt: "2026-09-01",
  };
  assert.deepEqual(threadedFeedbackComments([comment]), [comment]);
});
