import type { Session } from "next-auth";
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
  session: Session,
  page: number = 1,
) => {
  const response = await get<ApiResponse<CommentsResponse>>(
    `stories/${id}/comments?page=${page}`,
    session,
  );
  return response.data!;
};
