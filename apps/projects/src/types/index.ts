import type { UserSummary } from "./user-summary";

export type { ApiResponse } from "./api-response";
export type { User } from "./user";
export type { UserSummary } from "./user-summary";
export type { UserRole } from "./user-role";
export type { Member, MembersPage } from "./member";
export type { Label, LabelsPage } from "./label";
export type { Link } from "./link";
export type { Workspace } from "./workspace";
export type {
  OnboardingTourStatus,
  OnboardingTourProgress,
  UpdateOnboardingTourProgress,
} from "@/shared/walkthrough/progress/types";

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

export type WorkspaceSettings = {
  storyTerm: "issue" | "task" | "story";
  sprintTerm: "sprint" | "cycle" | "iteration";
  objectiveTerm: "objective" | "goal" | "project";
  keyResultTerm: "key result" | "milestone" | "focus area";
  objectiveEnabled: boolean;
  keyResultEnabled: boolean;
  workingDays: number[];
  workingStartMinute: number;
  workingEndMinute: number;
};

export type AutomationPreferences = {
  id: string;
  autoAssignSelf: boolean;
  autoScheduling: boolean;
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
