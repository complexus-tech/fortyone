import { useCallback, useRef } from "react";
import type { RealtimeServerEvent } from "./maya-realtime-voice-contract";

/**
 * Maintains stable ordering for partial voice transcripts, even when Realtime
 * events arrive out of the order in which their corresponding messages render.
 */
export const useRealtimeVoiceMessageOrder = () => {
  const messageOrdersRef = useRef<Map<string, number>>(new Map());
  const nextOrderRef = useRef(0);

  const getMessageOrder = useCallback((messageId: string) => {
    const existingOrder = messageOrdersRef.current.get(messageId);
    if (existingOrder !== undefined) {
      return existingOrder;
    }

    const nextOrder = nextOrderRef.current;
    nextOrderRef.current += 1;
    messageOrdersRef.current.set(messageId, nextOrder);
    return nextOrder;
  }, []);

  const rememberEventItemOrder = useCallback(
    (event: RealtimeServerEvent) => {
      if (event.type === "input_audio_buffer.speech_started" && event.item_id) {
        getMessageOrder(`voice-user-${event.item_id}`);
        return;
      }

      if (
        (event.type === "conversation.item.added" ||
          event.type === "conversation.item.created") &&
        event.item?.id
      ) {
        const role = event.item.role === "assistant" ? "assistant" : "user";
        getMessageOrder(`voice-${role}-${event.item.id}`);
        return;
      }

      if (event.type === "response.output_item.added" && event.item?.id) {
        getMessageOrder(`voice-assistant-${event.item.id}`);
      }
    },
    [getMessageOrder],
  );

  const resetMessageOrders = useCallback(() => {
    messageOrdersRef.current.clear();
    nextOrderRef.current = 0;
  }, []);

  return { getMessageOrder, rememberEventItemOrder, resetMessageOrders };
};
