import { QueryClient } from "@tanstack/react-query";
import type { Comment } from "@/types";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
  StoryFilters,
} from "@/modules/stories/types";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "@/modules/stories/constants";
import { getStoryPath } from "@/shared/routing/story";
import { localStory } from "./story-offline";
import { discardDesktopProjection } from "./outbox-discard";
import {
  reconcileDesktopOutbox,
  retainDesktopComment,
  retainDesktopCreatedStory,
} from "./outbox-projection";
import type { DesktopOperation, DesktopOperationAck } from "./outbox-types";

const scope = {
  accountId: "alice",
  workspaceId: "workspace",
  workspaceSlug: "acme",
};
const date = "2026-10-15T10:00:00.000Z";
const story = (id: string, overrides: Partial<DetailedStory> = {}) => ({
  ...localStory(
    { title: id, statusId: "todo", teamId: "team-one", ...overrides },
    id,
    scope,
    date,
  ),
  desktopPending: false,
  desktopCreated: false,
});
const grouped = (
  stories: DetailedStory[],
  filters: StoryFilters = {},
): GroupedStoriesResponse => ({
  groups: [
    {
      key: "todo",
      stories,
      totalCount: stories.length,
      loadedCount: stories.length,
      hasMore: false,
      nextPage: 1,
    },
  ],
  meta: {
    groupBy: "status",
    orderBy: "created",
    orderDirection: "desc",
    totalGroups: 1,
    filters,
  },
});
const page = (
  stories: DetailedStory[],
  groupKey = "todo",
): GroupStoriesResponse => ({
  groupKey,
  stories,
  pagination: { page: 1, pageSize: 20, hasMore: false, nextPage: 2 },
  filters: {},
  orderBy: "created",
  orderDirection: "desc",
});
const operation = (
  patch: Partial<DesktopOperation> = {},
): DesktopOperation => ({
  id: "operation",
  kind: "story.update",
  resourceId: "item",
  payload: { statusId: "done" },
  baseVersion: date,
  status: "pending",
  attempted: false,
  ...patch,
});
const reconcile = (
  client: QueryClient,
  operations: DesktopOperation[] = [],
  completed: DesktopOperationAck[] = [],
) => {
  reconcileDesktopOutbox(client, scope, { operations, completed });
};
const comment = (id: string, parentId: string | null = null): Comment => ({
  id,
  parentId,
  storyId: "item",
  comment: id,
  userId: "alice",
  user: {
    id: "alice",
    fullName: "Alice",
    username: "alice",
    avatarUrl: null,
    isActive: true,
    isSystem: false,
  },
  createdAt: date,
  updatedAt: date,
  subComments: [],
});

it("forces provisional identity and uses the UUID route even if a form carries an old reference", () => {
  const pending = localStory(
    { id: "wrong", sequenceId: 12, teamCode: "ENG", workspaceId: "other" },
    "pending-id",
    scope,
    date,
  );
  expect(pending).toMatchObject({
    id: "pending-id",
    workspaceId: "workspace",
    desktopCreated: true,
    teamCode: "",
    sequenceId: 0,
  });
  expect(getStoryPath({ ...pending, teamCode: "ENG" })).toBe(
    "/work/pending-id",
  );
});

it("treats My Work ownership as OR and respects team and workspace view scopes", () => {
  const client = new QueryClient();
  const mine = storyKeys.mineGrouped("acme", {
    assignedToMe: true,
    createdByMe: true,
    collaboratingWithMe: true,
  });
  const otherTeam = storyKeys.teamGrouped("acme", "team-two", {});
  client.setQueryData(mine, grouped([]));
  client.setQueryData(otherTeam, grouped([]));
  client.setQueryData(storyKeys.mineGrouped("elsewhere", {}), grouped([]));
  retainDesktopCreatedStory(
    client,
    scope,
    story("created", { reporterId: "alice", assigneeId: "bob" }),
  );
  expect(
    client
      .getQueryData<GroupedStoriesResponse>(mine)
      ?.groups[0].stories.map((item) => item.id),
  ).toEqual(["created"]);
  expect(
    client.getQueryData<GroupedStoriesResponse>(otherTeam)?.groups[0].stories,
  ).toEqual([]);
  expect(
    client.getQueryData<GroupedStoriesResponse>(
      storyKeys.mineGrouped("elsewhere", {}),
    )?.groups[0].stories,
  ).toEqual([]);
});

