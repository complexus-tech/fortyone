import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import type { Comment } from "@/types/comment";
import { get, put, remove } from "@/lib/http";
import type {
  CommentEmoji,
  CommentReadInput,
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
  input: CommentReadInput,
  ctx: WorkspaceCtx,
) =>
  responseData(
    await put<CommentReadInput, ApiResponse<StoryCommentState>>(
      `stories/${storyId}/comments/read`,
      input,
      ctx,
    ),
  );

export const markCommentSeen = async (
  storyId: string,
  commentId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await put<undefined, ApiResponse<null>>(
    `stories/${storyId}/comments/${commentId}/seen`,
    undefined,
    ctx,
    { retry: 0 },
  );
  if (response.error) throw new Error(response.error.message);
};

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
