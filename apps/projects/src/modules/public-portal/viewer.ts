import { get } from "api-client";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getRedirectUrl } from "@/utils";
import { getFeedbackSetupHref } from "./feedback-setup";
import { getFeedbackSessionAuthorization } from "./guest-session";
import { anonymousPublicPortalParticipant } from "./participant";
import { getPortalPathBySlug } from "./utils";
import type {
  PublicPortalGuestParticipant,
  PublicPortalParticipant,
  PublicPortalViewer,
} from "./types";

type ApiResponse<T> = { data: T };

type ApiFeedbackSession = {
  participant: {
    id: string;
    kind: "verified_guest" | "external";
    displayName: string;
    publicName: string;
    email?: string;
    avatarUrl: string | null;
    masked: boolean;
  };
  session: { expiresAt: string };
  unreadUpdateCount?: number;
};

const getAccountParticipant = async (
  portalSlug: string,
  session: NonNullable<Awaited<ReturnType<typeof auth>>>,
): Promise<PublicPortalViewer> => {
  const workspaces = await getWorkspaces();
  const activeWorkspace =
    workspaces.find(
      (workspace) => workspace.id === session.user.lastUsedWorkspaceId,
    ) ?? workspaces.at(0);

  return {
    kind: "account",
    id: session.user.id,
    name: session.user.fullName || session.user.username || session.user.name,
    email: session.user.email,
    avatarUrl: session.user.image,
    appHref: activeWorkspace
      ? getRedirectUrl(workspaces, [], session.user.lastUsedWorkspaceId)
      : undefined,
    accountHref: getPortalPathBySlug(portalSlug, "account"),
    feedbackSetupHref: getFeedbackSetupHref(
      workspaces,
      session.user.lastUsedWorkspaceId,
    ),
    canReceiveUpdates: true,
  };
};

export const getPublicPortalViewer = async (
  portalSlug: string,
): Promise<PublicPortalViewer | null> => {
  const session = await auth();

  if (!session) {
    return null;
  }

  return getAccountParticipant(portalSlug, session);
};

export const getPublicPortalParticipant = async (
  portalSlug: string,
): Promise<PublicPortalParticipant> => {
  const accountSession = await auth();
  if (accountSession) {
    return getAccountParticipant(portalSlug, accountSession);
  }

  const authorization = await getFeedbackSessionAuthorization(portalSlug);
  if (!authorization) return anonymousPublicPortalParticipant;

  try {
    const response = await get<ApiResponse<ApiFeedbackSession>>(
      `portals/${encodeURIComponent(portalSlug)}/feedback/session`,
      {
        credentials: "omit",
        headers: { Authorization: authorization },
      },
    );
    const { participant, session, unreadUpdateCount = 0 } = response.data;
    return {
      kind: participant.kind,
      id: participant.id,
      name: participant.publicName,
      displayName: participant.displayName,
      email: participant.email,
      avatarUrl: participant.avatarUrl,
      masked: participant.masked,
      canReceiveUpdates: true,
      sessionExpiresAt: session.expiresAt,
      unreadUpdateCount,
    } satisfies PublicPortalGuestParticipant;
  } catch {
    return anonymousPublicPortalParticipant;
  }
};
