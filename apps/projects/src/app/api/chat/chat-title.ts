import type { UIMessage } from "ai";

const MAX_CHAT_TITLE_LENGTH = 64;
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
