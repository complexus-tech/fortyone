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

type VoteInput = ItemInput & {
  vote: -1 | 1;
};

type CommentInput = ItemInput & {
  body: string;
};

export type FeedbackVoteResult = {
  vote: -1 | 0 | 1;
  voteCount: number;
};

export type CreatedFeedbackComment = {
  id: string;
  authorName: string;
  authorAvatar?: string | null;
  body: string;
  createdAt: string;
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
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Please log in to submit feedback" },
      };
    }

    const ctx = { session, workspaceSlug: input.workspaceSlug };
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

export const toggleFeedbackVoteAction = async (input: VoteInput) => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Please log in to vote" },
      };
    }

    const ctx = { session, workspaceSlug: input.workspaceSlug };
    const response = await post<
      { vote: -1 | 1 },
      ApiResponse<FeedbackVoteResult>
    >(`feedback/items/${input.itemId}/vote`, { vote: input.vote }, ctx);
    refreshPortal(input.portalSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const createFeedbackCommentAction = async (input: CommentInput) => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Please log in to comment" },
      };
    }

    const ctx = { session, workspaceSlug: input.workspaceSlug };
    const response = await post<
      { body: string },
      ApiResponse<CreatedFeedbackComment>
    >(`feedback/items/${input.itemId}/comments`, { body: input.body }, ctx);
    refreshPortal(input.portalSlug);
    refreshFeedbackItem(input.portalSlug, input.itemSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
