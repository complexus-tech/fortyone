import type { FileUIPart } from "ai";

/** Fences the UI work that happens before the chat controller owns a request. */
export const sendMayaDraft = async ({
  text,
  prepare,
  flushVoiceHistory,
  send,
  isCurrent,
  onSent,
}: {
  text: string;
  prepare: () => Promise<FileUIPart[]>;
  flushVoiceHistory: () => Promise<void>;
  send: (text: string, files: FileUIPart[]) => Promise<void>;
  isCurrent: () => boolean;
  onSent: () => void;
}): Promise<boolean> => {
  if (!isCurrent()) return false;
  try {
    const files = await prepare();
    if (!isCurrent()) return false;
    await flushVoiceHistory();
    if (!isCurrent()) return false;
    await send(text, files);
    if (!isCurrent()) return false;
    onSent();
    return true;
  } catch (error) {
    if (!isCurrent()) return false;
    throw error;
  }
};
