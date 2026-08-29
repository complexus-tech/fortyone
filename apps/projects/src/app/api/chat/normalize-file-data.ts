import type { ModelMessage } from "ai";

const INLINE_DATA_URL_PREFIX = "data:";

const getBase64Payload = (data: string) => {
  if (!data.startsWith(INLINE_DATA_URL_PREFIX)) {
    return null;
  }

  const separatorIndex = data.indexOf(",");
  if (separatorIndex === -1) {
    return null;
  }

  const metadata = data.slice(
    INLINE_DATA_URL_PREFIX.length,
    separatorIndex,
  );
  const encodings = metadata.split(";").slice(1);
  if (!encodings.includes("base64")) {
    return null;
  }

  return data.slice(separatorIndex + 1);
};

/**
 * AI SDK 6 treats data URLs as remote assets during its download pass. Passing
 * the raw base64 payload keeps inline attachments local and lets each provider
 * serialize them using its native multimodal input format.
 */
export const normalizeInlineFileData = (
  messages: ModelMessage[],
): ModelMessage[] =>
  messages.map((message) => {
    if (message.role !== "user" || typeof message.content === "string") {
      return message;
    }

    const content = message.content.map((part) => {
      if (part.type !== "file" || typeof part.data !== "string") {
        return part;
      }

      const base64Payload = getBase64Payload(part.data);
      if (base64Payload === null) {
        return part;
      }

      return {
        ...part,
        data: base64Payload,
      };
    });

    return { ...message, content };
  });
