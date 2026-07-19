"use server";

import { revalidatePath } from "next/cache";
import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type CreateFeedbackInput = {
  workspaceSlug: string;
  portalSlug: string;
  portalId: string;
  boardId: string;
  title: string;
  description: string;
};

type ItemInput = {
  workspaceSlug: string;
  portalSlug: string;
  itemId: string;
  itemSlug?: string;
};

type CommentInput = ItemInput & {
  body: string;
};

type CreateStoryInput = ItemInput & {
  teamId: string;
};

const workspaceCtx = async (workspaceSlug: string) => {
  const session = await auth();
  return { session: session!, workspaceSlug };
};

const refreshPortal = (portalSlug: string) => {
  revalidatePath("/feedback");
  revalidatePath("/roadmap");
  revalidatePath("/updates");
  revalidatePath(`/portal/${portalSlug}`);
  revalidatePath(`/portal/${portalSlug}/feedback`);
  revalidatePath(`/portal/${portalSlug}/roadmap`);
};

const refreshFeedbackItem = (portalSlug: string, itemSlug?: string) => {
  if (!itemSlug) return;
  revalidatePath(`/feedback/${itemSlug}`);
  revalidatePath(`/portal/${portalSlug}/feedback/${itemSlug}`);
  revalidatePath(`/portal/${portalSlug}/requests/${itemSlug}`);
};

export const createFeedbackAction = async (input: CreateFeedbackInput) => {
  try {
    const ctx = await workspaceCtx(input.workspaceSlug);
    const response = await post<
      {
        boardId: string;
        description: string;
        portalId: string;
        title: string;
      },
      ApiResponse<unknown>
    >(
      "feedback/items",
      {
        boardId: input.boardId,
        description: input.description,
        portalId: input.portalId,
        title: input.title,
      },
      ctx,
    );
    refreshPortal(input.portalSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const toggleFeedbackVoteAction = async (input: ItemInput) => {
  try {
    const ctx = await workspaceCtx(input.workspaceSlug);
    const response = await post<Record<string, never>, ApiResponse<unknown>>(
      `feedback/items/${input.itemId}/vote`,
      {},
      ctx,
    );
    refreshPortal(input.portalSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const createFeedbackCommentAction = async (input: CommentInput) => {
  try {
    const ctx = await workspaceCtx(input.workspaceSlug);
    const response = await post<{ body: string }, ApiResponse<unknown>>(
      `feedback/items/${input.itemId}/comments`,
      { body: input.body },
      ctx,
    );
    refreshPortal(input.portalSlug);
    refreshFeedbackItem(input.portalSlug, input.itemSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const createStoryFromFeedbackAction = async (
  input: CreateStoryInput,
) => {
  try {
    const ctx = await workspaceCtx(input.workspaceSlug);
    const response = await post<{ teamId: string }, ApiResponse<unknown>>(
      `feedback/items/${input.itemId}/story`,
      { teamId: input.teamId },
      ctx,
    );
    refreshPortal(input.portalSlug);
    refreshFeedbackItem(input.portalSlug, input.itemSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
