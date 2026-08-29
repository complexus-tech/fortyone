"use server";

import { revalidatePath } from "next/cache";
import { get, post, put, remove } from "api-client";
import { auth } from "@/auth";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { toPublicRequest, type ApiFeedbackItem } from "./data";
import {
  createFeedbackIngressHeaders,
  normalizeFeedbackPortalSlug,
} from "./feedback-ingress";
import {
  clearFeedbackSessionToken,
  getFeedbackPreferenceAuthorization,
  getFeedbackSessionAuthorization,
  setFeedbackPreferenceSessionToken,
  setFeedbackSessionToken,
} from "./guest-session";
import type {
  PublicParticipantKind,
  PublicPortalGuestParticipant,
} from "./types";

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
  participantKind: PublicParticipantKind;
  vote: -1 | 1;
};

type CommentInput = ItemInput & {
  body: string;
  parentId?: string;
  participantKind: PublicParticipantKind;
};

export type FeedbackVoteResult = {
  participantKind?: PublicParticipantKind;
  vote: -1 | 0 | 1;
  voteCount: number;
};

export type CreatedFeedbackComment = {
  id: string;
  parentId?: string | null;
  authorMasked?: boolean;
  authorName: string;
  authorAvatar?: string | null;
  body: string;
  createdAt: string;
  participantKind?: PublicParticipantKind;
};

export type CreateFeedbackResult = {
  kind: PublicParticipantKind;
  request: ReturnType<typeof toPublicRequest>;
  following?: boolean;
};

type ApiFeedbackParticipant = {
  id: string;
  kind: "verified_guest" | "external";
  displayName: string;
  publicName: string;
  email?: string;
  avatarUrl: string | null;
  masked: boolean;
};

type FeedbackSessionResponse = {
  participant: ApiFeedbackParticipant;
  session: {
    token: string;
    expiresAt: string;
  };
  unreadUpdateCount?: number;
};

type FeedbackSessionStatusResponse = Omit<
  FeedbackSessionResponse,
  "session"
> & {
  session: { expiresAt: string };
};

export type FeedbackVerificationRequest = {
  portalSlug: string;
  email: string;
  displayName?: string;
  hideNamePublicly?: boolean;
};

export type FeedbackVerificationConfirmation = {
  portalSlug: string;
  email?: string;
  code?: string;
  token?: string;
};

export type FeedbackFollowState = {
  following: boolean;
};

export type FeedbackPreferenceItem = {
  itemId: string;
  itemSlug: string;
  title: string;
  following: boolean;
};

export type FeedbackPreferences = {
  portalEmailsEnabled: boolean;
  items: FeedbackPreferenceItem[];
};

const toGuestParticipant = (
  participant: ApiFeedbackParticipant,
  expiresAt: string,
  unreadUpdateCount = 0,
): PublicPortalGuestParticipant => ({
  id: participant.id,
  kind: participant.kind,
  name: participant.publicName,
  displayName: participant.displayName,
  email: participant.email,
  avatarUrl: participant.avatarUrl,
  masked: participant.masked,
  canReceiveUpdates: true,
  sessionExpiresAt: expiresAt,
  unreadUpdateCount,
});

