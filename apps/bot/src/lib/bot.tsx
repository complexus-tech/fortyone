/* @jsxImportSource chat */

import { Chat, type ModalErrorsResponse } from "chat";
import { createSlackAdapter } from "@chat-adapter/slack";
import { createAgentStream, generateAgentReply } from "@/lib/agent";
import { assertProductionConfig, getBotConfig } from "@/lib/config";
import {
  buildCreateStoryInput,
  CREATE_STORY_ACTION_ID,
  CREATE_STORY_FIELDS,
  CREATE_STORY_MODAL_ID,
  CreateStoryModal,
  decodeCreateStoryMetadata,
  toSelectOptions,
} from "@/lib/create-story-modal";
import {
  FortyOneClient,
  type SlackActor,
  type SlackIdentity,
  type StoryUnfurl,
} from "@/lib/fortyone-client";
import {
  findFortyOneUrls,
  slackActorFromEvent,
  slackMessageTextFromRaw,
} from "@/lib/slack-context";
import { createBotState } from "@/lib/state";
import { actorLogFields, errorMessage, logBotError } from "@/lib/logger";

const config = getBotConfig();
assertProductionConfig(config);

const client = new FortyOneClient(config);

const slackInstallationProvider = {
  getInstallation: async (installationId: string) => {
    try {
      return await client.getSlackInstallation(installationId);
    } catch (error) {
      logBotError("Failed to resolve Slack installation", {
        error: errorMessage(error),
        installationId,
      });
      return null;
    }
  },
};

const slackAdapter = createSlackAdapter({
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

  const teamBlock = (
    stateValues as Partial<Record<string, Record<string, unknown>>>
  )[CREATE_STORY_FIELDS.team];
  if (!teamBlock) {
    return undefined;
  }

  for (const value of Object.values(teamBlock)) {
    const selected = (value as { selected_option?: { value?: string } })
      .selected_option?.value;
    if (selected) {
      return selected;
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
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      response_type: responseType,
      text,
    }),
  });

  if (!response.ok) {
    throw new Error(
      `Slack slash response failed: ${response.status} ${response.statusText}`,
    );
  }
};

const isStoryUnfurl = (unfurl: StoryUnfurl | null): unfurl is StoryUnfurl =>
  Boolean(unfurl);

const connectAccountMessage = (identity?: SlackIdentity | null) => {
  if (identity?.connectUrl) {
    return `Connect your FortyOne account before using Maya in Slack: ${identity.connectUrl}`;
  }

  return "Connect your FortyOne account before using Maya in Slack.";
};

const recordRuntimeLog = async (
  action: string,
  actor: SlackActor,
  error: unknown,
  metadata: Record<string, unknown> = {},
) => {
  const message = errorMessage(error);
  logBotError(action, {
    ...actorLogFields(actor),
    error: message,
    ...metadata,
  });

  try {
    await client.recordSlackRuntimeLog({
      action,
      actor,
      error: message,
      metadata,
    });
  } catch (logError) {
    logBotError("Failed to record Slack runtime log", {
      ...actorLogFields(actor),
      error: errorMessage(logError),
      sourceAction: action,
    });
  }
};

const resolveIdentity = async (
  actor: SlackActor,
  action: string,
): Promise<SlackIdentity | null> => {
  try {
    return await client.resolveSlackIdentity(actor);
  } catch (error) {
    await recordRuntimeLog(action, actor, error);
    return null;
  }
};

