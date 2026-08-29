import type { UIMessage } from "ai";

const MAX_CHAT_TITLE_LENGTH = 64;
const MAX_TITLE_SOURCE_LENGTH = 500;
const FALLBACK_CHAT_TITLE = "New conversation";

const truncateAtWordBoundary = (value: string) => {
  if (value.length <= MAX_CHAT_TITLE_LENGTH) return value;

  const candidate = value.slice(0, MAX_CHAT_TITLE_LENGTH - 1);
  const lastSpaceIndex = candidate.lastIndexOf(" ");
  const truncated =
    lastSpaceIndex >= MAX_CHAT_TITLE_LENGTH / 2
      ? candidate.slice(0, lastSpaceIndex)
      : candidate;

  return `${truncated.trimEnd()}…`;
};

export const getChatTitle = (messages: UIMessage[]) => {
  const firstUserMessage = messages.find((message) => message.role === "user");
  const text = firstUserMessage?.parts
    .flatMap((part) => (part.type === "text" ? [part.text] : []))
    .join(" ")
    .replace(/\s+/g, " ")
    .trim();

  if (!text) return FALLBACK_CHAT_TITLE;

  const firstSentence = text.split(/(?<=[.!?])\s|\n/u, 1)[0] ?? text;
  return truncateAtWordBoundary(firstSentence);
};

export const getChatTitleSource = (messages: UIMessage[]) => {
  const firstUserMessage = messages.find((message) => message.role === "user");
  return (
    firstUserMessage?.parts
      .flatMap((part) => (part.type === "text" ? [part.text] : []))
      .join(" ")
      .replace(/\s+/g, " ")
      .trim()
      .slice(0, MAX_TITLE_SOURCE_LENGTH) ?? ""
  );
};

export const normalizeGeneratedChatTitle = (value: string) => {
  const normalized = value
    .split("\n", 1)[0]
    .replace(/^title:\s*/i, "")
    .replace(/^[\s"'`]+|[\s"'`]+$/g, "")
    .replace(/\s+/g, " ")
    .trim();

  return normalized ? truncateAtWordBoundary(normalized) : "";
};
