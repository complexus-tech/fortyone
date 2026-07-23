import type { MayaUIMessage } from "@/lib/ai/tools/types";

type RealtimeTranscriptEvent = {
  delta?: string;
  item_id?: string;
  response_id?: string;
  transcript?: string;
  type?: string;
};

type RealtimeTranscriptUpdate = {
  id: string;
  mode: "append" | "replace";
  role: "assistant" | "user";
  text: string;
};

export const getMayaMessageText = (message: MayaUIMessage) =>
  message.parts
    .filter((part) => part.type === "text")
    .map((part) => part.text)
    .join("");

const createRealtimeTextMessage = (
  update: RealtimeTranscriptUpdate,
  anchorMessageId: string | null,
  voiceOrder: number,
): MayaUIMessage => ({
  id: update.id,
  metadata: {
    source: "voice",
    voiceAnchorMessageId: anchorMessageId,
    voiceOrder,
  },
  parts: [{ type: "text", text: update.text }],
  role: update.role,
});

export const applyRealtimeTranscriptUpdate = (
  messages: MayaUIMessage[],
  update: RealtimeTranscriptUpdate,
  anchorMessageId: string | null,
  voiceOrder: number,
) => {
  if (!update.text) {
    return messages;
  }

  const existingIndex = messages.findIndex(
    (message) => message.id === update.id,
  );
  if (existingIndex === -1) {
    return [
      ...messages,
      createRealtimeTextMessage(update, anchorMessageId, voiceOrder),
    ];
  }

  return messages.map((message, index) => {
    if (index !== existingIndex) {
      return message;
    }

    const currentText = getMayaMessageText(message);
    const text =
      update.mode === "append" ? `${currentText}${update.text}` : update.text;

    const updatedMessage: MayaUIMessage = {
      ...message,
      parts: [{ type: "text", text }],
    };
    return updatedMessage;
  });
};

export const getRealtimeTranscriptUpdate = (
  event: RealtimeTranscriptEvent,
): RealtimeTranscriptUpdate | null => {
  switch (event.type) {
    case "conversation.item.input_audio_transcription.delta":
      return event.item_id && event.delta
        ? {
            id: `voice-user-${event.item_id}`,
            mode: "append",
            role: "user",
            text: event.delta,
          }
        : null;
    case "conversation.item.input_audio_transcription.completed":
      return event.item_id && event.transcript
        ? {
            id: `voice-user-${event.item_id}`,
            mode: "replace",
            role: "user",
            text: event.transcript,
          }
        : null;
    case "response.output_audio_transcript.delta":
    case "response.audio_transcript.delta": {
      const id = event.item_id ?? event.response_id;
      return id && event.delta
        ? {
            id: `voice-assistant-${id}`,
            mode: "append",
            role: "assistant",
            text: event.delta,
          }
        : null;
    }
    case "response.output_audio_transcript.done":
    case "response.audio_transcript.done": {
      const id = event.item_id ?? event.response_id;
      return id && event.transcript
        ? {
            id: `voice-assistant-${id}`,
            mode: "replace",
            role: "assistant",
            text: event.transcript,
          }
        : null;
    }
    default:
      return null;
  }
};

export const mergeRealtimeVoiceMessages = (
  conversationMessages: MayaUIMessage[],
  voiceMessages: MayaUIMessage[],
) => {
  if (voiceMessages.length === 0) {
    return conversationMessages;
  }

  const voiceByAnchor = new Map<string | null, MayaUIMessage[]>();
  for (const message of voiceMessages) {
    const anchor = message.metadata?.voiceAnchorMessageId ?? null;
    const group = voiceByAnchor.get(anchor);
    if (group) {
      group.push(message);
    } else {
      voiceByAnchor.set(anchor, [message]);
    }
  }

  const sortByVoiceOrder = (messages: MayaUIMessage[]) =>
    messages.toSorted(
      (a, b) => (a.metadata?.voiceOrder ?? 0) - (b.metadata?.voiceOrder ?? 0),
    );

  const merged = sortByVoiceOrder(voiceByAnchor.get(null) ?? []);
  const matchedAnchors = new Set<string | null>([null]);

  for (const message of conversationMessages) {
    merged.push(message);
    const anchoredVoiceMessages = voiceByAnchor.get(message.id);
    if (anchoredVoiceMessages) {
      merged.push(...sortByVoiceOrder(anchoredVoiceMessages));
      matchedAnchors.add(message.id);
    }
  }

  voiceByAnchor.forEach((messages, anchor) => {
    if (!matchedAnchors.has(anchor)) {
      merged.push(...sortByVoiceOrder(messages));
    }
  });

  return merged;
};
