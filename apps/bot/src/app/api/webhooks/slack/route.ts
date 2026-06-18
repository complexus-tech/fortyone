import { createHmac, timingSafeEqual } from "node:crypto";
import type { NextRequest } from "next/server";
import { bot, openSlackCreateStoryModal } from "@/lib/bot";
import { getBotConfig } from "@/lib/config";
import { CREATE_STORY_ACTION_ID } from "@/lib/create-story-modal";
import { errorMessage, logBotError } from "@/lib/logger";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

interface SlackMessageShortcutPayload {
  callback_id?: unknown;
  channel?: {
    id?: unknown;
    name?: unknown;
  };
  message?: {
    text?: unknown;
    thread_ts?: unknown;
    ts?: unknown;
  };
  team?: {
    id?: unknown;
  };
  trigger_id?: unknown;
  type?: unknown;
  user?: {
    id?: unknown;
    name?: unknown;
    username?: unknown;
  };
}

const config = getBotConfig();

const asString = (value: unknown): string | undefined =>
  typeof value === "string" && value.trim() ? value : undefined;

const isValidSlackSignature = (
  request: NextRequest,
  rawBody: string,
): boolean => {
  if (!config.slackSigningSecret) {
    logBotError("Slack shortcut rejected", {
      error: "Missing Slack signing secret",
    });
    return false;
  }

  const timestamp = request.headers.get("x-slack-request-timestamp");
  const signature = request.headers.get("x-slack-signature");
  if (!timestamp || !signature) {
    return false;
  }

  const timestampSeconds = Number(timestamp);
  if (
    !Number.isFinite(timestampSeconds) ||
    Math.abs(Date.now() / 1000 - timestampSeconds) > 60 * 5
  ) {
    return false;
  }

  const expected = `v0=${createHmac("sha256", config.slackSigningSecret)
    .update(`v0:${timestamp}:${rawBody}`)
    .digest("hex")}`;
  const expectedBuffer = Buffer.from(expected);
  const actualBuffer = Buffer.from(signature);

  return (
    expectedBuffer.length === actualBuffer.length &&
    timingSafeEqual(expectedBuffer, actualBuffer)
  );
};

const parseSlackInteractivityPayload = (
  rawBody: string,
  contentType: string | null,
): SlackMessageShortcutPayload | undefined => {
  if (!contentType?.includes("application/x-www-form-urlencoded")) {
    return undefined;
  }

  const payload = new URLSearchParams(rawBody).get("payload");
  if (!payload) {
    return undefined;
  }

  return JSON.parse(payload) as SlackMessageShortcutPayload;
};

const isCreateStoryMessageShortcut = (
  payload: SlackMessageShortcutPayload | undefined,
): payload is SlackMessageShortcutPayload =>
  payload?.type === "message_action" &&
  payload.callback_id === CREATE_STORY_ACTION_ID;

const titleFromMessageText = (text: string | undefined): string | undefined =>
  text
    ?.split("\n")
    .map((line) => line.trim())
    .find(Boolean)
    ?.slice(0, 120);

const handleCreateStoryMessageShortcut = async (
  request: NextRequest,
  rawBody: string,
  payload: SlackMessageShortcutPayload,
): Promise<Response> => {
  if (!isValidSlackSignature(request, rawBody)) {
    return new Response("Invalid signature", { status: 401 });
  }

  const triggerId = asString(payload.trigger_id);
  if (!triggerId) {
    logBotError("Slack create-story shortcut ignored", {
      error: "Missing trigger_id",
    });
    return new Response(null, { status: 200 });
  }

  const messageText = asString(payload.message?.text);
  await openSlackCreateStoryModal({
    description: messageText,
    source: {
      channelId: asString(payload.channel?.id),
      channelName: asString(payload.channel?.name),
      messageText,
      messageTs: asString(payload.message?.ts),
      teamId: asString(payload.team?.id),
      threadTs: asString(payload.message?.thread_ts),
      userId: asString(payload.user?.id) ?? "",
      userName:
        asString(payload.user?.username) ?? asString(payload.user?.name),
    },
    title: titleFromMessageText(messageText),
    triggerId,
  });

  return new Response(null, { status: 200 });
};

export const POST = async (request: NextRequest): Promise<Response> => {
  try {
    const rawBody = await request.text();
    const payload = parseSlackInteractivityPayload(
      rawBody,
      request.headers.get("content-type"),
    );

    if (isCreateStoryMessageShortcut(payload)) {
      return await handleCreateStoryMessageShortcut(request, rawBody, payload);
    }

    return await bot.webhooks.slack(
      new Request(request.url, {
        body: rawBody,
        headers: request.headers,
        method: request.method,
      }),
    );
  } catch (error) {
    logBotError("Unhandled Slack webhook error", {
      error: errorMessage(error),
    });
    return Response.json({ error: "Slack webhook failed" }, { status: 500 });
  }
};
