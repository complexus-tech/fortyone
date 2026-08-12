"use server";

import { revalidatePath } from "next/cache";
import { post } from "api-client";
import { auth } from "@/auth";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import {
  toPublicRequest,
  type ApiFeedbackItem,
} from "@/modules/public-portal/data";
import {
  createFeedbackIngressHeaders,
  normalizeFeedbackPortalSlug,
} from "@/modules/public-portal/feedback-ingress";

export type CreateWidgetFeedbackInput = {
  boardId: string;
  description: string;
  isAnonymous: boolean;
  portalSlug: string;
  title: string;
};

export type CreateWidgetFeedbackResult = {
  isAnonymous: boolean;
  request: ReturnType<typeof toPublicRequest>;
};

export const createWidgetFeedbackAction = async (
  input: CreateWidgetFeedbackInput,
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const session = await auth();
    if (!input.isAnonymous && !session) {
      return {
        data: null,
        error: {
          message:
            "Your session expired. Open the full portal to log in again.",
        },
      };
    }
    const ingressHeaders =
      input.isAnonymous || !session
        ? await createFeedbackIngressHeaders(portalSlug)
        : undefined;
    const response = await post<ApiResponse<ApiFeedbackItem>>(
      `portals/${encodeURIComponent(portalSlug)}/widget/feedback/items`,
      {
        boardId: input.boardId,
        description: input.description,
        participationIntent: input.isAnonymous ? "anonymous" : "account",
        title: input.title,
        website: "",
      },
      ingressHeaders
        ? {
            credentials: input.isAnonymous ? "omit" : "include",
            headers: ingressHeaders,
          }
        : undefined,
    );
    revalidatePath(`/embed/feedback/${portalSlug}`);

    return {
      ...response,
      data: response.data
        ? {
            isAnonymous: response.data.anonymous === true,
            request: toPublicRequest(response.data),
          }
        : response.data,
    };
  } catch (error) {
    return getApiError(error);
  }
};
