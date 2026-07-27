import { cache } from "react";
import { auth } from "@/auth";
import { getStory, getStoryRef } from "@/modules/story/queries/get-story";
import { isStoryUuid } from "@/modules/story/utils/story-url";

const getSession = cache(async () => auth());

export const getStoryRouteData = cache(
  async (identifier: string, workspaceSlug: string) => {
    const session = await getSession();
    if (!session) return null;

    const ctx = { session, workspaceSlug };
    return isStoryUuid(identifier)
      ? getStory(identifier, ctx)
      : getStoryRef(identifier, ctx);
  },
);
