/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";
import { syncFinalizedStoryMediaCache } from "./use-story-description-media";

describe("syncFinalizedStoryMediaCache", () => {
  it("updates the detail cache and invalidates all story queries", () => {
    const invalidateQueries = jest.fn().mockResolvedValue(undefined);
    const setQueryData = jest.fn();
    const queryClient = {
      invalidateQueries,
      setQueryData,
    } as unknown as Parameters<typeof syncFinalizedStoryMediaCache>[0];
    const content = {
      contentHtml: '<p>Brief</p><img src="/media/image.png">',
      contentText: "Brief",
    };

    syncFinalizedStoryMediaCache(queryClient, "workspace", "story-1", content);

    expect(setQueryData).toHaveBeenCalledWith(
      storyKeys.detail("workspace", "story-1"),
      expect.any(Function),
    );
    const updateCachedStory = setQueryData.mock.calls[0][1] as (
      story: DetailedStory | undefined,
    ) => DetailedStory | undefined;
    const cachedStory = {
      description: "Old",
      descriptionHTML: "<p>Old</p>",
      id: "story-1",
      title: "Story",
    } as DetailedStory;
    expect(updateCachedStory(cachedStory)).toEqual({
      ...cachedStory,
      description: content.contentText,
      descriptionHTML: content.contentHtml,
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.all("workspace"),
    });
  });
});
