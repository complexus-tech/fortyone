import type { Team } from "@/modules/teams/types";

export type FeedbackBoard = {
  id: string;
  workspaceId: string;
  portalId: string;
  teamId: string;
  name: string;
  slug: string;
  color: string;
  orderIndex: number;
  createdAt: string;
  updatedAt: string;
};

export type FeedbackParticipationMode =
  | "account_required"
  | "verified_guest"
  | "anonymous_allowed";

export type FeedbackGuestIdentityPolicy =
  | "show_identity"
  | "allow_public_masking"
  | "always_mask_guests";

export type FeedbackPortal = {
  id: string;
  workspaceId: string;
  name: string;
  slug: string;
  isPublic: boolean;
  participationMode: FeedbackParticipationMode;
  guestIdentityPolicy: FeedbackGuestIdentityPolicy;
  createdAt: string;
  updatedAt: string;
  boards?: FeedbackBoard[];
};

export type FeedbackBoardWithTeam = FeedbackBoard & {
  team?: Team;
};

export type UpdateFeedbackPortalInput = {
  guestIdentityPolicy?: FeedbackGuestIdentityPolicy;
  isPublic?: boolean;
  participationMode?: FeedbackParticipationMode;
};

export type CreateFeedbackBoardInput = {
  portalId: string;
  teamId: string;
  name: string;
  color: string;
};

export type FeedbackReviewerEmailFrequency = "off" | "daily" | "weekly";

export type FeedbackReviewer = {
  userId: string;
  name: string;
  email: string;
  avatarUrl?: string | null;
  role: "admin" | "member";
  emailFrequency: FeedbackReviewerEmailFrequency;
};

export type UpdateFeedbackReviewerInput = {
  emailFrequency: FeedbackReviewerEmailFrequency;
};

export type FeedbackWidgetSettings = {
  allowedOrigins: string[];
  enabled: boolean;
  hasSigningSecret: boolean;
  previousVersionExpiresAt?: string | null;
  signingSecretVersion: number;
  widgetKeyId: string;
};

export type UpdateFeedbackWidgetSettingsInput = Pick<
  FeedbackWidgetSettings,
  "allowedOrigins" | "enabled"
>;

export type FeedbackWidgetSigningSecret = FeedbackWidgetSettings & {
  signingSecret: string;
};

export type FeedbackUpdateLinkedItem = {
  id: string;
  slug: string;
  status: string;
  title: string;
};

export type FeedbackItemCandidate = FeedbackUpdateLinkedItem & {
  commentCount: number;
  voteCount: number;
};

export type FeedbackUpdate = {
  body: string;
  coverImageUrl?: string | null;
  createdAt?: string;
  id: string;
  linkedItems: FeedbackUpdateLinkedItem[];
  portalId: string;
  publishedAt?: string | null;
  publishedByUserId?: string | null;
  slug: string;
  status: "draft" | "published";
  summary?: string | null;
  title: string;
  updatedAt?: string;
  workspaceId: string;
};

export type UpsertFeedbackUpdateInput = {
  body: string;
  coverImageUrl?: string;
  itemIds: string[];
  portalId: string;
  summary?: string;
  title: string;
};
