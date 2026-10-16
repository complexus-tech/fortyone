/* global describe, expect, it -- Jest globals are provided by the projects test runner. */
import type { Comment } from "@/types/comment";
import {
  findCommentThread,
  isUnreadComment,
  latestComment,
} from "./comment-navigation";

const comment = (
  id: string,
  createdAt: string,
  overrides: Partial<Comment> = {},
): Comment => ({
  id,
  createdAt,
  updatedAt: createdAt,
  comment: "<p>Comment</p>",
  parentId: null,
  storyId: "story",
  userId: "author",
  subComments: [],
  user: {
    id: "author",
    username: "author",
    fullName: "Author",
    avatarUrl: null,
    isSystem: false,
    isActive: true,
  },
  ...overrides,
});

describe("comment navigation", () => {
  const older = "2026-10-15T12:00:00Z";
  const newer = "2026-10-16T12:00:00Z";
  it("finds the containing root and latest reply even in an older thread", () => {
    const reply = comment("reply", newer, { parentId: "nested" });
    const root = comment("root", older, {
      subComments: [comment("nested", older, { subComments: [reply] })],
    });
    const roots = [comment("recent-root", "2026-10-16T11:00:00Z"), root];
    expect(findCommentThread(roots, "reply")).toBe(root);
    expect(latestComment(roots)).toBe(reply);
    expect(findCommentThread(roots, "deleted")).toBeUndefined();
  });
  it("only marks persisted comments from other authors strictly after the read marker", () => {
    expect(isUnreadComment(comment("reply", newer), older, "me")).toBe(true);
    expect(isUnreadComment(comment("reply", older), older, "me")).toBe(false);
    expect(isUnreadComment(comment("reply", newer), undefined, "me")).toBe(
      false,
    );
    expect(isUnreadComment(comment("reply", newer), null, "author")).toBe(
      false,
    );
    expect(
      isUnreadComment(comment("pending-comment:1", newer), null, "me"),
    ).toBe(false);
    expect(isUnreadComment(comment("reply", newer), null, "me")).toBe(true);
  });
  it("does not jump to optimistic comments which may not be saved yet", () => {
    const saved = comment("saved", older);
    expect(latestComment([comment("pending-comment:1", newer), saved])).toBe(
      saved,
    );
    expect(latestComment([])).toBeUndefined();
  });
});
