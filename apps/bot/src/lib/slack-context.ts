import type { Author } from "chat";
import type { SlackActor } from "@/lib/runtime";

type SlackRawObject = Record<string, unknown>;

const isObject = (value: unknown): value is SlackRawObject =>
  typeof value === "object" && value !== null;

const readString = (value: unknown): string | undefined =>
  typeof value === "string" && value.trim() ? value.trim() : undefined;

const readPath = (value: unknown, path: string[]): unknown =>
  path.reduce<unknown>((current, key) => {
    if (!isObject(current)) {
      return undefined;
    }
    return current[key];
  }, value);

const firstString = (raw: unknown, paths: string[][]): string | undefined => {
  for (const path of paths) {
    const value = readString(readPath(raw, path));
    if (value) {
      return value;
    }
  }
  return undefined;
};

export const slackActorFromEvent = (
  user: Author,
  raw: unknown,
  overrides: Partial<SlackActor> = {},
): SlackActor => ({
  channelId: firstString(raw, [
    ["channel", "id"],
    ["channel_id"],
    ["event", "channel"],
    ["container", "channel_id"],
  ]),
  channelName: firstString(raw, [["channel", "name"], ["channel_name"]]),
  messageTs: firstString(raw, [
    ["message", "ts"],
    ["event", "ts"],
    ["container", "message_ts"],
  ]),
  teamId: firstString(raw, [
    ["team", "id"],
    ["team_id"],
    ["event", "team"],
    ["user", "team_id"],
  ]),
  threadTs: firstString(raw, [
    ["message", "thread_ts"],
    ["event", "thread_ts"],
    ["container", "thread_ts"],
  ]),
  userId: overrides.userId ?? user.userId,
  userName: user.userName || user.fullName,
  ...overrides,
});

export const slackMessageTextFromRaw = (raw: unknown): string | undefined =>
  firstString(raw, [["message", "text"], ["event", "text"], ["text"]]);

export const findFortyOneUrls = (text: string): string[] => {
  const matches = text.match(/https?:\/\/[^\s<>"]+/g) ?? [];
  return matches.filter((url) => /\/stor(?:y|ies)\//i.test(url));
};
