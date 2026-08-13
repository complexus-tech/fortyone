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
import type {
  PublicParticipantKind,
  PublicPortalGuestParticipant,
  PublicRequestComment,
} from "@/modules/public-portal/types";
import { getTrustedWidgetOrigin } from "./protocol";

type WidgetContributorSession = {
  expiresAt: string;
  token: string;
};

type ApiWidgetParticipant = {
  avatarUrl?: string | null;
  displayName: string;
  email?: string;
  id: string;
  kind: "external" | "verified_guest";
  masked: boolean;
  publicName?: string;
  unreadUpdateCount?: number;
};

type ApiWidgetSessionResult = {
  participant: ApiWidgetParticipant;
  session: WidgetContributorSession;
  unreadUpdateCount?: number;
};

export type WidgetParticipantSession = {
  participant: PublicPortalGuestParticipant;
  session: WidgetContributorSession;
};

const toWidgetParticipantSession = (
  result: ApiWidgetSessionResult,
): WidgetParticipantSession => ({
  participant: {
    avatarUrl: result.participant.avatarUrl ?? null,
    canReceiveUpdates: true,
    displayName: result.participant.displayName,
    email: result.participant.email,
    id: result.participant.id,
    kind: result.participant.kind,
    masked: result.participant.masked,
    name:
      result.participant.publicName ||
      result.participant.displayName ||
      "Customer",
    sessionExpiresAt: result.session.expiresAt,
    unreadUpdateCount:
      result.unreadUpdateCount ?? result.participant.unreadUpdateCount ?? 0,
  },
  session: result.session,
});

export type CreateWidgetFeedbackInput = {
  boardId: string;
  description: string;
  participationIntent: "account" | "anonymous" | "external" | "verified_guest";
  portalSlug: string;
  sessionToken?: string;
  title: string;
};

export type CreateWidgetFeedbackResult = {
  participantKind: PublicParticipantKind;
  request: ReturnType<typeof toPublicRequest>;
};

export type WidgetVoteResult = {
  participantKind?: PublicParticipantKind;
  vote: -1 | 0 | 1;
  voteCount: number;
};

type ApiWidgetComment = {
  authorAvatar?: string | null;
  authorName: string;
  body: string;
  createdAt: string;
  id: string;
  parentId?: string | null;
  participantKind?: PublicParticipantKind;
};

const widgetWriteOptions = (
  participantKind: PublicParticipantKind,
  sessionToken?: string,
) => {
  if (participantKind === "account") return undefined;
  if (
    (participantKind === "external" || participantKind === "verified_guest") &&
    sessionToken
  ) {
    return {
      credentials: "omit" as const,
      headers: { Authorization: `FeedbackSession ${sessionToken}` },
    };
  }

  throw new Error("Verify your email to continue.");
};

export const createWidgetFeedbackAction = async (
  input: CreateWidgetFeedbackInput,
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const session =
      input.participationIntent === "account" ? await auth() : null;
    if (input.participationIntent === "account" && !session) {
      return {
        data: null,
        error: {
          message:
            "Your session expired. Open the full portal to log in again.",
        },
      };
    }
    const usesContributorSession =
      input.participationIntent === "external" ||
      input.participationIntent === "verified_guest";
    if (usesContributorSession && !input.sessionToken) {
      return {
        data: null,
        error: { message: "Your feedback identity has expired. Verify again." },
      };
    }
    const ingressHeaders =
      input.participationIntent === "anonymous"
        ? await createFeedbackIngressHeaders(portalSlug)
        : undefined;
    const contributorHeaders = input.sessionToken
      ? { Authorization: `FeedbackSession ${input.sessionToken}` }
      : undefined;
    const response = await post<ApiResponse<ApiFeedbackItem>>(
      `portals/${encodeURIComponent(portalSlug)}/widget/feedback/items`,
      {
        boardId: input.boardId,
        description: input.description,
        participationIntent: input.participationIntent,
        title: input.title,
        website: "",
      },
      ingressHeaders || contributorHeaders
        ? {
            credentials:
              input.participationIntent === "account" ? "include" : "omit",
            headers: ingressHeaders ?? contributorHeaders,
          }
        : undefined,
    );
    revalidatePath(`/embed/feedback/${portalSlug}`);

    return {
      ...response,
      data: response.data
        ? {
            participantKind:
              response.data.participantKind ??
              (response.data.anonymous === true
                ? "anonymous"
                : input.participationIntent),
            request: toPublicRequest(response.data),
          }
        : response.data,
    };
  } catch (error) {
    return getApiError(error);
  }
};

