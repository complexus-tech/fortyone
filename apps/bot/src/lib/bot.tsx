/* @jsxImportSource chat */

import {
  Chat,
  type Message,
  type ModalErrorsResponse,
  type Thread,
} from "chat";
import { createSlackAdapter } from "@chat-adapter/slack";
import { createAgentStream, generateAgentReply } from "@/lib/agent";
import { assertProductionConfig, getBotConfig } from "@/lib/config";
import {
  buildCreateStoryInput,
  CREATE_STORY_ACTION_ID,
  CREATE_STORY_FIELDS,
  CREATE_STORY_MODAL_ID,
  CreateStoryModal,
  CreateStorySlackView,
  decodeCreateStoryMetadata,
  toSelectOptions,
} from "@/lib/create-story-modal";
import { FortyOneClient } from "@/lib/fortyone-client";
import { errorMessage, logBotError } from "@/lib/logger";
import { DEFAULT_TEAM_OPTION, localRuntime } from "@/lib/local-runtime";
import type { RuntimeOption, SlackActor, SlackIdentity } from "@/lib/runtime";
import {
  findFortyOneUrls,
  slackActorFromEvent,
  slackMessageTextFromRaw,
} from "@/lib/slack-context";
import { createBotState } from "@/lib/state";

const config = getBotConfig();
assertProductionConfig(config);

const fortyOneClient =
  config.fortyOneApiUrl && config.fortyOneServiceToken
    ? new FortyOneClient(config)
    : null;
const runtime = fortyOneClient ?? localRuntime;

const slackInstallationProvider = fortyOneClient
  ? {
      getInstallation: async (installationId: string) => {
        try {
          return await fortyOneClient.getSlackInstallation(installationId);
        } catch (error) {
          logBotError("Failed to resolve Slack installation", {
            error: errorMessage(error),
            installationId,
          });
          return null;
        }
      },
    }
  : undefined;

const slackAdapter = createSlackAdapter({
  botToken: config.slackBotToken || undefined,
  installationProvider: slackInstallationProvider,
  signingSecret: config.slackSigningSecret || undefined,
  userName: config.botUserName,
});

export const bot = new Chat({
  userName: config.botUserName,
  adapters: {
    slack: slackAdapter,
  },
  concurrency: "queue",
  dedupeTtlMs: 10 * 60 * 1000,
  fallbackStreamingPlaceholderText: null,
  state: createBotState(config),
  streamingUpdateIntervalMs: 1000,
  threadHistory: {
    maxMessages: 50,
    ttlMs: 30 * 24 * 60 * 60 * 1000,
  },
});

const isCreateIntent = (text: string) => {
  const normalized = text.trim().toLowerCase();
  return (
    normalized === "" ||
    normalized === "create" ||
    normalized.startsWith("create ")
  );
};

const selectedTeamFromRaw = (raw: unknown): string | undefined => {
  const stateValues = (raw as { view?: { state?: { values?: unknown } } }).view
    ?.state?.values;
  if (!stateValues || typeof stateValues !== "object") {
    return undefined;
  }

  for (const block of Object.values(
    stateValues as Record<string, Record<string, unknown>>,
  )) {
    const teamInput = block[CREATE_STORY_FIELDS.team];
    if (!teamInput || typeof teamInput !== "object") {
      continue;
    }

    const selected = (
      teamInput as {
        selected_option?: { value?: unknown };
        value?: unknown;
      }
    ).selected_option?.value;
    if (typeof selected === "string" && selected.trim()) {
      return selected;
    }

    const value = (teamInput as { value?: unknown }).value;
    if (typeof value === "string" && value.trim()) {
      return value;
    }
  }

  return undefined;
};

const modalError = (field: string, message: string): ModalErrorsResponse => ({
  action: "errors",
  errors: {
    [field]: message,
  },
});

const slashResponseUrlFromRaw = (raw: unknown) => {
  const responseUrl = (raw as { response_url?: unknown }).response_url;
  return typeof responseUrl === "string" && responseUrl.length > 0
    ? responseUrl
    : undefined;
};