const toCreateFeedbackResult = (
  item: ApiFeedbackItem,
): CreateFeedbackResult => {
  const request = toPublicRequest(item);
  const kind =
    item.participantKind ?? (item.anonymous ? "anonymous" : "account");

  return { kind, request, following: item.following ?? false };
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
  participationIntent: "account" | "verified_guest" | "external" | "anonymous",
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const session = participationIntent === "account" ? await auth() : null;
    if (participationIntent === "account" && !session) {
      return {
        data: null,
        error: {
          message: "Your session expired. Please log in and try again.",
        },
      };
    }
    const isAnonymous = participationIntent === "anonymous";
    const isGuest =
      participationIntent === "verified_guest" ||
      participationIntent === "external";
    const guestAuthorization = isGuest
      ? await getFeedbackSessionAuthorization(portalSlug)
      : null;
    if (isGuest && !guestAuthorization) {
      return {
        data: null,
        error: {
          message: "Your feedback session expired. Verify your email again.",
        },
      };
    }
    const ingressHeaders = isAnonymous
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
      ingressHeaders || guestAuthorization
        ? {
            credentials: "omit",
            headers: {
              ...ingressHeaders,
              ...(guestAuthorization
                ? { Authorization: guestAuthorization }
                : {}),
            },
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

export const createVerifiedGuestFeedbackAction = async (
  input: CreateFeedbackInput,
) => submitFeedback(input, "verified_guest");

const getParticipantWriteOptions = async (
  portalSlug: string,
  participantKind: PublicParticipantKind,
) => {
  if (participantKind === "account") {
    const session = await auth();
    if (!session) throw new Error("Please log in to continue");
    return undefined;
  }
  if (participantKind === "anonymous") {
    throw new Error("Verify your email or log in to continue");
  }

  const authorization = await getFeedbackSessionAuthorization(portalSlug);
  if (!authorization) {
    throw new Error("Your feedback session expired. Verify your email again.");
  }

  return {
    credentials: "omit" as const,
    headers: { Authorization: authorization },
  };
};

export const toggleFeedbackVoteAction = async (input: VoteInput) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const options = await getParticipantWriteOptions(
      portalSlug,
      input.participantKind,
    );

    const response = await post<ApiResponse<FeedbackVoteResult>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/items/${encodeURIComponent(input.itemId)}/vote`,
      { vote: input.vote },
      options,
    );
    refreshPortal(portalSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const createFeedbackCommentAction = async (input: CommentInput) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const options = await getParticipantWriteOptions(
      portalSlug,
      input.participantKind,
    );

    const response = await post<ApiResponse<CreatedFeedbackComment>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/items/${encodeURIComponent(input.itemId)}/comments`,
      {
        body: input.body,
        ...(input.parentId ? { parentId: input.parentId } : {}),
      },
      options,
    );
    refreshPortal(portalSlug);
    refreshFeedbackItem(portalSlug, input.itemSlug);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const requestFeedbackVerificationAction = async (
  input: FeedbackVerificationRequest,
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const ingressHeaders = await createFeedbackIngressHeaders(portalSlug);
    return await post<ApiResponse<{ accepted: true; expiresAt: string }>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/verifications`,
      {
        email: input.email.trim(),
        ...(input.displayName?.trim()
          ? { displayName: input.displayName.trim() }
          : {}),
        ...(input.hideNamePublicly !== undefined
          ? { hideNamePublicly: input.hideNamePublicly }
          : {}),
        source: "portal",
      },
      { credentials: "omit", headers: ingressHeaders },
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const confirmFeedbackVerificationAction = async (
  input: FeedbackVerificationConfirmation,
): Promise<ApiResponse<{ participant: PublicPortalGuestParticipant }>> => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const ingressHeaders = await createFeedbackIngressHeaders(portalSlug);
    const response = await post<ApiResponse<FeedbackSessionResponse>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/verifications/confirm`,
      {
        ...(input.token ? { token: input.token } : {}),
        ...(input.code ? { code: input.code.replace(/\s/g, "") } : {}),
        ...(input.email ? { email: input.email.trim() } : {}),
        source: "portal",
      },
      { credentials: "omit", headers: ingressHeaders },
    );
    if (!response.data) {
      return { data: null, error: response.error };
    }

    await setFeedbackSessionToken({
      expiresAt: response.data.session.expiresAt,
      portalSlug,
      token: response.data.session.token,
    });

    return {
      data: {
        participant: toGuestParticipant(
          response.data.participant,
          response.data.session.expiresAt,
          response.data.unreadUpdateCount,
        ),
      },
    };
  } catch (error) {
    const response = getApiError(error);
    return { data: null, error: response.error };
  }
};

export const revokeFeedbackSessionAction = async (portalSlugInput: string) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(portalSlugInput);
    const authorization = await getFeedbackSessionAuthorization(portalSlug);
    if (authorization) {
      await post<ApiResponse<null>>(
        `portals/${encodeURIComponent(portalSlug)}/feedback/sessions/revoke`,
        {},
        {
          credentials: "omit",
          headers: { Authorization: authorization },
        },
      );
    }
    await clearFeedbackSessionToken(portalSlug);
    refreshPortal(portalSlug);
    return { data: null };
  } catch (error) {
    return getApiError(error);
  }
};

