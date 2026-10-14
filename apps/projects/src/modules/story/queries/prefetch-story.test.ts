/* global afterEach, beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { QueryClient } from "@tanstack/react-query";
import type { WorkspaceCtx } from "@/lib/http";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "@/lib/queries/links/get-links";
import { storyKeys } from "@/modules/stories/constants";
import { preloadMainDetails } from "../load-main-details";
import { getStoryAttachments } from "./get-attachments";
import { getStory } from "./get-story";
import { prefetchStory } from "./prefetch-story";

jest.mock("./get-story", () => ({ getStory: jest.fn() }));
jest.mock("./get-attachments", () => ({ getStoryAttachments: jest.fn() }));
jest.mock("@/lib/queries/links/get-links", () => ({ getLinks: jest.fn() }));
jest.mock("../load-main-details", () => ({
  preloadMainDetails: jest.fn(),
}));

const ctx = {
  session: { token: "test-session" },
  workspaceSlug: "first-workspace",
} satisfies WorkspaceCtx;
const mockedGetStory = jest.mocked(getStory);
const mockedGetAttachments = jest.mocked(getStoryAttachments);
const mockedGetLinks = jest.mocked(getLinks);
let queryClient: QueryClient;

beforeEach(() => {
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  mockedGetStory.mockReset().mockResolvedValue(null);
  mockedGetAttachments.mockReset().mockResolvedValue([]);
  mockedGetLinks.mockReset().mockResolvedValue([]);
  jest.mocked(preloadMainDetails).mockReset().mockResolvedValue(undefined);
});

afterEach(() => {
  queryClient.clear();
});

it("fetches the three ticket resources once across concurrent and repeated hovers", async () => {
  await Promise.all([
    prefetchStory(queryClient, "story", ctx),
    prefetchStory(queryClient, "story", ctx),
  ]);
  await prefetchStory(queryClient, "story", ctx);

  expect(mockedGetStory).toHaveBeenCalledTimes(1);
  expect(mockedGetAttachments).toHaveBeenCalledTimes(1);
  expect(mockedGetLinks).toHaveBeenCalledTimes(1);
  expect(mockedGetStory).toHaveBeenCalledWith("story", ctx);
  expect(mockedGetAttachments).toHaveBeenCalledWith("story", ctx);
  expect(mockedGetLinks).toHaveBeenCalledWith("story", ctx);
});

it("starts every ticket resource without waiting for the editor module", async () => {
  let finishPreload!: () => void;
  jest.mocked(preloadMainDetails).mockReturnValue(
    new Promise<void>((resolve) => {
      finishPreload = resolve;
    }),
  );

  const prefetch = prefetchStory(queryClient, "story", ctx);
  expect(preloadMainDetails).toHaveBeenCalledTimes(1);
  expect(mockedGetStory).toHaveBeenCalledTimes(1);
  expect(mockedGetAttachments).toHaveBeenCalledTimes(1);
  expect(mockedGetLinks).toHaveBeenCalledTimes(1);

  finishPreload();
  await prefetch;
});

it("refreshes each resource when its existing reader freshness expires", async () => {
  await prefetchStory(queryClient, "story", ctx);
  const expiredAt = (minutes: number) =>
    Date.now() - DURATION_FROM_MILLISECONDS.MINUTE * minutes - 1;

  queryClient.setQueryData(storyKeys.detail(ctx.workspaceSlug, "story"), null, {
    updatedAt: expiredAt(1),
  });
  await prefetchStory(queryClient, "story", ctx);
  expect(mockedGetStory).toHaveBeenCalledTimes(2);
  expect(mockedGetAttachments).toHaveBeenCalledTimes(1);
  expect(mockedGetLinks).toHaveBeenCalledTimes(1);

  queryClient.setQueryData(
    storyKeys.attachments(ctx.workspaceSlug, "story"),
    [],
    {
      updatedAt: expiredAt(5),
    },
  );
  queryClient.setQueryData(linkKeys.story("story"), [], {
    updatedAt: expiredAt(10),
  });
  await prefetchStory(queryClient, "story", ctx);
  expect(mockedGetStory).toHaveBeenCalledTimes(2);
  expect(mockedGetAttachments).toHaveBeenCalledTimes(2);
  expect(mockedGetLinks).toHaveBeenCalledTimes(2);
});

it("does not reuse detail or attachment data from a different workspace", async () => {
  const otherCtx = { ...ctx, workspaceSlug: "second-workspace" };
  await prefetchStory(queryClient, "story", ctx);
  await prefetchStory(queryClient, "story", otherCtx);

  expect(mockedGetStory).toHaveBeenCalledTimes(2);
  expect(mockedGetAttachments).toHaveBeenCalledTimes(2);
  expect(mockedGetStory).toHaveBeenLastCalledWith("story", otherCtx);
  expect(mockedGetAttachments).toHaveBeenLastCalledWith("story", otherCtx);
});

it("refreshes invalidated ticket details and attachments within their freshness windows", async () => {
  await prefetchStory(queryClient, "story", ctx);
  await queryClient.invalidateQueries({
    queryKey: storyKeys.detail(ctx.workspaceSlug, "story"),
  });
  await prefetchStory(queryClient, "story", ctx);

  expect(mockedGetStory).toHaveBeenCalledTimes(2);
  expect(mockedGetAttachments).toHaveBeenCalledTimes(2);
  expect(mockedGetLinks).toHaveBeenCalledTimes(1);
});
