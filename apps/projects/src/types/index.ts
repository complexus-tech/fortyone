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
  avatarUrl: string | null;
  isActive: boolean;
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

export type Link = {
  id: string;
  url: string;
  title: string;
  storyId: string;
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

export type StoriesSummary = {
  closed: number;
  overdue: number;
  inProgress: number;
  created: number;
  assigned: number;
};

export type StatusSummary = {
  name: string;
  count: number;
};

export type PrioritySummary = {
  priority: string;
  count: number;
};

export type Contribution = {
  date: string;
  contributions: number;
};

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  color: string;
  avatarUrl: string | null;
  userRole: UserRole;
  trialEndsOn: string | null;
  deletedAt: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type WorkspaceSettings = {
  storyTerm: "issue" | "task" | "story";
  sprintTerm: "sprint" | "cycle" | "iteration";
  objectiveTerm: "objective" | "goal" | "project";
  keyResultTerm: "key result" | "milestone" | "focus area";
  objectiveEnabled: boolean;
  keyResultEnabled: boolean;
};

export type AutomationPreferences = {
  id: string;
  autoAssignSelf: boolean;
  assignSelfOnBranchCopy: boolean;
  moveStoryToStartedOnBranch: boolean;
  openStoryInDialog: boolean;
};

export type UpdateAutomationPreferences = Partial<
  Omit<AutomationPreferences, "id">
>;

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

export type Invoice = {
  invoiceId: number;
  stripeInvoiceId: string;
  amountPaid: number;
  invoiceDate: string;
  seatsCount: number;
  customerName: string;
  hostedUrl: string;
  createdAt: string;
};

export * from "./tts";
