import type { QueryClient } from "@tanstack/react-query";
import type { WorkspaceCtx } from "@/lib/http";
import { linkKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getLinks } from "@/lib/queries/links/get-links";
import { storyKeys } from "@/modules/stories/constants";
import { getStoryAttachments } from "./get-attachments";
import { getStory } from "./get-story";

export const prefetchStory = (
  queryClient: QueryClient,
  storyId: string,
  ctx: WorkspaceCtx,
) =>
  Promise.all([
    queryClient.prefetchQuery({
      queryKey: storyKeys.detail(ctx.workspaceSlug, storyId),
      queryFn: () => getStory(storyId, ctx),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
    }),
    queryClient.prefetchQuery({
      queryKey: storyKeys.attachments(ctx.workspaceSlug, storyId),
      queryFn: () => getStoryAttachments(storyId, ctx),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
    }),
    queryClient.prefetchQuery({
      queryKey: linkKeys.story(storyId),
      queryFn: () => getLinks(storyId, ctx),
      staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    }),
  ]);
