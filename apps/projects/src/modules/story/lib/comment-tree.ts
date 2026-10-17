import type { Comment } from "@/types/comment";

export const MAX_COMMENT_DEPTH = 3;

type CommentEntry = {
  comment: Comment;
  rootId: string;
  containerId: string | null;
};

const collectComments = (threads: readonly Comment[]) => {
  const entries = new Map<string, CommentEntry>();
  for (const comment of threads) {
    if (!entries.has(comment.id)) {
      entries.set(comment.id, {
        comment,
        rootId: comment.id,
        containerId: null,
      });
    }
  }
  const pending: CommentEntry[] = threads
    .map((comment) => ({ comment, rootId: comment.id, containerId: null }))
    .reverse();
  const visited = new Set<Comment>();

  while (pending.length) {
    const entry = pending.pop()!;
    if (visited.has(entry.comment)) continue;
    visited.add(entry.comment);
    if (!entries.has(entry.comment.id)) entries.set(entry.comment.id, entry);
    // Keep the first payload, but collect children from every duplicate copy.
    const rootId = entries.get(entry.comment.id)!.rootId;
    for (
      let index = entry.comment.subComments.length - 1;
      index >= 0;
      index--
    ) {
      pending.push({
        comment: entry.comment.subComments[index],
        rootId,
        containerId: entry.comment.id,
      });
    }
  }
  return entries;
};

const compareComments = (left: Comment, right: Comment) => {
  const chronological =
    Date.parse(left.createdAt) - Date.parse(right.createdAt);
  return (
    (Number.isFinite(chronological) ? chronological : 0) ||
    left.id.localeCompare(right.id)
  );
};

/** Rebuild mixed API/optimistic replies for display without changing parentId. */
export const normalizeCommentThreads = (
  threads: readonly Comment[],
): Comment[] => {
  const entries = collectComments(threads);
  const roots = Array.from(entries.values()).filter(
    ({ comment, rootId }) => comment.id === rootId,
  );
  const copies = new Map(
    Array.from(entries, ([id, { comment }]) => [
      id,
      { ...comment, subComments: [] as Comment[] },
    ]),
  );
  const children = new Map<string, string[]>();
  for (const [id, { comment, rootId, containerId }] of entries) {
    if (id === rootId) continue;
    const parentId = comment.parentId ?? containerId;
    const parent = parentId ? entries.get(parentId) : undefined;
    const targetId =
      parent && parent.rootId === rootId && parentId !== id
        ? parent.comment.id
        : rootId;
    const siblings = children.get(targetId) ?? [];
    siblings.push(id);
    children.set(targetId, siblings);
  }

  const displayed = new Set<string>();
  for (const { rootId } of roots) {
    const pending = [{ id: rootId, branchId: rootId }];
    displayed.add(rootId);
    for (const { id, branchId } of pending) {
      for (const childId of children.get(id) ?? []) {
        if (displayed.has(childId)) continue;
        displayed.add(childId);
        // Descendants share their level-two ancestor once visual depth reaches three.
        copies.get(branchId)!.subComments.push(copies.get(childId)!);
        pending.push({
          id: childId,
          branchId: id === rootId ? childId : branchId,
        });
      }
    }
  }

  // Cycles cannot be reached from a root. Keep their comments visible once.
  for (const [id, { rootId }] of entries) {
    if (!displayed.has(id))
      copies.get(rootId)!.subComments.push(copies.get(id)!);
  }
  for (const comment of copies.values())
    comment.subComments.sort(compareComments);
  return roots.map(({ rootId }) => copies.get(rootId)!);
};
