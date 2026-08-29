import type { ChatStatus } from "ai";

export const isChatResponseInProgress = (status: ChatStatus) =>
  status === "submitted" || status === "streaming";

type SendGuard = { current: boolean };

export const runWithAtomicSendGuard = async (
  sendGuard: SendGuard,
  task: () => Promise<void>,
) => {
  if (sendGuard.current) return false;

  sendGuard.current = true;
  try {
    await task();
    return true;
  } finally {
    sendGuard.current = false;
  }
};

export const runWithChatSendGuard = async ({
  sendGuard,
  status,
  task,
}: {
  sendGuard: SendGuard;
  status: ChatStatus;
  task: () => Promise<void>;
}) => {
  if (isChatResponseInProgress(status)) return false;
  return runWithAtomicSendGuard(sendGuard, task);
};
