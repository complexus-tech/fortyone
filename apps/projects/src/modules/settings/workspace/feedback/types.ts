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

export type FeedbackPortal = {
  id: string;
  workspaceId: string;
  name: string;
  slug: string;
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
  boards: FeedbackBoard[];
};

export type FeedbackBoardWithTeam = FeedbackBoard & {
  team?: Team;
};

export type UpdateFeedbackPortalInput = {
  isPublic: boolean;
};

export type CreateFeedbackBoardInput = {
  portalId: string;
  teamId: string;
  name: string;
  color: string;
};