export const getCurrentFeedbackGuestAction = async (
  portalSlugInput: string,
): Promise<ApiResponse<{ participant: PublicPortalGuestParticipant }>> => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(portalSlugInput);
    const authorization = await getFeedbackSessionAuthorization(portalSlug);
    if (!authorization) return { data: null };

    const response = await get<ApiResponse<FeedbackSessionStatusResponse>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/session`,
      {
        credentials: "omit",
        headers: { Authorization: authorization },
      },
    );
    if (!response.data) return { data: null, error: response.error };

    return {
      data: {
        participant: toGuestParticipant(
          response.data.participant,
          response.data.session.expiresAt,
          response.data.unreadUpdateCount,
        ),
      },
    };
  } catch (error) {
    const response = getApiError(error);
    return { data: null, error: response.error };
  }
};

export const getFeedbackFollowStateAction = async (
  input: ItemInput & { participantKind: PublicParticipantKind },
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const options = await getParticipantWriteOptions(
      portalSlug,
      input.participantKind,
    );

    return await get<ApiResponse<FeedbackFollowState>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/items/${encodeURIComponent(input.itemId)}/follow`,
      options,
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const updateFeedbackFollowAction = async (
  input: ItemInput & {
    following: boolean;
    participantKind: PublicParticipantKind;
  },
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(input.portalSlug);
    const options = await getParticipantWriteOptions(
      portalSlug,
      input.participantKind,
    );
    const path = `portals/${encodeURIComponent(portalSlug)}/feedback/items/${encodeURIComponent(input.itemId)}/follow`;
    const response = input.following
      ? await put<ApiResponse<FeedbackFollowState>>(
          path,
          { notify: true },
          options,
        )
      : await remove<ApiResponse<FeedbackFollowState>>(path, options);
    refreshFeedbackItem(portalSlug, input.itemSlug);
    return response ?? { data: { following: input.following } };
  } catch (error) {
    return getApiError(error);
  }
};

export const exchangeFeedbackPreferenceTokenAction = async ({
  portalSlug: portalSlugInput,
  token,
}: {
  portalSlug: string;
  token: string;
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(portalSlugInput);
    const response = await post<ApiResponse<FeedbackSessionResponse>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/preferences/exchange`,
      { token },
      { credentials: "omit" },
    );
    if (!response.data) return response;
    await setFeedbackPreferenceSessionToken({
      expiresAt: response.data.session.expiresAt,
      portalSlug,
      token: response.data.session.token,
    });
    return { data: { exchanged: true } };
  } catch (error) {
    return getApiError(error);
  }
};

export const getFeedbackPreferencesAction = async (portalSlugInput: string) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(portalSlugInput);
    const authorization = await getFeedbackPreferenceAuthorization(portalSlug);
    if (!authorization) {
      return {
        data: null,
        error: { message: "This preference link has expired." },
      };
    }
    return await get<ApiResponse<FeedbackPreferences>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/preferences`,
      {
        credentials: "omit",
        headers: { Authorization: authorization },
      },
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const updateFeedbackPreferencesAction = async ({
  portalEmailsEnabled,
  portalSlug: portalSlugInput,
  items,
}: {
  portalSlug: string;
  portalEmailsEnabled?: boolean;
  items?: { itemId: string; following: boolean }[];
}) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(portalSlugInput);
    const authorization = await getFeedbackPreferenceAuthorization(portalSlug);
    if (!authorization) {
      return {
        data: null,
        error: { message: "This preference session has expired." },
      };
    }
    return await put<ApiResponse<FeedbackPreferences>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/preferences`,
      {
        ...(portalEmailsEnabled !== undefined ? { portalEmailsEnabled } : {}),
        ...(items ? { items } : {}),
      },
      {
        credentials: "omit",
        headers: { Authorization: authorization },
      },
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const markFeedbackUpdatesSeenAction = async (
  portalSlugInput: string,
) => {
  try {
    const portalSlug = normalizeFeedbackPortalSlug(portalSlugInput);
    const authorization = await getFeedbackSessionAuthorization(portalSlug);
    if (!authorization) return { data: null };
    return await post<
      ApiResponse<{ unreadUpdateCount: 0; lastSeenAt: string }>
    >(
      `portals/${encodeURIComponent(portalSlug)}/feedback/updates/seen`,
      {},
      {
        credentials: "omit",
        headers: { Authorization: authorization },
      },
    );
  } catch (error) {
    return getApiError(error);
  }
};