it("moves grouped rows and counts once, patches nested copies, and stays stable on replay", () => {
  const client = new QueryClient();
  const child = story("item");
  const parent = story("parent", { subStories: [child] });
  const key = storyKeys.grouped("acme");
  client.setQueryData(key, grouped([child, parent]));
  reconcile(client, [operation()]);
  const first = client.getQueryData<GroupedStoriesResponse>(key)!;
  expect(
    first.groups.map(({ key: id, totalCount, loadedCount }) => [
      id,
      totalCount,
      loadedCount,
    ]),
  ).toEqual([
    ["todo", 1, 1],
    ["done", 1, 1],
  ]);
  expect(first.groups[0].stories[0].subStories[0].statusId).toBe("done");
  expect(first.meta.totalGroups).toBe(2);
  reconcile(client, [operation()]);
  expect(client.getQueryData(key)).toEqual(first);
});

it("moves an infinite row from later pages into an uncached destination without duplicating on replay", () => {
  const client = new QueryClient();
  const params = { groupBy: "status", groupKey: "todo", teamIds: ["team-one"] };
  const source = storyKeys.groupStories("acme", "todo", params);
  const destination = storyKeys.groupStories("acme", "done", {
    ...params,
    groupKey: "done",
  });
  client.setQueryData(source, {
    pageParams: [1, 2],
    pages: [page([]), page([story("item")])],
  });
  reconcile(client, [operation()]);
  reconcile(client, [operation()]);
  expect(
    client
      .getQueryData<{ pages: GroupStoriesResponse[] }>(source)
      ?.pages.flatMap((item) => item.stories),
  ).toHaveLength(0);
  expect(
    client
      .getQueryData<{ pages: GroupStoriesResponse[] }>(destination)
      ?.pages.flatMap((item) => item.stories)
      .map((item) => item.id),
  ).toEqual(["item"]);
});

it("removes updates from filtered views they no longer match", () => {
  const client = new QueryClient();
  const key = storyKeys.grouped("acme");
  client.setQueryData(key, grouped([story("item")], { statusIds: ["todo"] }));
  reconcile(client, [operation()]);
  expect(
    client.getQueryData<GroupedStoriesResponse>(key)?.groups[0],
  ).toMatchObject({ stories: [], totalCount: 0, loadedCount: 0 });
});

it("preserves newer server fields when old ACKs replay and clears pending-create markers", () => {
  const client = new QueryClient();
  const key = storyKeys.detail("acme", "item");
  client.setQueryData(
    key,
    story("item", {
      title: "Newer server edit",
      updatedAt: "2026-10-15T12:00:00.000Z",
    }),
  );
  reconcile(
    client,
    [],
    [
      {
        operationId: "old",
        kind: "story.update",
        provisionalId: "item",
        resourceId: "item",
        resource: { title: "Old ACK", updatedAt: date },
      },
    ],
  );
  expect(client.getQueryData(key)).toMatchObject({
    title: "Newer server edit",
    desktopPending: false,
    desktopCreated: false,
  });
});

it("deduplicates a provisional row already returned by a server refetch and repairs counts", () => {
  const client = new QueryClient();
  const key = storyKeys.grouped("acme");
  client.setQueryData(
    key,
    grouped([
      story("provisional"),
      story("server", {
        updatedAt: "2026-10-15T12:00:00.000Z",
        title: "Latest",
      }),
    ]),
  );
  reconcile(
    client,
    [],
    [
      {
        operationId: "create",
        kind: "story.create",
        provisionalId: "provisional",
        resourceId: "server",
        resource: { id: "server", updatedAt: date },
      },
    ],
  );
  const result = client.getQueryData<GroupedStoriesResponse>(key)!;
  expect(result.groups[0]).toMatchObject({ totalCount: 1, loadedCount: 1 });
  expect(result.groups[0].stories.map((item) => [item.id, item.title])).toEqual(
    [["server", "Latest"]],
  );
});

it("remaps resource and cached thread identities without replacing user-authored text", () => {
  const client = new QueryClient();
  const original = localStory(
    {
      title: "Mention provisional",
      description: "provisional",
      statusId: "todo",
    },
    "provisional",
    scope,
    date,
  );
  client.setQueryData(storyKeys.detail("acme", "provisional"), original);
  client.setQueryData(storyKeys.commentsInfinite("acme", "provisional"), {
    pageParams: [1],
    pages: [
      {
        comments: [
          {
            ...comment("reply"),
            storyId: "provisional",
            comment: "provisional",
          },
        ],
      },
    ],
  });
  reconcile(
    client,
    [],
    [
      {
        operationId: "create",
        kind: "story.create",
        provisionalId: "provisional",
        resourceId: "server",
        resource: {
          id: "server",
          sequenceId: 13,
          teamCode: "ENG",
          updatedAt: date,
        },
      },
    ],
  );
  expect(client.getQueryData(storyKeys.detail("acme", "server"))).toMatchObject(
    {
      id: "server",
      title: "Mention provisional",
      description: "provisional",
      desktopCreated: false,
    },
  );
  expect(
    client.getQueryData<{ pages: { comments: Comment[] }[] }>(
      storyKeys.commentsInfinite("acme", "server"),
    )?.pages[0].comments[0],
  ).toMatchObject({ storyId: "server", comment: "provisional" });
});

