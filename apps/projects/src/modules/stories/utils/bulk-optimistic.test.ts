import type { InfiniteData, QueryKey } from "@tanstack/react-query";
import { QueryClient, QueryObserver } from "@tanstack/react-query";
import { statusKeys } from "@/constants/keys";
import { storyKeys } from "../constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
  GroupStoryParams,
  Story,
} from "../types";
import {
  applyBulkStoryUpdate,
  restoreBulkStoryUpdate,
} from "./bulk-optimistic";

const WORKSPACE = "workspace";
const GROUP_PARAMS = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
  teamIds: ["team"],
} satisfies Partial<GroupStoryParams>;
const story = (id: string, subStories: Story[] = []): Story =>
  ({ id, statusId: "started", subStories }) as Story;
const first = story("first");
const later = story("later");
const parent = story("parent", [story("child"), story("sibling")]);
const grouped = (): GroupedStoriesResponse => ({
  groups: [
    {
      key: "started",
      stories: [first, parent],
      loadedCount: 2,
      totalCount: 3,
      hasMore: true,
      nextPage: 2,
    },
    {
      key: "backlog",
      stories: [],
      loadedCount: 0,
      totalCount: 0,
      hasMore: false,
      nextPage: 1,
    },
  ],
  meta: {
    groupBy: "status",
    filters: { teamIds: ["team"] },
    orderBy: "created",
    orderDirection: "desc",
    totalGroups: 2,
  },
});
const pages = (
  groupKey: string,
  records: Story[][],
): InfiniteData<GroupStoriesResponse> => ({
  pages: records.map((stories, index) => ({
    groupKey,
    stories,
    filters: { teamIds: ["team"] },
    orderBy: "created",
    orderDirection: "desc",
    pagination: {
      page: index + 1,
      pageSize: 2,
      hasMore: index < records.length - 1,
      nextPage: index + 2,
    },
  })),
  pageParams: records.map((_, index) => index + 1),
});

let client: QueryClient;
let unsubscribes: (() => void)[];
const observe = (queryKey: QueryKey) => {
  unsubscribes.push(
    new QueryObserver(client, {
      queryKey,
      staleTime: Infinity,
    }).subscribe(() => undefined),
  );
};
const sourceKey = storyKeys.groupStories(WORKSPACE, "started", {
  ...GROUP_PARAMS,
  groupKey: "started",
});
const targetKey = storyKeys.groupStories(WORKSPACE, "backlog", {
  ...GROUP_PARAMS,
  groupKey: "backlog",
});
const groupedKey = storyKeys.teamGrouped(WORKSPACE, "team", GROUP_PARAMS);
const setup = () => {
  client.setQueryData(groupedKey, grouped());
  client.setQueryData(sourceKey, pages("started", [[first, parent], [later]]));
  observe(groupedKey);
  observe(sourceKey);
};
const update = (storyIds = [first.id, later.id]) =>
  applyBulkStoryUpdate({
    queryClient: client,
    workspaceSlug: WORKSPACE,
    storyIds,
    payload: { statusId: "backlog" },
  });

beforeEach(() => {
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  unsubscribes = [];
});
afterEach(() => {
  for (const unsubscribe of unsubscribes) unsubscribe();
  client.clear();
});

it.each(["active", "inactive", "missing"])(
  "moves first-page and later-page selections into an %s target and restores every cache",
  (targetState) => {
    setup();
    if (targetState !== "missing") {
      client.setQueryData(targetKey, pages("backlog", [[]]));
    }
    if (targetState === "active") observe(targetKey);
    const originals = [groupedKey, sourceKey, targetKey].map((key) =>
      client.getQueryData(key),
    );

    const snapshot = update([first.id, later.id, first.id]);

    expect(
      client.getQueryData<GroupedStoriesResponse>(groupedKey)?.groups,
    ).toEqual([
      expect.objectContaining({
        key: "started",
        loadedCount: 1,
        totalCount: 1,
        stories: [parent],
      }),
      expect.objectContaining({
        key: "backlog",
        loadedCount: 2,
        totalCount: 2,
        stories: expect.arrayContaining([
          expect.objectContaining({ id: first.id, statusId: "backlog" }),
          expect.objectContaining({ id: later.id, statusId: "backlog" }),
        ]),
      }),
    ]);
    const source =
      client.getQueryData<InfiniteData<GroupStoriesResponse>>(sourceKey)!;
    expect(source.pages.map((page) => page.stories)).toEqual([[parent], []]);
    expect(source.pageParams).toEqual([1, 2]);
    const target =
      client.getQueryData<InfiniteData<GroupStoriesResponse>>(targetKey)!;
    expect(
      target.pages.flatMap((page) => page.stories.map(({ id }) => id)),
    ).toEqual([first.id, later.id]);
    expect(target.pages[0].filters).toEqual({ teamIds: ["team"] });

    restoreBulkStoryUpdate(client, snapshot);

    expect(
      [groupedKey, sourceKey, targetKey].map((key) => client.getQueryData(key)),
    ).toEqual(originals);
    if (targetState === "missing") {
      expect(
        client.getQueryCache().find({ queryKey: targetKey, exact: true }),
      ).toBeUndefined();
    }
  },
);

