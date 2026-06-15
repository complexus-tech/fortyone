import type { BotConfig } from "@/lib/config";

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
  url: string;
}

export interface StoryUnfurl {
  assignee?: string;
  priority?: string;
  status?: string;
  title: string;
  url: string;
}

export interface NotificationDelivery {
  channelId: string;
  text: string;
  threadTs?: string;
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
  action: string;
  actor?: SlackActor;
  error?: string;
  metadata?: Record<string, unknown>;
}

interface RequestOptions {
  body?: unknown;
  method?: "GET" | "POST";
  query?: Record<string, string | undefined>;
}

interface ApiEnvelope<T> {
  data: T;
  error?: {
    message?: string;
  } | null;
}

const isApiEnvelope = <T>(value: unknown): value is ApiEnvelope<T> =>
  value !== null &&
  typeof value === "object" &&
  ("data" in value || "error" in value);

export class FortyOneClient {
  private readonly apiUrl: string;
  private readonly serviceToken: string;

  constructor(config: BotConfig) {
    this.apiUrl = config.fortyOneApiUrl.replace(/\/+$/, "");
    this.serviceToken = config.fortyOneServiceToken;
  }

  async searchTeams(
    actor: SlackActor,
    query: string,
  ): Promise<RuntimeOption[]> {
    return this.request("/internal/bot/slack/options/teams", {
      method: "POST",
      body: { actor, query },
    });
  }

  async searchStatuses(
    actor: SlackActor,
    teamId: string,
    query: string,
  ): Promise<RuntimeOption[]> {
    return this.request("/internal/bot/slack/options/statuses", {
      method: "POST",
      body: { actor, query, teamId },
    });
  }

  async searchMembers(
    actor: SlackActor,
    teamId: string,
    query: string,
  ): Promise<RuntimeOption[]> {
    return this.request("/internal/bot/slack/options/members", {
      method: "POST",
      body: { actor, query, teamId },
    });
  }

  async searchObjectives(
    actor: SlackActor,
    teamId: string,
    query: string,
  ): Promise<RuntimeOption[]> {
    return this.request("/internal/bot/slack/options/objectives", {
      method: "POST",
      body: { actor, query, teamId },
    });
  }

  async searchLabels(
    actor: SlackActor,
    teamId: string,
    query: string,
  ): Promise<RuntimeOption[]> {
    return this.request("/internal/bot/slack/options/labels", {
      method: "POST",
      body: { actor, query, teamId },
    });
  }

  async createStoryFromSlackForm(
    input: CreateStoryFromSlackInput,
  ): Promise<CreatedStory> {
    return this.request("/internal/bot/slack/stories", {
      method: "POST",
      body: input,
    });
  }

  async getSlackInstallation(
    teamId: string,
  ): Promise<SlackInstallation | null> {
    try {
      return await this.request(
        `/internal/bot/slack/installations/${encodeURIComponent(teamId)}`,
      );
    } catch (error) {
      if (
        error instanceof Error &&
        error.message.includes("FortyOne bot API request failed: 404")
      ) {
        return null;
      }
      throw error;
    }
  }

  async resolveSlackIdentity(actor: SlackActor): Promise<SlackIdentity> {
    return this.request("/internal/bot/slack/identity/resolve", {
      method: "POST",
      body: { actor },
    });
  }

  async recordSlackRuntimeLog(input: RuntimeLogInput): Promise<void> {
    await this.request("/internal/bot/slack/logs", {
      method: "POST",
      body: input,
    });
  }

  async recordSlackThreadComment(input: {
    actor: SlackActor;
    messageText: string;
    storyId: string;
  }): Promise<void> {
    await this.request("/internal/bot/slack/thread-comments", {
      method: "POST",
      body: input,
    });
  }

  async getStoryUnfurl(
    url: string,
    actor: SlackActor,
  ): Promise<StoryUnfurl | null> {
    return this.request("/internal/bot/slack/unfurls/story", {
      method: "POST",
      body: { actor, url },
    });
  }

  async listMentionNotifications(
    actor: SlackActor,
  ): Promise<NotificationDelivery[]> {
    return this.request("/internal/bot/slack/notifications/mentions", {
      method: "POST",
      body: { actor },
    });
  }

  private async request<T>(
    path: string,
    options: RequestOptions = {},
  ): Promise<T> {
    if (!this.serviceToken) {
      throw new Error(
        "FORTYONE_BOT_TOKEN is required for FortyOne bot API calls.",
      );
    }

    const url = new URL(`${this.apiUrl}${path}`);
    Object.entries(options.query ?? {}).forEach(([key, value]) => {
      if (value) {
        url.searchParams.set(key, value);
      }
    });

    const response = await fetch(url, {
      method: options.method ?? "GET",
      headers: {
        Authorization: `Bearer ${this.serviceToken}`,
        "Content-Type": "application/json",
      },
      body: options.body ? JSON.stringify(options.body) : undefined,
    });

    if (!response.ok) {
      const text = await response.text();
      throw new Error(
        `FortyOne bot API request failed: ${response.status} ${text}`,
      );
    }

    if (response.status === 204) {
      return undefined as T;
    }

    const json = (await response.json()) as unknown;
    if (!isApiEnvelope<T>(json)) {
      return json as T;
    }
    if (json.error) {
      throw new Error(json.error.message || "FortyOne bot API request failed.");
    }
    return json.data;
  }
}
