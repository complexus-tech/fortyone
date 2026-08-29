import type {
  PublicPortalAnonymousParticipant,
  PublicPortalGuestParticipant,
  PublicPortalParticipant,
  PublicPortalViewer,
} from "./types";

export const anonymousPublicPortalParticipant = {
  canReceiveUpdates: false,
  kind: "anonymous",
} as const satisfies PublicPortalAnonymousParticipant;

export const isAccountParticipant = (
  participant: PublicPortalParticipant,
): participant is PublicPortalViewer => participant.kind === "account";

export const isGuestParticipant = (
  participant: PublicPortalParticipant,
): participant is PublicPortalGuestParticipant =>
  participant.kind === "verified_guest" || participant.kind === "external";

export const isContactableParticipant = (
  participant: PublicPortalParticipant,
): participant is PublicPortalViewer | PublicPortalGuestParticipant =>
  participant.kind !== "anonymous";

export const canVerifyAsGuest = (
  participationMode:
    | "account_required"
    | "verified_guest"
    | "anonymous_allowed",
) => participationMode !== "account_required";
