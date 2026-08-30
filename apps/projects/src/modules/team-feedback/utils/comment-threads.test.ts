/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { TeamFeedbackComment } from "../types";
import { getCommentThreads } from "./comment-threads";

const createComment = (
  id: string,
  createdAt: string,
  parentId?: string,
): TeamFeedbackComment => ({
  authorId: "author-id",
  authorName: "Ari",
  body: id,
  createdAt,
  id,
  itemId: "feedback-id",
  parentId,
  updatedAt: createdAt,
  workspaceId: "workspace-id",
});

describe("getCommentThreads", () => {
  it("sorts top-level comments newest first and replies oldest first", () => {
    const comments = [
      createComment("first", "2026-08-01T12:00:00.000Z"),
      createComment("reply-newer", "2026-08-03T12:00:00.000Z", "first"),
      createComment("second", "2026-08-02T12:00:00.000Z"),
      createComment("reply-older", "2026-08-02T12:00:00.000Z", "first"),
    ];

    expect(getCommentThreads(comments)).toEqual([
      {
        comment: comments[2],
        replies: [],
      },
      {
        comment: comments[0],
        replies: [comments[3], comments[1]],
      },
    ]);
    expect(comments.map(({ id }) => id)).toEqual([
      "first",
      "reply-newer",
      "second",
      "reply-older",
    ]);
  });
});