it("inserts replies under parents on later pages and preserves unrelated in-flight placeholders", () => {
  const client = new QueryClient();
  const key = storyKeys.commentsInfinite("acme", "item");
  const unrelated = { ...comment("new comment"), comment: "Other submission" };
  client.setQueryData(key, {
    pageParams: [1, 2],
    pages: [{ comments: [unrelated] }, { comments: [comment("parent")] }],
  });
  const reply = comment("reply", "parent");
  retainDesktopComment(client, scope, reply);
  retainDesktopComment(client, scope, reply);
  const data = client.getQueryData<{ pages: { comments: Comment[] }[] }>(key)!;
  expect(data.pages[0].comments[0].comment).toBe("Other submission");
  expect(data.pages[1].comments[0].subComments.map((item) => item.id)).toEqual([
    "reply",
  ]);
});

it("removes discarded offline creates and comments without waiting for a server refetch", () => {
  const client = new QueryClient();
  const key = storyKeys.grouped("acme");
  client.setQueryData(key, grouped([story("item")]));
  client.setQueryData(storyKeys.detail("acme", "item"), story("item"));
  discardDesktopProjection(client, scope, operation({ kind: "story.create" }));
  expect(
    client.getQueryData<GroupedStoriesResponse>(key)?.groups[0],
  ).toMatchObject({ stories: [], loadedCount: 0, totalCount: 0 });
  expect(client.getQueryData(storyKeys.detail("acme", "item"))).toBeUndefined();
  const commentsKey = storyKeys.commentsInfinite("acme", "other-story");
  client.setQueryData(commentsKey, {
    pages: [
      {
        comments: [
          { ...comment("parent"), subComments: [comment("reply", "parent")] },
        ],
      },
    ],
  });
  discardDesktopProjection(
    client,
    scope,
    operation({ kind: "comment.create", resourceId: "reply" }),
  );
  expect(
    client.getQueryData<{ pages: { comments: Comment[] }[] }>(commentsKey)
      ?.pages[0].comments[0].subComments,
  ).toEqual([]);
});

it("restores conflict server state and refuses an offline update discard without a before-image", () => {
  const client = new QueryClient();
  const key = storyKeys.grouped("acme");
  client.setQueryData(key, grouped([story("item")]));
  reconcile(client, [operation()]);
  expect(() => {
    discardDesktopProjection(client, scope, operation());
  }).toThrow("server version");
  discardDesktopProjection(
    client,
    scope,
    operation({
      status: "conflict",
      serverValue: { statusId: "todo", title: "Server title", updatedAt: date },
    }),
  );
  const data = client.getQueryData<GroupedStoriesResponse>(key)!;
  expect(data.groups[0].stories[0]).toMatchObject({
    title: "Server title",
    statusId: "todo",
    desktopPending: false,
  });
  expect(data.groups[1].stories).toEqual([]);
});

it("stages, acknowledges and discards attachment rows without remapping or removing the owning story", () => {
  const client = new QueryClient();
  const storyKey = storyKeys.detail("acme", "item");
  const attachmentsKey = storyKeys.attachments("acme", "item");
  client.setQueryData(storyKey, story("item"));
  const upload = operation({
    kind: "attachment.upload",
    resourceId: "item",
    payload: {
      attachmentId: "local-file",
      fileName: "brief.pdf",
      contentType: "application/pdf",
      size: 120,
      createdAt: date,
    },
  });
  reconcile(client, [upload]);
  expect(client.getQueryData(attachmentsKey)).toEqual([
    expect.objectContaining({
      id: "local-file",
      pending: true,
      filename: "brief.pdf",
    }),
  ]);
  discardDesktopProjection(client, scope, upload);
  expect(client.getQueryData(attachmentsKey)).toEqual([]);
  reconcile(client, [upload]);
  reconcile(
    client,
    [],
    [
      {
        operationId: "upload",
        kind: "attachment.upload",
        provisionalId: "local-file",
        resourceId: "server-file",
        storyId: "item",
        resource: {
          id: "server-file",
          filename: "brief.pdf",
          url: "https://files.example/brief.pdf",
        },
      },
    ],
  );
  expect(client.getQueryData(attachmentsKey)).toEqual([
    expect.objectContaining({ id: "server-file", pending: false }),
  ]);
  expect(client.getQueryData(storyKey)).toMatchObject({
    id: "item",
    title: "item",
  });
});
