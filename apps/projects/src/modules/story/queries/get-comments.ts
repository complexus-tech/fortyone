import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { Comment, ApiResponse } from "@/types";

type CommentsResponse = {
  comments: Comment[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};

export const getStoryComments = async (
  id: string,
  ctx: WorkspaceCtx,
  page = 1,
  pageSize = 20,
) => {
  const response = await get<ApiResponse<CommentsResponse>>(
    `stories/${id}/comments?page=${page}&pageSize=${pageSize}`,
    ctx,
  );
  return response.data!;
};
