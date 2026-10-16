import { QueryClient } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { getStoryRouteData } from "@/modules/story/server/get-story-route-data";
import Page from "./page";

jest.mock("@/modules/story", () => ({ StoryPage: () => null }));
jest.mock("@/modules/story/components/story-route-not-found", () => ({
  StoryRouteNotFound: () => null,
}));
jest.mock("@/modules/story/server/get-story-route-data", () => ({
  getStoryRouteData: jest.fn(),
}));
jest.mock("@/app/get-query-client", () => ({ getQueryClient: jest.fn() }));
jest.mock("@/utils", () => ({
  withWorkspacePath: (path: string, workspaceSlug: string) =>
    `/${workspaceSlug}${path}`,
}));
jest.mock("next/navigation", () => ({
  redirect: (path: string) => {
    throw new Error(`redirect:${path}`);
  },
}));

const story = {
  id: "11111111-1111-4111-8111-111111111111",
  sequenceId: 42,
  teamCode: "eng",
  title: "Checkout spacing",
} as NonNullable<Awaited<ReturnType<typeof getStoryRouteData>>>;

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(getStoryRouteData).mockResolvedValue(story);
  jest.mocked(getQueryClient).mockReturnValue(new QueryClient());
});

it("preserves comment targets and repeated query values when canonicalizing a UUID link", async () => {
  const commentId = "55555555-5555-4555-8555-555555555555";
  await expect(
    Page({
      params: Promise.resolve({ storyRef: story.id, workspaceSlug: "acme" }),
      searchParams: Promise.resolve({
        comment: commentId,
        filter: ["first", "second"],
      }),
    }),
  ).rejects.toThrow(
    `redirect:/acme/work/ENG-42?comment=${commentId}&filter=first&filter=second`,
  );
});

it("keeps canonical redirects clean when no query is supplied", async () => {
  await expect(
    Page({
      params: Promise.resolve({ storyRef: story.id, workspaceSlug: "acme" }),
      searchParams: Promise.resolve({}),
    }),
  ).rejects.toThrow("redirect:/acme/work/ENG-42");
});

it("renders an already canonical story with its comment query intact", async () => {
  const page = await Page({
    params: Promise.resolve({ storyRef: "ENG-42", workspaceSlug: "acme" }),
    searchParams: Promise.resolve({
      comment: "55555555-5555-4555-8555-555555555555",
    }),
  });
  expect(page.props.children.props.storyId).toBe(story.id);
});
