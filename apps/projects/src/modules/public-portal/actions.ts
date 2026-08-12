"use server";

import { revalidatePath } from "next/cache";
import { post } from "api-client";
import { auth } from "@/auth";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { toPublicRequest, type ApiFeedbackItem } from "./data";
import {
  createFeedbackIngressHeaders,
  normalizeFeedbackPortalSlug,
} from "./feedback-ingress";

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
  parentId?: string;
};

export type FeedbackVoteResult = {
  vote: -1 | 0 | 1;
  voteCount: number;
};

export type CreatedFeedbackComment = {
  id: string;
  parentId?: string | null;
  authorName: string;
  authorAvatar?: string | null;
  body: string;
  createdAt: string;
};

export type CreateFeedbackResult =
  | {
      kind: "authenticated";
      request: ReturnType<typeof toPublicRequest>;
    }
  | {
      kind: "anonymous";
      request: ReturnType<typeof toPublicRequest>;
    };

const toCreateFeedbackResult = (
  item: ApiFeedbackItem,
): CreateFeedbackResult => {
  const request = toPublicRequest(item);

  if (item.anonymous) {
    return {
      kind: "anonymous",
      request,
    };
  }

  return { kind: "authenticated", request };
};

const refreshPortal = (portalSlug: string) => {
  revalidatePath("/feedback");
  revalidatePath("/feedback/roadmap");
  revalidatePath("/updates");
  revalidatePath(`/portal/${portalSlug}`);
  revalidatePath(`/portal/${portalSlug}/feedback`);
  revalidatePath(`/portal/${portalSlug}/feedback/roadmap`);
};

const refreshFeedbackItem = (portalSlug: string, itemSlug?: string) => {
  if (!itemSlug) return;
  revalidatePath(`/feedback/${itemSlug}`);
  revalidatePath(`/portal/${portalSlug}/feedback/${itemSlug}`);
  revalidatePath(`/portal/${portalSlug}/requests/${itemSlug}`);
};

const submitFeedback = async (
  input: CreateFeedbackInput,
  participationIntent: "account" | "anonymous",
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const session = await auth();
    if (participationIntent === "account" && !session) {
      return {
        data: null,
        error: {
          message: "Your session expired. Please log in and try again.",
        },
      };
    }
    const isAnonymous = participationIntent === "anonymous";
    const ingressHeaders =
      isAnonymous || !session
        ? await createFeedbackIngressHeaders(portalSlug)
        : undefined;
    const response = await post<ApiResponse<ApiFeedbackItem>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/items`,
      {
        boardId: input.boardId,
        description: input.description,
        participationIntent,
        title: input.title,
        website: "",
      },
      ingressHeaders
        ? {
            credentials: isAnonymous ? "omit" : "include",
            headers: ingressHeaders,
          }
        : undefined,
    );
    refreshPortal(portalSlug);
    return {
      ...response,
      data: response.data ? toCreateFeedbackResult(response.data) : null,
    };
  } catch (error) {
    return getApiError(error);
  }
};

export const createFeedbackAction = async (input: CreateFeedbackInput) =>
  submitFeedback(input, "account");

export const createAnonymousFeedbackAction = async (
  input: CreateFeedbackInput,
) => submitFeedback(input, "anonymous");

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
      {
        body: input.body,
        ...(input.parentId ? { parentId: input.parentId } : {}),
      },
    );
    refreshPortal(input.portalSlug);
    refreshFeedbackItem(input.portalSlug, input.itemSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
