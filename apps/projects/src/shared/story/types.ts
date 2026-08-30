import type { UserSummary } from "@/types/user-summary";

export type StoryPriority =
  | "No Priority"
  | "Urgent"
  | "High"
  | "Medium"
  | "Low";

export type AutoSchedulingStatus =
  | "off"
  | "needs_owner"
  | "needs_time"
  | "planning"
  | "scheduled"
  | "at_risk"
  | "cannot_fit"
  | "locked";

export type StoryTeamSummary = {
  id: string;
  name: string;
  code: string;
};

export type StoryObjectiveSummary = {
  id: string;
  name: string;
  description: string | null;
};

export type StorySprintSummary = {
  id: string;
  name: string;
  goal: string | null;
  startDate: string;
  endDate: string;
};

/**
 * Stable story-resource data shared by collection and detail surfaces.
 * Collection query envelopes remain owned by the Stories feature.
 */
export type Story = {
  id: string;
  title: string;
  estimateLabel: string | null;
  estimateValue: number | null;
  estimateScheme: "points" | "tshirt";
  estimatedDurationMinutes: number | null;
  minimumFocusBlockMinutes: number | null;
  autoSchedulingEnabled: boolean;
  autoSchedulingLocked: boolean;
  autoSchedulingStatus: AutoSchedulingStatus;
  autoSchedulingReason: string | null;
  autoSchedulingUpdatedAt: string | null;
  description?: string;
  statusId: string;
  sprintId: string | null;
  sprint?: StorySprintSummary | null;
  objectiveId: string | null;
  objective?: StoryObjectiveSummary | null;
  keyResultId: string | null;
  teamId: string;
  team?: StoryTeamSummary | null;
  workspaceId: string;
  assigneeId: string | null;
  assignee?: UserSummary | null;
  collaboratorCount: number;
  reporterId: string;
  reporter?: UserSummary | null;
  epicId: string | null;
  sequenceId: number;
  priority: StoryPriority;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
  deletedAt: string | null;
  archivedAt: string | null;
  labels: string[] | null;
  subStories: Story[];
};

export type StoryActivity = {
  id: string;
  storyId: string;
  userId: string;
  user: UserSummary;
  type: "update" | "create" | "link";
  field: string;
  currentValue: string;
  oldValue?: unknown;
  newValue?: unknown;
  reason?: string | null;
  createdAt: string;
};

export type StoryAssociationType = "related" | "blocking" | "duplicate";

export type StoryAssociation = {
  id: string;
  fromStoryId: string;
  toStoryId: string;
  type: StoryAssociationType;
  story: Story;
};

export type DetailedStory = {
  id: string;
  sequenceId: number;
  title: string;
  estimateLabel: string | null;
  estimateValue: number | null;
  estimateScheme: "points" | "tshirt";
  estimatedDurationMinutes: number | null;
  minimumFocusBlockMinutes: number | null;
  autoSchedulingEnabled: boolean;
  autoSchedulingLocked: boolean;
  autoSchedulingStatus: AutoSchedulingStatus;
  autoSchedulingReason: string | null;
  autoSchedulingUpdatedAt: string | null;
  description: string;
  descriptionHTML: string;
  parentId: string;
  teamId: string;
  teamCode: string;
  workspaceId: string;
  objectiveId: string | null;
  keyResultId: string | null;
  statusId: string;
  assigneeId: string | null;
  assignee?: UserSummary | null;
  collaboratorIds: string[];
  collaborators: UserSummary[];
  collaboratorCount: number;
  watcherCount: number;
  watchers: UserSummary[];
  isWatching: boolean;
  watchingReason: "assignee" | "collaborator" | "watcher" | null;
  reporterId: string;
  reporter?: UserSummary | null;
  priority: StoryPriority;
  sprintId: string | null;
  epicId: string | null;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
  completedAt: string | null;
  archivedAt: string | null;
  subStories: Story[];
  labels: string[] | null;
  associations: StoryAssociation[];
};

export type NewStory = Partial<DetailedStory> & {
  /** Stable client operation key used to make retried creates idempotent. */
  idempotencyKey?: string;
  labelIds?: string[];
};

export type StoryUpdate = Partial<DetailedStory> & {
  reconcileDescriptionMedia?: boolean;
};

export type StoryAttachment = {
  id: string;
  filename: string;
  size: number;
  mimeType: string;
  url: string;
  createdAt: string;
  uploadedBy: string;
};
