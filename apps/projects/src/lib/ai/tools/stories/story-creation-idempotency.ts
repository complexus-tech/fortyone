import { createHash } from "node:crypto";

const MAX_IDEMPOTENCY_KEY_CHARACTERS = 128;
const DIGEST_CHARACTERS = 16;

type StoryCreationContext = {
  chatId?: string;
};

const fitIdempotencyKey = (key: string) => {
  if (key.length <= MAX_IDEMPOTENCY_KEY_CHARACTERS) return key;

  const digest = createHash("sha256")
    .update(key)
    .digest("hex")
    .slice(0, DIGEST_CHARACTERS);
  const prefixCharacters = MAX_IDEMPOTENCY_KEY_CHARACTERS - digest.length - 1;

  return `${key.slice(0, prefixCharacters)}:${digest}`;
};

export const getStoryCreationIdempotencyKey = ({
  context,
  index,
  toolCallId,
}: {
  context: unknown;
  index?: number;
  toolCallId: string;
}) => {
  const chatId = (context as StoryCreationContext | undefined)?.chatId?.trim();
  const operationKey = chatId
    ? `maya:${chatId}:${toolCallId}`
    : `maya:${toolCallId}`;
  const indexedKey =
    index === undefined ? operationKey : `${operationKey}:${index}`;

  return fitIdempotencyKey(indexedKey);
};
