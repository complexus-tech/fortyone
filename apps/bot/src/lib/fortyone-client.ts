import type { BotConfig } from "@/lib/config";
import type {
  CreatedStory,
  CreateStoryFromSlackInput,
  RuntimeLogInput,
  RuntimeOption,
  SlackActor,
  SlackIdentity,
  SlackInstallation,
  StoryFormOptions,
  StoryRuntime,
  StoryUnfurl,
} from "@/lib/runtime";

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

export class FortyOneClient implements StoryRuntime {
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
      body: { actor, query },
      method: "POST",
    });
  }

  async searchStatuses(
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): Promise<RuntimeOption[]> {
    if (!teamId) {
      return [];
    }
    return this.request("/internal/bot/slack/options/statuses", {
      body: { actor, query, teamId },
      method: "POST",
    });
  }

  async searchMembers(
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): Promise<RuntimeOption[]> {
    if (!teamId) {
      return [];
    }
    return this.request("/internal/bot/slack/options/members", {
      body: { actor, query, teamId },
      method: "POST",
    });
  }

  async searchObjectives(
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): Promise<RuntimeOption[]> {
    if (!teamId) {
      return [];
    }
    return this.request("/internal/bot/slack/options/objectives", {
      body: { actor, query, teamId },
      method: "POST",
    });
  }

  async searchLabels(
    actor: SlackActor,
    teamId: string | undefined,
    query: string,
  ): Promise<RuntimeOption[]> {
    if (!teamId) {
      return [];
    }
    return this.request("/internal/bot/slack/options/labels", {
      body: { actor, query, teamId },
      method: "POST",
    });
  }

  async listStoryOptions(actor: SlackActor): Promise<StoryFormOptions> {
    const teams = await this.searchTeams(actor, "");
    const defaultTeamId = teams[0]?.value;

    const [statuses, members, objectives, labels] = await Promise.all([
      this.searchStatuses(actor, defaultTeamId, ""),
      this.searchMembers(actor, defaultTeamId, ""),
      this.searchObjectives(actor, defaultTeamId, ""),
      this.searchLabels(actor, defaultTeamId, ""),
    ]);

    return {
      labels,
      members,
      objectives,
      statuses,
      teams,
    };
  }

  async createStoryFromSlackForm(
    input: CreateStoryFromSlackInput,
  ): Promise<CreatedStory> {
    return this.request("/internal/bot/slack/stories", {
      body: input,
      method: "POST",
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
      body: { actor },
      method: "POST",
    });
  }

  async recordSlackRuntimeLog(input: RuntimeLogInput): Promise<void> {
    await this.request("/internal/bot/slack/logs", {
      body: {
        actor: input.actor,
        endpoint: input.endpoint,
        errorMessage: input.errorMessage ?? "",
        outcome: input.outcome,
        requestType: input.requestType,
        responseCode: input.responseCode,
      },
      method: "POST",
    });
  }

  async recordSlackThreadComment(input: {
    actor: SlackActor;
    messageText: string;
    storyId: string;
  }): Promise<void> {
    await this.request("/internal/bot/slack/thread-comments", {
      body: input,
      method: "POST",
    });
  }

  async getStoryUnfurl(
    url: string,
    actor: SlackActor,
  ): Promise<StoryUnfurl | null> {
    return this.request("/internal/bot/slack/unfurls/story", {
      body: { actor, url },
      method: "POST",
    });
  }

  private async request<T>(
    path: string,
    options: RequestOptions = {},
  ): Promise<T> {
    if (!this.apiUrl) {
      throw new Error("FORTYONE_API_URL is required for FortyOne API calls.");
    }
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
      body: options.body ? JSON.stringify(options.body) : undefined,
      headers: {
        Authorization: `Bearer ${this.serviceToken}`,
        "Content-Type": "application/json",
      },
      method: options.method ?? "GET",
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
