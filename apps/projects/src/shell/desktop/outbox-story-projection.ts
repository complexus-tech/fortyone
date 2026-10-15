import type { QueryClient, InfiniteData } from "@tanstack/react-query";
import type { DetailedStory } from "@/modules/story/types";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
  GroupStoryParams,
} from "@/modules/stories/types";
import { storyKeys } from "@/modules/stories/constants";
import {
  computeTargetKey,
  patchStories,
} from "@/modules/stories/utils/optimistic";
import type { DesktopCacheScope } from "./cache-types";
import { acceptsStory, queryFilters } from "./outbox-filters";
import {
  findResource,
  isRecord,
  patchResource,
  setChanged,
} from "./outbox-records";

export function projectGroups(
  data: GroupedStoriesResponse,
  story: DetailedStory,
  allowed: boolean,
) {
  const target =
    data.meta.groupBy === "none"
      ? data.groups[0]?.key ?? "none"
      : computeTargetKey(data.meta.groupBy, story);
  const groups = data.groups.map((group) => {
    const patched = patchStories(group.stories, story.id, story);
    const had = group.stories.some((item) => item.id === story.id);
    const keep = allowed && group.key === target;
    const stories = keep
      ? patched
      : patched.filter((item) => item.id !== story.id);
    const removed = had && !keep ? 1 : 0;
    return {
      ...group,
      stories,
      loadedCount: Math.max(0, group.loadedCount - removed),
      totalCount: Math.max(0, group.totalCount - removed),
    };
  });
  if (
    allowed &&
    !groups.some((group) =>
      group.stories.some((item) => item.id === story.id),
    ) &&
    target
  ) {
    let destination = groups.find((group) => group.key === target);
    if (!destination) {
      destination = {
        key: target,
        stories: [],
        loadedCount: 0,
        totalCount: 0,
        hasMore: false,
        nextPage: 1,
      };
      groups.push(destination);
    }
    destination.stories = [story, ...destination.stories];
    destination.loadedCount++;
    destination.totalCount++;
  }
  return {
    ...data,
    groups,
    meta: { ...data.meta, totalGroups: groups.length },
  };
}

function projectInfinite(
  data: InfiniteData<GroupStoriesResponse>,
  story: DetailedStory,
  allowed: boolean,
) {
  const pages = data.pages.map((page) => {
    if (!Array.isArray(page.stories)) return page;
    const patched = patchStories(page.stories, story.id, story);
    const stories = allowed
      ? patched
      : patched.filter((item) => item.id !== story.id);
    return { ...page, stories };
  });
  if (
    allowed &&
    pages.length > 0 &&
    !pages.some((page) => page.stories.some((item) => item.id === story.id))
  )
    pages[0] = { ...pages[0], stories: [story, ...pages[0].stories] };
  return { ...data, pages };
}

export function projectStory(
  client: QueryClient,
  scope: DesktopCacheScope,
  story: DetailedStory,
) {
  const entries = client.getQueriesData({
    queryKey: storyKeys.all(scope.workspaceSlug),
  });
  for (const [key, data] of entries) {
    if (!isRecord(data)) continue;
    if (Array.isArray(data.groups) && isRecord(data.meta)) {
      const grouped = data as GroupedStoriesResponse;
      setChanged(
        client,
        key,
        projectGroups(
          grouped,
          story,
          acceptsStory(
            client,
            scope,
            story,
            queryFilters(key, grouped.meta.filters),
          ),
        ),
      );
    } else if (
      key[2] === "group" &&
      Array.isArray(data.pages) &&
      isRecord(key[4])
    ) {
      const params = key[4];
      const target = computeTargetKey(
        params.groupBy as GroupStoryParams["groupBy"],
        story,
      );
      const accepted = acceptsStory(client, scope, story, params);
      const allowed =
        accepted && (params.groupBy === "none" || target === key[3]);
      const infinite = data as unknown as InfiniteData<GroupStoriesResponse>;
      const hadTopLevel = infinite.pages.some((page) =>
        page.stories.some((item) => item.id === story.id),
      );
      setChanged(client, key, projectInfinite(infinite, story, allowed));
      if (hadTopLevel && accepted && !allowed && target) {
        const targetKey = storyKeys.groupStories(scope.workspaceSlug, target, {
          ...params,
          groupKey: target,
        });
        const current =
          client.getQueryData<InfiniteData<GroupStoriesResponse>>(targetKey);
        const initial = current ?? {
          pageParams: [1],
          pages: [
            {
              ...infinite.pages[0],
              groupKey: target,
              stories: [],
              filters: params,
              pagination: {
                page: 1,
                pageSize: Number(params.pageSize) || 20,
                hasMore: false,
                nextPage: 2,
              },
            },
          ],
        };
        setChanged(client, targetKey, projectInfinite(initial, story, true));
      }
    } else if (key[2] === "detail" && key.length === 4) {
      const updated = patchResource(data, story.id, story);
      if (
        story.parentId &&
        data.id === story.parentId &&
        Array.isArray(data.subStories) &&
        !findResource(data.subStories, story.id)
      )
        setChanged(client, key, {
          ...data,
          subStories: [...data.subStories, story],
        });
      else setChanged(client, key, updated);
    }
  }
}

export function findCachedStory(
  client: QueryClient,
  scope: DesktopCacheScope,
  id: string,
) {
  const detail = client.getQueryData<DetailedStory>(
    storyKeys.detail(scope.workspaceSlug, id),
  );
  if (detail) return detail;
  for (const [, data] of client.getQueriesData({
    queryKey: storyKeys.all(scope.workspaceSlug),
  })) {
    const story = findResource(data, id);
    if (story && typeof story.title === "string") return story as DetailedStory;
  }
  return null;
}
