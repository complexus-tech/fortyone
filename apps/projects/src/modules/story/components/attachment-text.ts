import type { StoryAttachment } from "../types";
import { resolveCommentAttachmentUrl } from "./comment-attachment-download";

export const MAX_TEXT_PREVIEW_BYTES = 256 * 1024;

export const readAttachmentText = async (
  file: StoryAttachment,
  signal: AbortSignal,
) => {
  const url = await resolveCommentAttachmentUrl(file, signal);
  const response = await fetch(url, {
    credentials: "omit",
    headers: { Range: `bytes=0-${MAX_TEXT_PREVIEW_BYTES - 1}` },
    signal,
  });
  if (!response.ok || !response.body)
    throw new Error("This text preview could not be loaded.");
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let text = "";
  let bytes = 0;
  let truncated = file.size > MAX_TEXT_PREVIEW_BYTES;
  try {
    while (bytes < MAX_TEXT_PREVIEW_BYTES) {
      // eslint-disable-next-line no-await-in-loop -- Read sequential stream chunks so the byte limit stops further downloads.
      const { value, done } = await reader.read();
      if (done) break;
      const remaining = MAX_TEXT_PREVIEW_BYTES - bytes;
      text += decoder.decode(value.subarray(0, remaining), { stream: true });
      bytes += Math.min(value.byteLength, remaining);
      if (value.byteLength > remaining) truncated = true;
    }
    text += decoder.decode();
  } finally {
    await reader.cancel();
  }
  return { text, truncated };
};