const postSlashResponse = async (
  responseUrl: string,
  text: string,
  responseType: "ephemeral" | "in_channel" = "ephemeral",
) => {
  const response = await fetch(responseUrl, {
    body: JSON.stringify({
      response_type: responseType,
      text,
    }),
    headers: {
      "Content-Type": "application/json",
    },
    method: "POST",
  });

  if (!response.ok) {
    throw new Error(
      `Slack slash response failed: ${response.status} ${response.statusText}`,
    );
  }
};

const connectAccountMessage = (identity?: SlackIdentity | null) => {
  if (identity?.connectUrl) {
    return `Connect your FortyOne account before using Maya in Slack: ${identity.connectUrl}`;
  }

  return "Connect your FortyOne account before using Maya in Slack.";
};

const recordRuntimeLog = async (
  action: string,
  actor: SlackActor | undefined,
  error: unknown,
  metadata: Record<string, unknown> = {},
) => {
  const message = errorMessage(error);
  logBotError(action, {
    ...metadata,
    error: message,
    slackChannelId: actor?.channelId,
    slackTeamId: actor?.teamId,
    slackUserId: actor?.userId,
  });

  try {
    await runtime.recordSlackRuntimeLog?.({
      actor,
      endpoint: action,
      errorMessage: message,
      outcome: "failed",
      requestType: "bot_runtime",
      responseCode: 500,
    });
  } catch (logError) {
    logBotError("Failed to record Slack runtime log", {
      error: errorMessage(logError),
      sourceAction: action,
      slackTeamId: actor?.teamId,
      slackUserId: actor?.userId,
    });
  }
};

const resolveIdentity = async (
  actor: SlackActor,
  action: string,
): Promise<SlackIdentity | null> => {
  try {
    return await runtime.resolveSlackIdentity(actor);
  } catch (error) {
    await recordRuntimeLog(action, actor, error);
    return null;
  }
};

const defaultTeamOptionForActor = async (
  actor: SlackActor,
): Promise<RuntimeOption | undefined> => {
  try {
    const teams = await runtime.searchTeams(actor, "");
    return teams[0] ?? (fortyOneClient ? undefined : DEFAULT_TEAM_OPTION);
  } catch (error) {
    await recordRuntimeLog("default_team_failed", actor, error);
    return fortyOneClient ? undefined : DEFAULT_TEAM_OPTION;
  }
};

const teamIdForOptions = async (actor: SlackActor, raw: unknown) =>
  selectedTeamFromRaw(raw) ?? (await defaultTeamOptionForActor(actor))?.value;

const postAgentReply = async (thread: Thread, message: Message) => {
  const actor = slackActorFromEvent(message.author, message.raw);
  const identity = await resolveIdentity(actor, "ai_reply_identity");

  if (!identity?.userId) {
    await thread.post(connectAccountMessage(identity));
    return;
  }

  await thread.post(createAgentStream(message.text, config, runtime, actor));
};

const slackBotTokenForSource = async (source: SlackActor) => {
  if (source.teamId && fortyOneClient) {
    try {
      const installation = await fortyOneClient.getSlackInstallation(
        source.teamId,
      );
      if (installation?.botToken) {
        return installation.botToken;
      }
    } catch (error) {
      await recordRuntimeLog("slack_installation_lookup_failed", source, error);
    }
  }

  return config.slackBotToken || undefined;
};

export const openSlackCreateStoryModal = async ({
  description,
  initialTeam,
  source,
  title,
  triggerId,
}: {
  description?: string;
  initialTeam?: RuntimeOption;
  source: Parameters<typeof CreateStorySlackView>[0]["source"];
  title?: string;
  triggerId: string;
}) => {
  const botToken = await slackBotTokenForSource(source);
  if (!botToken) {
    logBotError("Slack create-story modal skipped", {
      error: "Missing Slack bot token",
      slackTeamId: source.teamId,
    });
    return undefined;
  }

  const resolvedInitialTeam =
    initialTeam ?? (await defaultTeamOptionForActor(source));
  const response = await fetch("https://slack.com/api/views.open", {
    body: JSON.stringify({
      trigger_id: triggerId,
      view: CreateStorySlackView({
        description,
        initialTeam: resolvedInitialTeam,
        source,
        title,
      }),
    }),
    headers: {
      Authorization: `Bearer ${botToken}`,
      "Content-Type": "application/json",
    },
    method: "POST",
  });

  const body = (await response.json()) as {
    error?: string;
    ok?: boolean;
    view?: { id?: string };
  };

  if (!response.ok || !body.ok) {
    logBotError("Slack raw create-story modal failed", {
      error: body.error ?? `${response.status} ${response.statusText}`,
    });
    return undefined;
  }

  return body.view?.id ? { viewId: body.view.id } : undefined;
};

