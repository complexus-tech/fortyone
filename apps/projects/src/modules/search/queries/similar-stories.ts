import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SimilarStory } from "../types";

export type SimilarStoriesQueryParams = {
  title: string;
  teamId?: string;
  limit?: number;
};

export const similarStoriesQuery = async (
  ctx: WorkspaceCtx,
  params: SimilarStoriesQueryParams,
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const response = await get<ApiResponse<SimilarStory[]>>(
    `search/similar-stories${query}`,
    ctx,
  );
  return response.data ?? [];
};
