/** @jsxImportSource chat */

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
import { FortyOneClient } from "@/lib/fortyone-client";
import {
  findFortyOneUrls,
  slackActorFromEvent,
  slackMessageTextFromRaw,
} from "@/lib/slack-context";
import { createBotState } from "@/lib/state";

const config = getBotConfig();
assertProductionConfig(config);

const client = new FortyOneClient(config);

const slackAdapter = createSlackAdapter({
  botToken: config.slackBotToken || undefined,
  clientId: config.slackClientId || undefined,
  clientSecret: config.slackClientSecret || undefined,
  encryptionKey: config.slackEncryptionKey || undefined,
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
  return normalized === "" || normalized === "create" || normalized.startsWith("create ");
};

const selectedTeamFromRaw = (raw: unknown): string | undefined => {
  const stateValues = (raw as { view?: { state?: { values?: unknown } } })?.view
    ?.state?.values;
  if (!stateValues || typeof stateValues !== "object") {
    return undefined;
  }

  const teamBlock = (stateValues as Record<string, Record<string, unknown>>)[
    CREATE_STORY_FIELDS.team
  ];
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
  const responseUrl = (raw as { response_url?: unknown })?.response_url;
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

const postAgentReply = async (
  thread: Parameters<Parameters<typeof bot.onNewMention>[0]>[0],
  message: Parameters<Parameters<typeof bot.onNewMention>[0]>[1],
) => {
  const actor = slackActorFromEvent(message.author, message.raw);
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

  if (state?.fortyOneStoryId && message.text.trim()) {
    await client.recordSlackThreadComment({
      actor,
      messageText: message.text,
      storyId: state.fortyOneStoryId,
    });
  }

  for (const url of findFortyOneUrls(message.text)) {
    const unfurl = await client.getStoryUnfurl(url, actor);
    if (unfurl) {
      await thread.post({
        markdown: `*${unfurl.title}*\n${[
          unfurl.status,
          unfurl.priority,
          unfurl.assignee ? `Assigned to ${unfurl.assignee}` : undefined,
        ]
          .filter(Boolean)
          .join(" • ")}\n${unfurl.url}`,
      });
    }
  }

  if (message.isMention || thread.isDM) {
    await thread.post(createAgentStream(message.text, config, client, actor));
  }
});

bot.onSlashCommand("/fortyone", async (event) => {
  const actor = slackActorFromEvent(event.user, event.raw);
  const slashResponseUrl = slashResponseUrlFromRaw(event.raw);

  if (!isCreateIntent(event.text)) {
    if (slashResponseUrl) {
      const reply = await generateAgentReply(event.text, config, client, actor);
      await postSlashResponse(slashResponseUrl, reply, "in_channel");
      return;
    }

    await event.channel.post(createAgentStream(event.text, config, client, actor));
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

  const result = await event.openModal(
    CreateStoryModal({
      description: source.messageText,
      source,
      title: source.messageText?.split("\n")[0]?.slice(0, 120),
    }),
  );

  if (!result && event.thread) {
    await event.thread.post("I couldn't open the create story form. Please try again.");
  }
});

bot.onOptionsLoad(async (event) => {
  const actor = slackActorFromEvent(event.user, event.raw);
  const teamId = selectedTeamFromRaw(event.raw);

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
    default:
      return [];
  }
});

bot.onModalSubmit(CREATE_STORY_MODAL_ID, async (event) => {
  const metadata = decodeCreateStoryMetadata(event.privateMetadata);
  const input = buildCreateStoryInput(event.values, metadata);

  if (!input.title.trim()) {
    return modalError(CREATE_STORY_FIELDS.title, "Add a title.");
  }

  if (!input.teamId.trim()) {
    return modalError(CREATE_STORY_FIELDS.team, "Choose a team.");
  }

  const story = await client.createStoryFromSlackForm(input);

  if (event.relatedThread) {
    await event.relatedThread.subscribe();
    await event.relatedThread.setState({ fortyOneStoryId: story.id });
    await event.relatedThread.post(`Created *${story.ref}*: ${story.url}`);
  } else if (event.relatedChannel) {
    await event.relatedChannel.post(`Created *${story.ref}*: ${story.url}`);
  }

  return { action: "close" };
});