const openCreateStoryModal = async ({
  description,
  fallback,
  initialTeam,
  source,
  title,
  triggerId,
}: {
  description?: string;
  fallback: () => Promise<{ viewId: string } | undefined>;
  initialTeam?: RuntimeOption;
  source: Parameters<typeof CreateStorySlackView>[0]["source"];
  title?: string;
  triggerId?: string;
}) => {
  if (!triggerId) {
    return fallback();
  }

  const result = await openSlackCreateStoryModal({
    description,
    initialTeam,
    source,
    title,
    triggerId,
  });

  return result ?? fallback();
};

bot.onNewMention(async (thread, message) => {
  await thread.subscribe();
  await postAgentReply(thread, message);
});

bot.onDirectMessage(async (thread, message) => {
  await thread.subscribe();
  await postAgentReply(thread, message);
});

bot.onSubscribedMessage(async (thread, message) => {
  if (message.author.isBot || message.author.isMe) {
    return;
  }

  const actor = slackActorFromEvent(message.author, message.raw);
  const identity = await resolveIdentity(actor, "subscribed_message_identity");
  if (!identity?.userId) {
    if (message.isMention || thread.isDM) {
      await thread.post(connectAccountMessage(identity));
    }
    return;
  }

  const state = (await thread.state) as { fortyOneStoryId?: string } | null;
  if (state?.fortyOneStoryId && message.text.trim()) {
    try {
      await runtime.recordSlackThreadComment?.({
        actor,
        messageText: message.text,
        storyId: state.fortyOneStoryId,
      });
    } catch (error) {
      await recordRuntimeLog("slack_thread_comment_failed", actor, error, {
        storyId: state.fortyOneStoryId,
      });
    }
  }

  if (runtime.getStoryUnfurl) {
    const unfurls = await Promise.all(
      findFortyOneUrls(message.text).map((url) =>
        runtime.getStoryUnfurl?.(url, actor),
      ),
    );

    await Promise.all(
      unfurls
        .filter((unfurl): unfurl is NonNullable<typeof unfurl> =>
          Boolean(unfurl),
        )
        .map((unfurl) =>
          thread.post({
            markdown: `*${unfurl.title}*\n${[
              unfurl.status,
              unfurl.priority,
              unfurl.assignee ? `Assigned to ${unfurl.assignee}` : undefined,
            ]
              .filter(Boolean)
              .join(" • ")}\n${unfurl.url}`,
          }),
        ),
    );
  }

  if (message.isMention || thread.isDM) {
    await thread.post(createAgentStream(message.text, config, runtime, actor));
  }
});

bot.onSlashCommand("/fortyone", async (event) => {
  const actor = slackActorFromEvent(event.user, event.raw);
  const slashResponseUrl = slashResponseUrlFromRaw(event.raw);
  const identity = await resolveIdentity(actor, "slash_command_identity");

  if (!identity?.userId) {
    const text = connectAccountMessage(identity);
    if (slashResponseUrl) {
      await postSlashResponse(slashResponseUrl, text);
      return;
    }
    await event.channel.postEphemeral(event.user, text, { fallbackToDM: true });
    return;
  }

  if (!isCreateIntent(event.text)) {
    const reply = await generateAgentReply(event.text, config, runtime, actor);

    if (slashResponseUrl) {
      await postSlashResponse(slashResponseUrl, reply, "in_channel");
      return;
    }

    await event.channel.post(reply);
    return;
  }

  const initialTeam = await defaultTeamOptionForActor(actor);
  const result = await openCreateStoryModal({
    fallback: () =>
      event.openModal(
        CreateStoryModal({
          initialTeam,
          source: actor,
          title: event.text.replace(/^create\s*/i, ""),
        }),
      ),
    initialTeam,
    source: actor,
    title: event.text.replace(/^create\s*/i, ""),
    triggerId: event.triggerId,
  });

  if (!result) {
    const text = "I couldn't open the create story form. Please try again.";
    if (slashResponseUrl) {
      await postSlashResponse(slashResponseUrl, text);
      return;
    }

    await event.channel.postEphemeral(event.user, text, { fallbackToDM: true });
  }
});

