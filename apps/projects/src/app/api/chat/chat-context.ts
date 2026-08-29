import type { ModelMessage, UIMessage } from "ai";
import { pruneMessages } from "ai";
import { compactToolOutput } from "@/lib/ai/compact-tool-output";

// Roughly 600k typical English text tokens, leaving substantial room in Maya's
// model context for system instructions, tool schemas, images, and generation.
// Bytes are used instead of JavaScript character count so Unicode cannot make
// the request several times larger than the safety budget.
export const MAX_CHAT_CONTEXT_BYTES = 2_400_000;
const HISTORICAL_ATTACHMENT_PLACEHOLDER =
  "[Historical attachment omitted from model context]";
const textEncoder = new TextEncoder();

export const assertLatestUserTextWithinContextBudget = (
  messages: UIMessage[],
) => {
  const latestUserMessage = messages.findLast(
    (message) => message.role === "user",
  );
  if (!latestUserMessage) return;

  let textBytes = 0;
  for (const part of latestUserMessage.parts) {
    if (part.type !== "text") continue;
    textBytes += textEncoder.encode(part.text).byteLength;
    if (textBytes <= MAX_CHAT_CONTEXT_BYTES) continue;

    throw Object.assign(new Error("Maya's latest message text is too large."), {
      code: "request_too_large",
    });
  }
};

export const getChatContextStartIndex = (messages: UIMessage[]) => {
  const firstUserMessageIndex = messages.findIndex(
    (message) => message.role === "user",
  );
  const userBoundary = firstUserMessageIndex > 0 ? firstUserMessageIndex : 0;

  let retainedBytes = 0;
  let retainedStartIndex = messages.length;
  for (let index = messages.length - 1; index >= userBoundary; index -= 1) {
    const messageBytes = textEncoder.encode(
      JSON.stringify(messages[index]),
    ).byteLength;
    if (
      retainedBytes > 0 &&
      retainedBytes + messageBytes > MAX_CHAT_CONTEXT_BYTES
    ) {
      break;
    }

    retainedBytes += messageBytes;
    retainedStartIndex = index;
  }

  const retainedMessages = messages.slice(retainedStartIndex);
  const retainedUserBoundary = retainedMessages.findIndex(
    (message) => message.role === "user",
  );
  return retainedUserBoundary > 0
    ? retainedStartIndex + retainedUserBoundary
    : retainedStartIndex;
};

export const selectChatContextMessages = (messages: UIMessage[]) =>
  messages.slice(getChatContextStartIndex(messages));

/**
 * Keep attachments on the current user turn, but do not repeatedly send prior
 * binary data or hosted files to the model. The placeholder preserves the fact
 * that an attachment existed without retaining user-controlled file metadata.
 */
export const omitHistoricalChatAttachments = (messages: UIMessage[]) => {
  const latestUserMessageIndex = messages.findLastIndex(
    (message) => message.role === "user",
  );

  return messages.map((message, messageIndex): UIMessage => {
    if (messageIndex === latestUserMessageIndex) return message;

    const parts = message.parts.flatMap((part) =>
      part.type === "file"
        ? [
            {
              text: HISTORICAL_ATTACHMENT_PLACEHOLDER,
              type: "text" as const,
            },
          ]
        : [part],
    );

    return { ...message, parts };
  });
};

export const compactChatToolOutputs = (messages: UIMessage[]) =>
  messages.map(
    (message): UIMessage => ({
      ...message,
      parts: message.parts.map((part) => {
        if (!part.type.startsWith("tool-") || !("output" in part)) return part;

        return {
          ...part,
          output: compactToolOutput(part.output, {
            toolName: part.type.slice("tool-".length),
          }),
        } as typeof part;
      }),
    }),
  );

/**
 * Registered tools project their own raw UI receipts exactly once during
 * convertToModelMessages. Historical tools that no longer exist in the
 * registry have no projector, so compact only those fallback outputs before
 * conversion instead of sending an unbounded legacy payload to the model.
 */
export const compactUnknownChatToolOutputs = (
  messages: UIMessage[],
  registeredToolNames: ReadonlySet<string>,
) =>
  messages.map(
    (message): UIMessage => ({
      ...message,
      parts: message.parts.map((part) => {
        if (!part.type.startsWith("tool-") || !("output" in part)) return part;

        const toolName = part.type.slice("tool-".length);
        if (registeredToolNames.has(toolName)) return part;

        return {
          ...part,
          output: compactToolOutput(part.output, { toolName }),
        } as typeof part;
      }),
    }),
  );

export const pruneChatModelMessages = (messages: ModelMessage[]) =>
  pruneMessages({
    messages,
    // OpenAI Responses links assistant output items to their preceding
    // reasoning item. Removing only the reasoning part leaves an otherwise
    // valid-looking history that the provider rejects on the next turn.
    reasoning: "none",
    // Tool outputs are compacted before conversion and the outer byte budget
    // bounds the full request. Keep exact mutation receipts so later requests
    // such as "delete those stories" still resolve the original IDs.
    toolCalls: "none",
    emptyMessages: "remove",
  });
