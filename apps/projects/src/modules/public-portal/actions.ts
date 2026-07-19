"use server";

import { revalidatePath } from "next/cache";
import { post } from "api-client";
import { auth } from "@/auth";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { toPublicRequest, type ApiFeedbackItem } from "./data";

type CreateFeedbackInput = {
  portalSlug: string;
  boardId: string;
  title: string;
  description: string;
};

type ItemInput = {
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

    const response = await post<ApiResponse<ApiFeedbackItem>>(
      `portals/${input.portalSlug}/feedback/items`,
      {
        boardId: input.boardId,
        description: input.description,
        title: input.title,
      },
    );
    refreshPortal(input.portalSlug);
    return {
      ...response,
      data: response.data ? toPublicRequest(response.data) : response.data,
    };
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

    const response = await post<ApiResponse<FeedbackVoteResult>>(
      `portals/${input.portalSlug}/feedback/items/${input.itemId}/vote`,
      { vote: input.vote },
    );
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

    const response = await post<ApiResponse<CreatedFeedbackComment>>(
      `portals/${input.portalSlug}/feedback/items/${input.itemId}/comments`,
      { body: input.body },
    );
    refreshPortal(input.portalSlug);
    refreshFeedbackItem(input.portalSlug, input.itemSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