bot.onAction(CREATE_STORY_ACTION_ID, async (event) => {
  const source = {
    ...slackActorFromEvent(event.user, event.raw),
    messageText: slackMessageTextFromRaw(event.raw),
  };
  const identity = await resolveIdentity(source, "create_action_identity");

  if (!identity?.userId) {
    const text = connectAccountMessage(identity);
    if (event.thread) {
      await event.thread.post(text);
      return;
    }

    const responseUrl = slashResponseUrlFromRaw(event.raw);
    if (responseUrl) {
      await postSlashResponse(responseUrl, text);
    }
    return;
  }

  const initialTeam = await defaultTeamOptionForActor(source);
  const result = await openCreateStoryModal({
    description: source.messageText,
    fallback: () =>
      event.openModal(
        CreateStoryModal({
          description: source.messageText,
          initialTeam,
          source,
          title: source.messageText?.split("\n")[0]?.slice(0, 120),
        }),
      ),
    initialTeam,
    source,
    title: source.messageText?.split("\n")[0]?.slice(0, 120),
    triggerId: event.triggerId,
  });

  if (!result && event.thread) {
    await event.thread.post(
      "I couldn't open the create story form. Please try again.",
    );
  }
});

bot.onOptionsLoad(async (event) => {
  const actor = slackActorFromEvent(event.user, event.raw);
  const teamId = await teamIdForOptions(actor, event.raw);

  try {
    switch (event.actionId) {
      case CREATE_STORY_FIELDS.team:
        return toSelectOptions(await runtime.searchTeams(actor, event.query));
      case CREATE_STORY_FIELDS.status:
        return toSelectOptions(
          await runtime.searchStatuses(actor, teamId, event.query),
        );
      case CREATE_STORY_FIELDS.assignee:
        return toSelectOptions(
          await runtime.searchMembers(actor, teamId, event.query),
        );
      case CREATE_STORY_FIELDS.objective:
        return toSelectOptions(
          await runtime.searchObjectives(actor, teamId, event.query),
        );
      case CREATE_STORY_FIELDS.label:
        return toSelectOptions(
          await runtime.searchLabels(actor, teamId, event.query),
        );
      default:
        return [];
    }
  } catch (error) {
    await recordRuntimeLog("slack_options_load_failed", actor, error, {
      actionId: event.actionId,
      teamId,
    });
    return [];
  }
});

bot.onModalSubmit(CREATE_STORY_MODAL_ID, async (event) => {
  const metadata = decodeCreateStoryMetadata(event.privateMetadata);
  const input = buildCreateStoryInput(event.values, metadata, event.raw);
  const identity = await resolveIdentity(
    input.source,
    "create_story_submit_identity",
  );

  if (!identity?.userId) {
    return modalError(
      CREATE_STORY_FIELDS.title,
      "Connect your FortyOne account before creating stories from Slack.",
    );
  }

  if (!input.title.trim()) {
    return modalError(CREATE_STORY_FIELDS.title, "Add a title.");
  }

  if (!input.teamId.trim()) {
    return modalError(CREATE_STORY_FIELDS.team, "Choose a team.");
  }

  let story;
  try {
    story = await runtime.createStoryFromSlackForm(input);
  } catch (error) {
    await recordRuntimeLog("create_story_failed", input.source, error, {
      teamId: input.teamId,
    });
    return modalError(
      CREATE_STORY_FIELDS.title,
      "I couldn't create the story. Please try again.",
    );
  }

  const linkedTitle = story.url ? `<${story.url}|${story.title}>` : story.title;
  const text = `Created *${story.ref}*: ${linkedTitle}`;

  try {
    if (event.relatedThread) {
      await event.relatedThread.subscribe();
      await event.relatedThread.setState({ fortyOneStoryId: story.id });
      await event.relatedThread.post(text);
    } else if (event.relatedChannel) {
      await event.relatedChannel.post(text);
    }
  } catch (error) {
    await recordRuntimeLog(
      "slack_story_confirmation_failed",
      input.source,
      error,
      {
        storyId: story.id,
        teamId: input.teamId,
      },
    );
  }

  return { action: "close" };
});
