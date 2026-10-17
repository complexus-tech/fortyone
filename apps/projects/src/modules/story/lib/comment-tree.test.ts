/* global describe, expect, it -- Jest globals are provided by the projects test runner. */
import type { Comment } from "@/types/comment";
import { flattenComments } from "./comment-navigation";
import { normalizeCommentThreads } from "./comment-tree";

const comment = (id: string, overrides: Partial<Comment> = {}): Comment => ({
  id,
  storyId: "story",
  parentId: null,
  userId: "author",
  comment: `<p>${id}</p>`,
  createdAt: "2026-10-16T12:00:00Z",
  updatedAt: "2026-10-16T12:00:00Z",
  subComments: [],
  user: {
    id: "author",
    username: "author",
    fullName: "Author",
    avatarUrl: null,
    isActive: true,
    isSystem: false,
  },
  ...overrides,
});

describe("normalizeCommentThreads", () => {
  it("renders flat server replies and nested optimistic replies identically", () => {
    const reply = comment("reply", { parentId: "root" });
    const nested = comment("nested", { parentId: "reply" });
    const server = comment("root", { subComments: [nested, reply] });
    const optimistic = comment("root", {
      subComments: [{ ...reply, subComments: [nested] }],
    });

    expect(normalizeCommentThreads([server])).toEqual(
      normalizeCommentThreads([optimistic]),
    );
    expect(normalizeCommentThreads([server])[0].subComments).toEqual([
      { ...reply, subComments: [nested] },
    ]);
  });

  it("preserves root order and normalizes a single linked thread independently", () => {
    const root = comment("older", {
      subComments: [
        comment("child", { parentId: "parent" }),
        comment("parent", { parentId: "older" }),
      ],
    });
    const recent = comment("recent", { createdAt: "2026-10-17T12:00:00Z" });
    const threads = normalizeCommentThreads([recent, root]);

    expect(threads.map(({ id }) => id)).toEqual(["recent", "older"]);
    expect(threads[1]).toEqual(normalizeCommentThreads([root])[0]);
  });

  it("uses actual parents over containers and legacy containers only when absent", () => {
    const explicit = comment("explicit", { parentId: "second" });
    const legacy = comment("legacy");
    const missing = comment("missing");
    Reflect.deleteProperty(missing, "parentId");
    const roots = normalizeCommentThreads([
      comment("root", {
        subComments: [
          comment("first", { subComments: [explicit, legacy, missing] }),
          comment("second", { parentId: "root" }),
        ],
      }),
    ]);

    expect(
      roots[0].subComments.map(({ id, subComments }) => [
        id,
        subComments.map((child) => child.id),
      ]),
    ).toEqual([
      ["first", ["legacy", "missing"]],
      ["second", ["explicit"]],
    ]);
    expect(roots[0].subComments[0].subComments[0].parentId).toBeNull();
    expect(roots[0].subComments[1].subComments[0].parentId).toBe("second");
  });

  it("keeps every historical descendant at depth three without rewriting parents", () => {
    const descendants = Array.from({ length: 30 }, (_, index) =>
      comment(`reply-${String(index).padStart(2, "0")}`, {
        parentId: index
          ? `reply-${String(index - 1).padStart(2, "0")}`
          : "root",
      }),
    );
    const normalized = normalizeCommentThreads([
      comment("root", { subComments: descendants.slice().reverse() }),
    ])[0];

    expect(normalized.subComments).toHaveLength(1);
    expect(normalized.subComments[0].subComments).toHaveLength(29);
    expect(
      normalized.subComments[0].subComments.every(
        (child) => child.subComments.length === 0,
      ),
    ).toBe(true);
    expect(flattenComments([normalized]).map(({ id }) => id)).toEqual([
      "root",
      ...descendants.map(({ id }) => id),
    ]);
    expect(normalized.subComments[0].subComments[28].parentId).toBe("reply-28");
  });

  it("falls back under the known root for invalid, self, cross-root and cyclic parents", () => {
    const normalized = normalizeCommentThreads([
      comment("root", {
        subComments: [
          comment("orphan", { parentId: "missing" }),
          comment("self", { parentId: "self" }),
          comment("cross-root", { parentId: "other" }),
          comment("cycle-a", { parentId: "cycle-b" }),
          comment("cycle-b", { parentId: "cycle-a" }),
          comment("cycle-child", { parentId: "cycle-a" }),
        ],
      }),
      comment("other"),
    ]);

    expect(normalized[0].subComments.map(({ id }) => id)).toEqual([
      "cross-root",
      "cycle-a",
      "cycle-b",
      "cycle-child",
      "orphan",
      "self",
    ]);
    expect(normalized[1].subComments).toEqual([]);
    expect(
      normalized[0].subComments.every(
        (child) => child.subComments.length === 0,
      ),
    ).toBe(true);
  });

  it("deduplicates mixed copies while collecting all their descendants", () => {
    const original = comment("reply", { parentId: "root" });
    const root = comment("root", {
      subComments: [
        original,
        {
          ...original,
          comment: "Duplicate",
          subComments: [comment("child", { parentId: "reply" })],
        },
      ],
    });
    const duplicateRoot = {
      ...root,
      subComments: [comment("another", { parentId: "root" })],
    };
    const normalized = normalizeCommentThreads([root, duplicateRoot]);

    expect(normalized).toHaveLength(1);
    expect(flattenComments(normalized).map(({ id }) => id)).toEqual([
      "root",
      "another",
      "reply",
      "child",
    ]);
    expect(normalized[0].subComments[1].comment).toBe(original.comment);
  });

  it("orders siblings chronologically with a deterministic ID tie break", () => {
    const normalized = normalizeCommentThreads([
      comment("root", {
        subComments: [
          comment("later", { createdAt: "2026-10-17T12:00:00Z" }),
          comment("b"),
          comment("a"),
        ],
      }),
    ]);
    expect(normalized[0].subComments.map(({ id }) => id)).toEqual([
      "a",
      "b",
      "later",
    ]);
  });

  it("does not mutate input comments and remains stable when normalized again", () => {
    const root = comment("root", {
      subComments: [comment("reply", { parentId: "root" })],
    });
    Object.freeze(root.subComments[0]);
    Object.freeze(root.subComments);
    Object.freeze(root);
    const normalized = normalizeCommentThreads([root]);

    expect(normalized[0]).not.toBe(root);
    expect(normalized[0].subComments[0]).not.toBe(root.subComments[0]);
    expect(root.subComments).toHaveLength(1);
    expect(normalizeCommentThreads(normalized)).toEqual(normalized);
    expect(normalizeCommentThreads([])).toEqual([]);
  });

  it("handles accidental cyclic containers without dropping other comments", () => {
    const root = comment("root");
    const reply = comment("reply", { parentId: "root", subComments: [root] });
    root.subComments = [reply];

    expect(
      flattenComments(normalizeCommentThreads([root])).map(({ id }) => id),
    ).toEqual(["root", "reply"]);
  });
});
