import type { QueryClient } from "@tanstack/react-query";
import type { Comment } from "@/types";
import type { DetailedStory, NewStory } from "@/modules/story/types";
import { storyKeys } from "@/modules/stories/constants";
import { userKeys } from "@/constants/keys";
import type { DesktopCacheScope } from "./cache-types";
import type { DesktopOutboxSnapshot } from "./outbox-types";
import { localStory } from "./story-offline";
import { acknowledgeAttachment, projectAttachment } from "./outbox-attachments";
import { findCachedStory, projectStory } from "./outbox-story-projection";
import { patchResource, remapResource, setChanged } from "./outbox-records";

const createdTimes = new Map<string, string>();
export function retainDesktopCreatedStory(
  client: QueryClient,
  scope: DesktopCacheScope,
  story: DetailedStory,
) {
  setChanged(client, storyKeys.detail(scope.workspaceSlug, story.id), story);
  projectStory(client, scope, story);
}

export function retainDesktopComment(
  client: QueryClient,
  scope: DesktopCacheScope,
  comment: Comment,
) {
  const key = storyKeys.commentsInfinite(scope.workspaceSlug, comment.storyId);
  const original = client.getQueryData<{
    pages: { comments: Comment[]; pagination: unknown }[];
    pageParams: unknown[];
  }>(key);
  if (!original) {
    if (!comment.parentId)
      setChanged(client, key, {
        pageParams: [1],
        pages: [
          {
            comments: [comment],
            pagination: { page: 1, pageSize: 20, hasMore: false, nextPage: 2 },
          },
        ],
      });
    return;
  }
  const withoutPlaceholder = (comments: Comment[]): Comment[] =>
    comments
      .filter(
        (item) =>
          item.id !== "new comment" ||
          item.comment !== comment.comment ||
          item.parentId !== comment.parentId,
      )
      .map((item) => ({
        ...item,
        subComments: withoutPlaceholder(item.subComments),
      }));
  const contains = (comments: Comment[]): boolean =>
    comments.some(
      (item) => item.id === comment.id || contains(item.subComments),
    );
  if (original.pages.some((page) => contains(page.comments))) return;
  const insertReply = (comments: Comment[]): Comment[] =>
    comments.map((item) => {
      if (item.id === comment.parentId) {
        return { ...item, subComments: [...item.subComments, comment] };
      }
      return { ...item, subComments: insertReply(item.subComments) };
    });
  const pages = original.pages.map((page, index) => {
    let comments = withoutPlaceholder(page.comments);
    if (comment.parentId) comments = insertReply(comments);
    else if (index === 0) {
      comments = [comment, ...comments];
    }
    return { ...page, comments };
  });
  if (pages.some((page) => contains(page.comments)))
    setChanged(client, key, { ...original, pages });
}

/** Reapply durable local work after hydration, background refetches, and SSE invalidations. */
export function reconcileDesktopOutbox(
  client: QueryClient,
  scope: DesktopCacheScope,
  snapshot: DesktopOutboxSnapshot,
) {
  const queries = () =>
    client.getQueriesData({ queryKey: storyKeys.all(scope.workspaceSlug) });
  for (const ack of snapshot.completed) {
    if (ack.kind === "attachment.upload") {
      acknowledgeAttachment(client, scope, ack);
      continue;
    }
    const fields: Record<string, unknown> = {
      ...ack.resource,
      id: ack.resourceId,
      desktopPending: false,
      desktopCreated: false,
    };
    if (ack.kind === "comment.create") {
      // A creation receipt owns the comment, not the current thread or author
      // profile. Replies and profile changes do not advance its updatedAt.
      delete fields.subComments;
      delete fields.user;
    }
    for (const [key, data] of queries()) {
      const mapped = remapResource(data, ack.provisionalId, ack.resourceId, {
        preferCanonical: ack.kind === "comment.create",
      });
      const updated = patchResource(mapped, ack.resourceId, fields, true);
      setChanged(client, key, updated);
      if (
        ack.kind === "story.create" &&
        key[2] === "detail" &&
        key[3] === ack.provisionalId &&
        ack.provisionalId !== ack.resourceId
      ) {
        const destinationKey = [
          ...key.slice(0, 3),
          ack.resourceId,
          ...key.slice(4),
        ];
        const current = client.getQueryData(destinationKey);
        if (!current) setChanged(client, destinationKey, updated);
        else
          setChanged(
            client,
            destinationKey,
            patchResource(current, ack.resourceId, fields, true),
          );
      }
    }
    if (ack.kind.startsWith("story.")) {
      const current = findCachedStory(client, scope, ack.resourceId);
      if (current) projectStory(client, scope, current);
    }
  }
  for (const operation of snapshot.operations) {
    const timeKey = `${scope.accountId}:${scope.workspaceId}:${operation.id}`;
    let createdAt = createdTimes.get(timeKey);
    if (!createdAt) {
      createdAt = new Date().toISOString();
      createdTimes.set(timeKey, createdAt);
    }
    if (operation.kind === "story.create") {
      const old = findCachedStory(client, scope, operation.resourceId);
      retainDesktopCreatedStory(
        client,
        scope,
        localStory(
          { ...operation.payload, ...old } as NewStory,
          operation.resourceId,
          scope,
          old?.createdAt ?? createdAt,
        ),
      );
    } else if (operation.kind === "story.update") {
      const current = findCachedStory(client, scope, operation.resourceId);
      if (current) {
        const story = {
          ...current,
          ...operation.payload,
          id: operation.resourceId,
          desktopPending: true,
        };
        setChanged(
          client,
          storyKeys.detail(scope.workspaceSlug, story.id),
          story,
        );
        projectStory(client, scope, story);
      }
    } else if (operation.kind === "attachment.upload") {
      projectAttachment(client, scope, operation);
    } else if (
      typeof operation.payload.storyId === "string" &&
      typeof operation.payload.comment === "string"
    ) {
      const profile = client.getQueryData<{
        fullName?: string;
        username?: string;
        avatarUrl?: string | null;
      }>(userKeys.profile());
      retainDesktopComment(client, scope, {
        id: operation.resourceId,
        storyId: operation.payload.storyId,
        parentId:
          typeof operation.payload.parentId === "string"
            ? operation.payload.parentId
            : null,
        userId: scope.accountId,
        user: {
          id: scope.accountId,
          fullName: profile?.fullName ?? "",
          username: profile?.username ?? "",
          avatarUrl: profile?.avatarUrl ?? null,
          isActive: true,
          isSystem: false,
        },
        comment: operation.payload.comment,
        createdAt,
        updatedAt: createdAt,
        subComments: [],
      });
    }
  }
}
