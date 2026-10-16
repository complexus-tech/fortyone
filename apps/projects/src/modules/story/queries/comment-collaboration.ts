import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import type { Comment } from "@/types/comment";
import { get, put, remove } from "@/lib/http";
import type {
  CommentEmoji,
  CommentState,
  StoryCommentState,
} from "../types/comment-collaboration";

const responseData = <T>(response: ApiResponse<T>): T => {
  if (response.error) throw new Error(response.error.message);
  if (!response.data) throw new Error("Comment data is unavailable");
  return response.data;
};

export const getCommentState = async (storyId: string, ctx: WorkspaceCtx) =>
  responseData(
    await get<ApiResponse<StoryCommentState>>(
      `stories/${storyId}/comments/state`,
      ctx,
    ),
  );

export const getCommentThread = async (
  storyId: string,
  commentId: string,
  ctx: WorkspaceCtx,
) =>
  responseData(
    await get<ApiResponse<Comment>>(
      `stories/${storyId}/comments/${commentId}/thread`,
      ctx,
    ),
  );

export const markCommentsRead = async (
  storyId: string,
  observedAt: string,
  ctx: WorkspaceCtx,
) =>
  responseData(
    await put<{ observedAt: string }, ApiResponse<StoryCommentState>>(
      `stories/${storyId}/comments/read`,
      { observedAt },
      ctx,
    ),
  );

export const setCommentReaction = async (
  commentId: string,
  emoji: CommentEmoji,
  reacted: boolean,
  ctx: WorkspaceCtx,
) => {
  const path = `comments/${commentId}/reactions`;
  return responseData(
    reacted
      ? await put<{ emoji: CommentEmoji }, ApiResponse<CommentState>>(
          path,
          { emoji },
          ctx,
        )
      : await remove<ApiResponse<CommentState>>(path, ctx, { json: { emoji } }),
  );
};

export const setCommentResolution = async (
  commentId: string,
  resolved: boolean,
  ctx: WorkspaceCtx,
) =>
  responseData(
    await put<{ resolved: boolean }, ApiResponse<CommentState>>(
      `comments/${commentId}/resolution`,
      { resolved },
      ctx,
    ),
  );
