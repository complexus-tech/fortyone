export type ApiResponse<T> = {
  data?: T | null;
  error?: {
    message: string;
  };
};

export type UserRole = "admin" | "member" | "guest" | "system";

export type User = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string | null;
  isActive: boolean;
  lastUsedWorkspaceId: string;
  hasSeenWalkthrough: boolean;
  timezone: string;
  createdAt: string;
  updatedAt: string;
};

export type Member = {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  fullName: string;
  avatarUrl: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type Comment = {
  id: string;
  storyId: string;
  parentId: string | null;
  userId: string;
  comment: string;
  createdAt: string;
  updatedAt: string;
  subComments: Comment[];
};

export type Subscription = {
  workspaceId: string;
  stripeCustomerId: string;
  stripeSubscriptionId: string;
  status:
    | "active"
    | "incomplete"
    | "incomplete_expired"
    | "trialing"
    | "past_due"
    | "unpaid"
    | "canceled"
    | "paused";
  tier: "free" | "pro" | "business" | "enterprise";
  seatCount: number;
  billingInterval: "month" | "year" | "week" | "day";
  billingEndsAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type Label = {
  id: string;
  name: string;
  color: string;
  teamId: string | null;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};
