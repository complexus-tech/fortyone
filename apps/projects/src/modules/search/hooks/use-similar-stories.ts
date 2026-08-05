import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import {
  similarStoriesQuery,
  type SimilarStoriesQueryParams,
} from "../queries/similar-stories";

const similarStoryKeys = {
  query: (workspaceSlug: string, params: SimilarStoriesQueryParams) =>
    ["search", workspaceSlug, "similar-stories", params] as const,
};

export const useSimilarStories = (params: SimilarStoriesQueryParams) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: similarStoryKeys.query(workspaceSlug, params),
    queryFn: () =>
      similarStoriesQuery({ session: session!, workspaceSlug }, params),
    enabled: params.title.trim().length >= 3,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
  });
};
