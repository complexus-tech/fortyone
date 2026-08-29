import type { PrepareSendMessagesRequest, UIMessage } from "ai";
import type { MayaUIMessage } from "@/lib/ai/tools/types";

export const HISTORICAL_ATTACHMENT_PLACEHOLDER =
  "[Historical attachment omitted from this request.]";

export const omitHistoricalFileParts = <MESSAGE extends UIMessage>(
  messages: MESSAGE[],
  { preserveLastUserFiles = true }: { preserveLastUserFiles?: boolean } = {},
): MESSAGE[] => {
  const lastUserMessageIndex = preserveLastUserFiles
    ? messages.findLastIndex((message) => message.role === "user")
    : -1;

  return messages.map((message, messageIndex) => {
    if (messageIndex === lastUserMessageIndex) return message;
    if (!message.parts.some((part) => part.type === "file")) return message;

    return {
      ...message,
      parts: message.parts.map((part) =>
        part.type === "file"
          ? {
              text: HISTORICAL_ATTACHMENT_PLACEHOLDER,
              type: "text" as const,
            }
          : part,
      ),
    } as MESSAGE;
  });
};

type PrepareMayaChatRequestOptions = Parameters<
  PrepareSendMessagesRequest<MayaUIMessage>
>[0];

export const prepareMayaChatSendRequest = ({
  body,
  id,
  messageId,
  messages,
  trigger,
}: PrepareMayaChatRequestOptions) => ({
  body: {
    ...(body ?? {}),
    id,
    messages: omitHistoricalFileParts(messages, {
      // A normal submit has a new user turn at the end. Regeneration needs the
      // source user's files again, while an approval auto-submit does not.
      preserveLastUserFiles:
        messages.at(-1)?.role === "user" || trigger === "regenerate-message",
    }),
    trigger,
    messageId,
  },
});
