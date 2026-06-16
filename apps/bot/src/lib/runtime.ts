export interface RuntimeOption {
  label: string;
  value: string;
}

export interface SlackActor {
  channelId?: string;
  channelName?: string;
  messageTs?: string;
  teamId?: string;
  threadTs?: string;
  userId: string;
  userName?: string;
}

export interface CreateStoryFromSlackInput {
  assigneeId?: string;
  description?: string;
  labelIds?: string[];
  objectiveId?: string;
  priority: string;
  source: SlackActor & {
    messageText?: string;
  };
  statusId?: string;
  teamId: string;
  title: string;
}

export interface CreatedStory {
  id: string;
  ref: string;
  title: string;
  url?: string;
}

export interface StoryFormOptions {
  labels: Record<string, RuntimeOption[]> | RuntimeOption[];
  members: Record<string, RuntimeOption[]> | RuntimeOption[];
  objectives: Record<string, RuntimeOption[]> | RuntimeOption[];
  statuses: Record<string, RuntimeOption[]> | RuntimeOption[];
  teams: RuntimeOption[];
}

export interface SlackInstallation {
  botToken: string;
  botUserId?: string;
  teamName?: string;
}

export interface SlackIdentity {
  connectUrl?: string;
  userId?: string;
  workspaceId: string;
  workspaceSlug: string;
}

export interface RuntimeLogInput {
  actor?: SlackActor;
  endpoint: string;
  errorMessage?: string;
  outcome: string;
  requestType: string;
  responseCode: number;
}

export interface StoryUnfurl {
  assignee?: string;
  priority?: string;
  status?: string;
  title: string;
  url: string;
}

export interface StoryRuntime {
  createStoryFromSlackForm: (
    input: CreateStoryFromSlackInput,
  ) => CreatedStory | Promise<CreatedStory>;
  getSlackInstallation?: (teamId: string) => Promise<SlackInstallation | null>;
  getStoryUnfurl?: (
    url: string,
    actor: SlackActor,
  ) => Promise<StoryUnfurl | null>;
  listStoryOptions: (
    actor: SlackActor,
  ) => StoryFormOptions | Promise<StoryFormOptions>;
  recordSlackRuntimeLog?: (input: RuntimeLogInput) => Promise<void>;
  recordSlackThreadComment?: (input: {
    actor: SlackActor;
    messageText: string;
    storyId: string;
  }) => Promise<void>;
  resolveSlackIdentity: (actor: SlackActor) => Promise<SlackIdentity>;
  searchLabels: (
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ) => RuntimeOption[] | Promise<RuntimeOption[]>;
  searchMembers: (
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ) => RuntimeOption[] | Promise<RuntimeOption[]>;
  searchObjectives: (
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ) => RuntimeOption[] | Promise<RuntimeOption[]>;
  searchStatuses: (
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ) => RuntimeOption[] | Promise<RuntimeOption[]>;
  searchTeams: (
    actor: SlackActor,
    query: string,
  ) => RuntimeOption[] | Promise<RuntimeOption[]>;
}
