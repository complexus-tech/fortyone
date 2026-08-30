import { act, renderHook } from "@testing-library/react";
import { useRealtimeVoiceMessageOrder } from "./use-realtime-voice-message-order";

describe("useRealtimeVoiceMessageOrder", () => {
  it("assigns stable chronological positions to event-derived transcript IDs", () => {
    const { result } = renderHook(() => useRealtimeVoiceMessageOrder());

    act(() => {
      result.current.rememberEventItemOrder({
        item_id: "audio-1",
        type: "input_audio_buffer.speech_started",
      });
      result.current.rememberEventItemOrder({
        item: { id: "response-1" },
        type: "response.output_item.added",
      });
    });

    expect(result.current.getMessageOrder("voice-user-audio-1")).toBe(0);
    expect(result.current.getMessageOrder("voice-assistant-response-1")).toBe(
      1,
    );
    expect(result.current.getMessageOrder("voice-user-audio-1")).toBe(0);
  });

  it("resets its ordering state with a new voice conversation", () => {
    const { result } = renderHook(() => useRealtimeVoiceMessageOrder());

    expect(result.current.getMessageOrder("voice-user-first")).toBe(0);
    expect(result.current.getMessageOrder("voice-assistant-first")).toBe(1);

    act(() => {
      result.current.resetMessageOrders();
    });

    expect(result.current.getMessageOrder("voice-user-next")).toBe(0);
  });
});
