import { QueryClient } from "@tanstack/react-query";
import type { Comment } from "@/types";
import { storyKeys } from "@/modules/stories/constants";
import { reconcileDesktopOutbox } from "./outbox-projection";
import type { DesktopOperationAck } from "./outbox-types";

const scope = {
  accountId: "alice",
  workspaceId: "workspace",
  workspaceSlug: "acme",
};
const createdAt = "2026-10-17T10:00:00.000Z";
const comment = (id: string, parentId: string | null = null): Comment => ({
  id,
  parentId,
  storyId: "story",
  userId: "alice",
  user: {
    id: "alice",
    fullName: "Alice",
    username: "alice",
    avatarUrl: null,
    isActive: true,
    isSystem: false,
  },
  comment: id,
  createdAt,
  updatedAt: createdAt,
  subComments: [],
});
const ack = (
  resource: Comment,
  provisionalId = resource.id,
): DesktopOperationAck => ({
  operationId: `create-${resource.id}`,
  kind: "comment.create",
  provisionalId,
  resourceId: resource.id,
  resource: { ...resource },
});
const reconcile = (client: QueryClient, ...completed: DesktopOperationAck[]) => {
  reconcileDesktopOutbox(client, scope, { operations: [], completed });
};

it("retains published replies when a durable creation receipt replays over main and linked thread caches", () => {
  const client = new QueryClient();
  const root = comment("root");
  const reply = comment("maya-reply", root.id);
  const fetched = { ...root, subComments: [reply] };
  const mainKey = storyKeys.commentsInfinite("acme", "story");
  const threadKey = storyKeys.commentThread("acme", "story", reply.id);
  const foreignKey = storyKeys.commentThread("elsewhere", "story", reply.id);
  const pages = {
    pageParams: [1, 2],
    pages: [{ comments: [] }, { comments: [fetched] }],
  };
  client.setQueryData(mainKey, pages);
  client.setQueryData(threadKey, fetched);
  client.setQueryData(foreignKey, fetched);

  reconcile(client, ack(root));
  reconcile(client, ack(root));

  expect(client.getQueryData(mainKey)).toMatchObject(pages);
  expect(client.getQueryData(threadKey)).toMatchObject(fetched);
  expect(client.getQueryData(foreignKey)).toEqual(fetched);
  expect(client.getQueryData(threadKey)).toMatchObject({
    desktopPending: false,
    desktopCreated: false,
  });
});

it("preserves nested replies and newer comment edits and profile data", () => {
  const client = new QueryClient();
  const root = comment("root");
  const reply = comment("reply", root.id);
  const nested = comment("maya-reply", reply.id);
  const current = {
    ...root,
    subComments: [
      {
        ...reply,
        comment: "An edited follow-up",
        updatedAt: "2026-10-17T11:00:00.000Z",
        user: { ...reply.user, fullName: "Alice Renamed" },
        subComments: [nested],
      },
    ],
  };
  const key = storyKeys.commentThread("acme", "story", nested.id);
  client.setQueryData(key, current);

  reconcile(client, ack(root), ack(reply));

  expect(client.getQueryData(key)).toMatchObject(current);
});

it("does not roll back refreshed author data when the comment's own revision is unchanged", () => {
  const client = new QueryClient();
  const root = comment("root");
  const current = {
    ...root,
    user: {
      ...root.user,
      fullName: "Alice Renamed",
      avatarUrl: "/new-avatar.png",
    },
  };
  const key = storyKeys.commentThread("acme", "story", root.id);
  client.setQueryData(key, current);

  reconcile(client, ack(root));

  expect(client.getQueryData(key)).toMatchObject(current);
});

it.each(["canonical-first", "provisional-first"])(
  "keeps the canonical thread topology during equal-revision identity reconciliation (%s)",
  (order) => {
    const client = new QueryClient();
    const provisional = comment("provisional");
    const canonical = {
      ...comment("root"),
      user: { ...provisional.user, fullName: "Alice Renamed" },
      subComments: [comment("maya-reply", "root")],
    };
    const key = storyKeys.commentsInfinite("acme", "story");
    client.setQueryData(key, {
      pages: [
        {
          comments:
            order === "canonical-first"
              ? [canonical, provisional]
              : [provisional, canonical],
        },
      ],
    });

    reconcile(client, ack(comment("root"), "provisional"));

    expect(client.getQueryData(key)).toMatchObject({
      pages: [{ comments: [canonical] }],
    });
  },
);

it("does not resurrect deleted replies from an old acknowledgement or provisional duplicate", () => {
  const client = new QueryClient();
  const root = comment("root");
  const oldRoot = { ...root, subComments: [comment("deleted-reply", root.id)] };
  const key = storyKeys.commentsInfinite("acme", "story");
  client.setQueryData(key, {
    pages: [{ comments: [root, { ...oldRoot, id: "provisional" }] }],
  });

  reconcile(client, ack(oldRoot, "provisional"));
  reconcile(client, ack(oldRoot, "provisional"));

  expect(client.getQueryData(key)).toMatchObject({
    pages: [{ comments: [root] }],
  });
});
