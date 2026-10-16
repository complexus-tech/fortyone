import type { UserSummary } from "./user-summary";

export type Comment = {
  id: string;
  storyId: string;
  parentId: string | null;
  userId: string;
  user: UserSummary;
  comment: string;
  createdAt: string;
  updatedAt: string;
  subComments: Comment[];
  desktopPending?: boolean;
};