const postAgentReply = async (
  thread: Parameters<Parameters<typeof bot.onNewMention>[0]>[0],
  message: Parameters<Parameters<typeof bot.onNewMention>[0]>[1],
) => {
  const actor = slackActorFromEvent(message.author, message.raw);
  const identity = await resolveIdentity(actor, "ai_reply_identity");
  if (!identity?.userId) {
    await thread.post(connectAccountMessage(identity));
    return;
  }

  await thread.post(createAgentStream(message.text, config, client, actor));
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

  const state = (await thread.state) as { fortyOneStoryId?: string } | null;
  const actor = slackActorFromEvent(message.author, message.raw);
  const identity = await resolveIdentity(actor, "subscribed_message_identity");
  if (!identity?.userId) {
    if (message.isMention || thread.isDM) {
      await thread.post(connectAccountMessage(identity));
    }
    return;
  }

  if (state?.fortyOneStoryId && message.text.trim()) {
    await client.recordSlackThreadComment({
      actor,
      messageText: message.text,
      storyId: state.fortyOneStoryId,
    });
  }

  const unfurls = await Promise.all(
    findFortyOneUrls(message.text).map((url) =>
      client.getStoryUnfurl(url, actor),
    ),
  );

  await Promise.all(
    unfurls.filter(isStoryUnfurl).map((unfurl) =>
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

  if (message.isMention || thread.isDM) {
    await thread.post(createAgentStream(message.text, config, client, actor));
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
    if (slashResponseUrl) {
      const reply = await generateAgentReply(event.text, config, client, actor);
      await postSlashResponse(slashResponseUrl, reply, "in_channel");
      return;
    }

    await event.channel.post(
      createAgentStream(event.text, config, client, actor),
    );
    return;
  }

  const result = await event.openModal(
    CreateStoryModal({
      source: actor,
      title: event.text.replace(/^create\s*/i, ""),
    }),
  );

  if (!result) {
    if (slashResponseUrl) {
      await postSlashResponse(
        slashResponseUrl,
        "I couldn't open the create story form. Please try again.",
      );
      return;
    }

    await event.channel.postEphemeral(
      event.user,
      "I couldn't open the create story form. Please try again.",
      { fallbackToDM: true },
    );
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

  const result = await event.openModal(
    CreateStoryModal({
      description: source.messageText,
      source,
      title: source.messageText?.split("\n")[0]?.slice(0, 120),
    }),
  );

  if (!result && event.thread) {
    await event.thread.post(
      "I couldn't open the create story form. Please try again.",
    );
  }
});

bot.onOptionsLoad(async (event) => {
  const actor = slackActorFromEvent(event.user, event.raw);
  const teamId = selectedTeamFromRaw(event.raw);

  try {
    switch (event.actionId) {
      case CREATE_STORY_FIELDS.team:
        return toSelectOptions(await client.searchTeams(actor, event.query));
      case CREATE_STORY_FIELDS.status:
        if (!teamId) {
          return [];
        }
        return toSelectOptions(
          await client.searchStatuses(actor, teamId, event.query),
        );
      case CREATE_STORY_FIELDS.assignee:
        if (!teamId) {
          return [];
        }
        return toSelectOptions(
          await client.searchMembers(actor, teamId, event.query),
        );
      case CREATE_STORY_FIELDS.objective:
        if (!teamId) {
          return [];
        }
        return toSelectOptions(
          await client.searchObjectives(actor, teamId, event.query),
        );
      case CREATE_STORY_FIELDS.label:
        if (!teamId) {
          return [];
        }
        return toSelectOptions(
          await client.searchLabels(actor, teamId, event.query),
        );
      default:
        return [];
    }
  } catch (error) {
    await recordRuntimeLog("options_load_failed", actor, error, {
      actionId: event.actionId,
      teamId,
    });
    return [];
  }
});

bot.onModalSubmit(CREATE_STORY_MODAL_ID, async (event) => {
  const metadata = decodeCreateStoryMetadata(event.privateMetadata);
  const input = buildCreateStoryInput(event.values, metadata);
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
    story = await client.createStoryFromSlackForm(input);
  } catch (error) {
    await recordRuntimeLog("create_story_failed", input.source, error, {
      teamId: input.teamId,
    });
    return modalError(
      CREATE_STORY_FIELDS.title,
      "I couldn't create the story. Please try again.",
    );
  }

  if (event.relatedThread) {
    await event.relatedThread.subscribe();
    await event.relatedThread.setState({ fortyOneStoryId: story.id });
    await event.relatedThread.post(`Created *${story.ref}*: ${story.url}`);
  } else if (event.relatedChannel) {
    await event.relatedChannel.post(`Created *${story.ref}*: ${story.url}`);
  }

  return { action: "close" };
});
