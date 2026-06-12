import type { BotConfig } from "@/lib/config";

export type RuntimeOption = {
  label: string;
  value: string;
};

export type SlackActor = {
  channelId?: string;
  channelName?: string;
  messageTs?: string;
  teamId?: string;
  threadTs?: string;
  userId: string;
  userName?: string;
};

export type CreateStoryFromSlackInput = {
  assigneeId?: string;
  description?: string;
  objectiveId?: string;
  priority: string;
  source: SlackActor & {
    messageText?: string;
  };
  statusId?: string;
  teamId: string;
  title: string;
};

export type CreatedStory = {
  id: string;
  ref: string;
  title: string;
  url: string;
};

export type StoryUnfurl = {
  assignee?: string;
  priority?: string;
  status?: string;
  title: string;
  url: string;
};

export type NotificationDelivery = {
  channelId: string;
  text: string;
  threadTs?: string;
};

type RequestOptions = {
  body?: unknown;
  method?: "GET" | "POST";
  query?: Record<string, string | undefined>;
};

export class FortyOneClient {
  private readonly apiUrl: string;
  private readonly serviceToken: string;

  constructor(config: BotConfig) {
    this.apiUrl = config.fortyOneApiUrl.replace(/\/+$/, "");
    this.serviceToken = config.fortyOneServiceToken;
  }

  async searchTeams(actor: SlackActor, query: string): Promise<RuntimeOption[]> {
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

  async createStoryFromSlackForm(
    input: CreateStoryFromSlackInput,
  ): Promise<CreatedStory> {
    return this.request("/internal/bot/slack/stories", {
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

  async getStoryUnfurl(url: string, actor: SlackActor): Promise<StoryUnfurl | null> {
    return this.request("/internal/bot/slack/unfurls/story", {
      method: "POST",
      body: { actor, url },
    });
  }

  async listMentionNotifications(actor: SlackActor): Promise<NotificationDelivery[]> {
    return this.request("/internal/bot/slack/notifications/mentions", {
      method: "POST",
      body: { actor },
    });
  }

  private async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    if (!this.serviceToken) {
      throw new Error("FORTYONE_BOT_TOKEN is required for FortyOne bot API calls.");
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

    return response.json() as Promise<T>;
  }
}
