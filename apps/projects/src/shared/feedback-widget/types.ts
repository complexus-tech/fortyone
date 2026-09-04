/**
 * Public feedback data shared by the hosted portal and the embedded widget.
 * Feature-owned page/query envelopes remain in their respective modules.
 */
export type FeedbackParticipationMode =
  | "account_required"
  | "verified_guest"
  | "anonymous_allowed";

export type FeedbackGuestIdentityPolicy =
  | "show_identity"
  | "allow_public_masking"
  | "always_mask_guests";

export type PublicParticipantKind =
  | "account"
  | "verified_guest"
  | "external"
  | "anonymous";

export type PublicRequestStatus =
  | "pending"
  | "reviewing"
  | "planned"
  | "in_progress"
  | "completed"
  | "closed";

export type PublicFeedbackListStatus = "active" | PublicRequestStatus;

export type PublicPortalSort = "top" | "newest" | "oldest";

export type PublicRequestBoard = {
  id: string;
  teamId?: string;
  name: string;
  slug?: string;
  color?: string;
};

export type PublicRequestComment = {
  id: string;
  parentId?: string | null;
  participantKind?: PublicParticipantKind;
  authorMasked?: boolean;
  authorName: string;
  authorAvatar?: string | null;
  body: string;
  createdAtLabel: string;
};

export type PublicFeedbackStoryLink = {
  id: string;
  storyId: string;
  relationship: "created_from" | "linked" | "solves";
};

export type PublicRequestAttachment = {
  id: string;
  filename: string;
  size: number;
  mimeType: string;
  url: string;
};

export type PublicRequest = {
  id: string;
  authorId: string | null;
  slug: string;
  title: string;
  description: string;
  descriptionHTML?: string;
  authorMasked?: boolean;
  authorName: string;
  authorAvatar?: string | null;
  boardId: string;
  status: PublicRequestStatus;
  voteCount: number;
  commentCount: number;
  createdAtLabel: string;
  roadmapSummary?: string;
  comments: PublicRequestComment[];
  storyLinks: PublicFeedbackStoryLink[];
  attachments?: PublicRequestAttachment[];
  participantKind?: PublicParticipantKind;
  following?: boolean;
  viewerVote?: -1 | 0 | 1;
};

export type PublicPortalUpdate = {
  id: string;
  slug: string;
  title: string;
  summary?: string | null;
  body: string;
  coverImageUrl?: string | null;
  publishedAt: string;
  publishedAtLabel: string;
  linkedItems: {
    id: string;
    slug: string;
    title: string;
    status: PublicRequestStatus;
  }[];
};

export type PublicPortalWorkspace = {
  name: string;
  slug: string;
  avatarUrl: string | null;
  color: string;
};

export type PublicPortalViewer = {
  kind: "account";
  id: string;
  name: string;
  email: string;
  avatarUrl: string | null;
  appHref?: string;
  accountHref: string;
  feedbackSetupHref: string;
  canReceiveUpdates: true;
};

export type PublicPortalGuestParticipant = {
  kind: "verified_guest" | "external";
  id: string;
  name: string;
  displayName: string;
  email?: string;
  avatarUrl: string | null;
  masked: boolean;
  canReceiveUpdates: true;
  sessionExpiresAt: string;
  unreadUpdateCount: number;
};

export type PublicPortalAnonymousParticipant = {
  kind: "anonymous";
  canReceiveUpdates: false;
};

export type PublicPortalParticipant =
  | PublicPortalViewer
  | PublicPortalGuestParticipant
  | PublicPortalAnonymousParticipant;

export type LegacyPublicPortalViewer = Omit<
  PublicPortalViewer,
  "kind" | "canReceiveUpdates"
>;

export type PublicPortal = {
  id: string;
  name: string;
  slug: string;
  participationMode: FeedbackParticipationMode;
  guestIdentityPolicy: FeedbackGuestIdentityPolicy;
  hasPublishedUpdates: boolean;
  workspace: PublicPortalWorkspace;
  boards: PublicRequestBoard[];
  requests: PublicRequest[];
  requestsHasMore: boolean;
  updates: PublicPortalUpdate[];
};
