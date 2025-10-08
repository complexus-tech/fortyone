import { get } from "@/lib/http";
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
  page = 1
): Promise<CommentsResponse> => {
  const response = await get<ApiResponse<CommentsResponse>>(
    `stories/${id}/comments?page=${page}`
  );
  return response.data!;
};