it("keeps nested-only selections attached to the parent without changing counts or creating a destination", () => {
  setup();
  const snapshot = update(["child"]);

  const groups =
    client.getQueryData<GroupedStoriesResponse>(groupedKey)!.groups;
  expect(groups[0].totalCount).toBe(3);
  expect(groups[0].loadedCount).toBe(2);
  expect(groups[1].totalCount).toBe(0);
  const cachedParent = groups[0].stories.find(({ id }) => id === parent.id)!;
  expect(cachedParent.statusId).toBe("started");
  expect(cachedParent.subStories[0].statusId).toBe("backlog");
  expect(cachedParent.subStories[1]).toBe(parent.subStories[1]);
  expect(client.getQueryData(targetKey)).toBeUndefined();

  restoreBulkStoryUpdate(client, snapshot);
  expect(client.getQueryData(groupedKey)).toEqual(grouped());
});

it.each([
  { statusIds: ["started"] },
  { excludedStatusIds: ["backlog"] },
  { categories: ["started"] as GroupStoryParams["categories"] },
])("does not insert into a destination excluded by %j", (filters) => {
  const params = { ...GROUP_PARAMS, ...filters };
  const filteredGroupedKey = storyKeys.teamGrouped(WORKSPACE, "team", params);
  const filteredSourceKey = storyKeys.groupStories(WORKSPACE, "started", {
    ...params,
    groupKey: "started",
  });
  const filteredTargetKey = storyKeys.groupStories(WORKSPACE, "backlog", {
    ...params,
    groupKey: "backlog",
  });
  const data = grouped();
  data.meta.filters = { ...data.meta.filters, ...filters };
  client.setQueryData(filteredGroupedKey, data);
  client.setQueryData(
    filteredSourceKey,
    pages("started", [[first, parent], [later]]),
  );
  client.setQueryData(statusKeys.lists(WORKSPACE), [
    { id: "backlog", category: "backlog" },
  ]);
  observe(filteredGroupedKey);
  observe(filteredSourceKey);

  update();

  const result =
    client.getQueryData<GroupedStoriesResponse>(filteredGroupedKey)!;
  expect(result.groups[0].totalCount).toBe(1);
  expect(result.groups[1]).toEqual(data.groups[1]);
  expect(client.getQueryData(filteredTargetKey)).toBeUndefined();
});

it("preserves other workspace and filter caches while writing the matching target", () => {
  setup();
  const otherWorkspaceKey = storyKeys.groupStories(
    "other-workspace",
    "started",
    {
      ...GROUP_PARAMS,
      groupKey: "started",
    },
  );
  const otherFilterKey = storyKeys.groupStories(WORKSPACE, "backlog", {
    ...GROUP_PARAMS,
    groupKey: "backlog",
    teamIds: ["another-team"],
  });
  const otherWorkspace = pages("started", [[first]]);
  const otherFilter = pages("backlog", [[story("unrelated")]]);
  client.setQueryData(otherWorkspaceKey, otherWorkspace);
  client.setQueryData(otherFilterKey, otherFilter);
  observe(otherWorkspaceKey);
  observe(otherFilterKey);

  update();

  expect(client.getQueryData(otherWorkspaceKey)).toBe(otherWorkspace);
  expect(client.getQueryData(otherFilterKey)).toBe(otherFilter);
  expect(client.getQueryData(targetKey)).toBeDefined();
});

it("does not mutate count-only workflow caches", () => {
  const key = [...storyKeys.grouped(WORKSPACE), { groupBy: "status" }];
  const capacity = { groups: [{ key: "backlog", totalCount: 9 }] };
  client.setQueryData(key, capacity);
  observe(key);

  const snapshot = update();

  expect(client.getQueryData(key)).toBe(capacity);
  expect(snapshot.size).toBe(0);
});
