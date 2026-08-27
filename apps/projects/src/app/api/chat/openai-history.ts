import type { AssistantModelMessage, ModelMessage } from "ai";

type AssistantPart = Exclude<AssistantModelMessage["content"], string>[number];

const getOpenAIItemId = (part: AssistantPart) => {
  if (!("providerOptions" in part)) return undefined;

  const itemId = part.providerOptions?.openai?.itemId;
  return typeof itemId === "string" ? itemId : undefined;
};

const removeOpenAIItemId = (part: AssistantPart): AssistantPart => {
  if (!("providerOptions" in part)) return part;

  const providerOptions = part.providerOptions;
  const openAIOptions = providerOptions?.openai;
  if (typeof openAIOptions?.itemId !== "string") return part;

  const { itemId: _itemId, ...remainingOpenAIOptions } = openAIOptions;
  const remainingProviderOptions = { ...providerOptions };

  if (Object.keys(remainingOpenAIOptions).length === 0) {
    delete remainingProviderOptions.openai;
  } else {
    remainingProviderOptions.openai = remainingOpenAIOptions;
  }

  return {
    ...part,
    providerOptions:
      Object.keys(remainingProviderOptions).length === 0
        ? undefined
        : remainingProviderOptions,
  } as AssistantPart;
};

/**
 * OpenAI Responses output-message references are linked to the reasoning item
 * from the same response. UI streams may intentionally omit reasoning while
 * retaining the message item ID. Replaying that orphaned ID makes OpenAI reject
 * the entire follow-up request, so reconstruct only those assistant blocks as
 * ordinary assistant input while preserving all other provider metadata.
 */
export const sanitizeOpenAIHistoryItemReferences = (
  messages: ModelMessage[],
): ModelMessage[] =>
  messages.map((message): ModelMessage => {
    if (message.role !== "assistant" || typeof message.content === "string") {
      return message;
    }

    const hasLinkedReasoningItem = message.content.some(
      (part) =>
        part.type === "reasoning" && getOpenAIItemId(part) !== undefined,
    );
    if (hasLinkedReasoningItem) return message;

    const content = message.content.map(removeOpenAIItemId);
    const changed = content.some(
      (part, index) => part !== message.content[index],
    );

    return changed ? { ...message, content } : message;
  });