export const toggleWidgetFeedbackVoteAction = async (input: {
  itemId: string;
  participantKind: PublicParticipantKind;
  portalSlug: string;
  sessionToken?: string;
  vote: -1 | 1;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const response = await post<ApiResponse<WidgetVoteResult>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/items/${encodeURIComponent(input.itemId)}/vote`,
      { vote: input.vote },
      widgetWriteOptions(input.participantKind, input.sessionToken),
    );
    revalidatePath(`/embed/feedback/${portalSlug}`);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const createWidgetFeedbackCommentAction = async (input: {
  body: string;
  itemId: string;
  parentId?: string;
  participantKind: PublicParticipantKind;
  portalSlug: string;
  sessionToken?: string;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const response = await post<ApiResponse<ApiWidgetComment>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/items/${encodeURIComponent(input.itemId)}/comments`,
      {
        body: input.body.trim(),
        ...(input.parentId ? { parentId: input.parentId } : {}),
      },
      widgetWriteOptions(input.participantKind, input.sessionToken),
    );
    revalidatePath(`/embed/feedback/${portalSlug}`);
    return {
      ...response,
      data: response.data
        ? ({
            authorAvatar: response.data.authorAvatar,
            authorName: response.data.authorName,
            body: response.data.body,
            createdAtLabel: "Just now",
            id: response.data.id,
            parentId: response.data.parentId,
            participantKind: response.data.participantKind,
          } satisfies PublicRequestComment)
        : response.data,
    };
  } catch (error) {
    return getApiError(error);
  }
};

export const requestWidgetFeedbackVerificationAction = async (input: {
  displayName?: string;
  email: string;
  hideNamePublicly?: boolean;
  portalSlug: string;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const ingressHeaders = await createFeedbackIngressHeaders(portalSlug);
    return await post<ApiResponse<{ accepted: true; expiresAt: string }>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/verifications`,
      {
        displayName: input.displayName?.trim() || undefined,
        email: input.email.trim().toLowerCase(),
        hideNamePublicly: input.hideNamePublicly,
        source: "widget",
      },
      { credentials: "omit", headers: ingressHeaders },
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const confirmWidgetFeedbackVerificationAction = async (input: {
  code: string;
  email: string;
  portalSlug: string;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const ingressHeaders = await createFeedbackIngressHeaders(portalSlug);
    const response = await post<ApiResponse<ApiWidgetSessionResult>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/verifications/confirm`,
      {
        code: input.code.trim(),
        email: input.email.trim().toLowerCase(),
        source: "widget",
      },
      { credentials: "omit", headers: ingressHeaders },
    );
    return {
      ...response,
      data: response.data
        ? toWidgetParticipantSession(response.data)
        : response.data,
    };
  } catch (error) {
    return getApiError(error);
  }
};

export const exchangeWidgetIdentityAction = async (input: {
  assertion: string;
  parentOrigin: string;
  portalSlug: string;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const parentOrigin = getTrustedWidgetOrigin(input.parentOrigin);
    if (!parentOrigin || parentOrigin !== input.parentOrigin) {
      throw new Error("The widget parent origin is invalid");
    }
    if (!input.assertion || input.assertion.length > 16384) {
      throw new Error("The signed identity assertion is invalid");
    }
    const ingressHeaders = await createFeedbackIngressHeaders(portalSlug);
    const response = await post<ApiResponse<ApiWidgetSessionResult>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/widget/sessions`,
      { assertion: input.assertion, parentOrigin },
      { credentials: "omit", headers: ingressHeaders },
    );
    return {
      ...response,
      data: response.data
        ? toWidgetParticipantSession(response.data)
        : response.data,
    };
  } catch (error) {
    return getApiError(error);
  }
};

export const revokeWidgetIdentityAction = async (input: {
  portalSlug: string;
  sessionToken: string;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    return await post<ApiResponse<null>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/sessions/revoke`,
      {},
      {
        credentials: "omit",
        headers: { Authorization: `FeedbackSession ${input.sessionToken}` },
      },
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const markWidgetFeedbackUpdatesSeenAction = async (input: {
  portalSlug: string;
  sessionToken: string;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    return await post<
      ApiResponse<{ lastSeenAt: string; unreadUpdateCount: 0 }>
    >(
      `portals/${encodeURIComponent(portalSlug)}/feedback/updates/seen`,
      {},
      {
        credentials: "omit",
        headers: { Authorization: `FeedbackSession ${input.sessionToken}` },
      },
    );
  } catch (error) {
    return getApiError(error);
  }
};
