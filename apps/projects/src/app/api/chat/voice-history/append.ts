import type { UIMessage } from "ai";
import type { MessageWriteReservation } from "@/modules/ai-chats/actions/message-write";
import { ChatAdmissionError } from "../request-context-policy";

export type VoiceHistoryMessage = {
  id: string;
  role: "assistant" | "user";
  text: string;
};
export type VoiceHistoryDependencies = {
  read: () => Promise<UIMessage[]>;
  begin: (messages: UIMessage[]) => Promise<MessageWriteReservation>;
  finalize: (
    messages: UIMessage[],
    reservation: MessageWriteReservation,
  ) => Promise<{ applied: boolean }>;
};

export const appendVoiceHistory = async (
  incoming: VoiceHistoryMessage[],
  dependencies: VoiceHistoryDependencies,
): Promise<UIMessage[]> => {
  const current = await dependencies.read();
  const byId = new Map(current.map((message) => [message.id, message]));
  const additions: UIMessage[] = [];
  for (const message of incoming) {
    const existing = byId.get(message.id);
    if (existing) {
      if (
        existing.role !== message.role ||
        existing.parts.length !== 1 ||
        existing.parts[0]?.type !== "text" ||
        existing.parts[0].text !== message.text
      ) {
        throw new ChatAdmissionError(
          "A voice transcript ID already has different content. Reload this conversation.",
          409,
        );
      }
      continue;
    }
    const addition: UIMessage = {
      id: message.id,
      role: message.role,
      parts: [{ type: "text", text: message.text }],
    };
    byId.set(message.id, addition);
    additions.push(addition);
  }
  if (additions.length === 0) return current;
  const merged = [...current, ...additions];
  const lastUserIndex = merged.findLastIndex(
    (message) => message.role === "user",
  );
  if (lastUserIndex < current.length - 1 || lastUserIndex < 0) {
    throw new ChatAdmissionError(
      "Keep the voice reply with its user transcript and retry saving both together.",
      409,
    );
  }
  if (
    lastUserIndex < current.length &&
    incoming.findLast((message) => message.role === "user")?.id !==
      current.at(-1)?.id
  ) {
    throw new ChatAdmissionError(
      "Include the original user transcript with this voice reply before retrying.",
      409,
    );
  }
  // The existing generation ledger validates the canonical prefix, blocks open
  // approvals, and accepts only assistant suffixes during finalization.
  const reservationMessages = merged.slice(0, lastUserIndex + 1);
  const reservation = await dependencies.begin(reservationMessages);
  const canonical = [
    ...(reservation.messages ?? reservationMessages),
    ...merged.slice(lastUserIndex + 1),
  ];
  const result = await dependencies.finalize(canonical, reservation);
  if (!result.applied)
    throw new ChatAdmissionError(
      "This conversation changed while saving voice history. Reload and retry the same transcripts.",
      409,
    );
  return canonical;
};
