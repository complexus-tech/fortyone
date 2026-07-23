/* global describe, expect, it -- Jest provides these test globals. */

import type { MayaUIMessage } from "@/lib/ai/tools/types";
import {
  applyRealtimeTranscriptUpdate,
  getRealtimeTranscriptUpdate,
  mergeRealtimeVoiceMessages,
} from "../utils/realtime-voice-messages";

const textMessage = (
  id: string,
  role: "assistant" | "user",
  text: string,
): MayaUIMessage => ({
  id,
  parts: [{ type: "text", text }],
  role,
});

describe("Maya realtime voice transcript state", () => {
  it("maps current GA input and output transcript events", () => {
    expect(
      getRealtimeTranscriptUpdate({
        item_id: "user-item",
        transcript: "Show my tasks",
        type: "conversation.item.input_audio_transcription.completed",
      }),
    ).toEqual({
      id: "voice-user-user-item",
      mode: "replace",
      role: "user",
      text: "Show my tasks",
    });

    expect(
      getRealtimeTranscriptUpdate({
        delta: "You have ",
        item_id: "assistant-item",
        type: "response.output_audio_transcript.delta",
      }),
    ).toEqual({
      id: "voice-assistant-assistant-item",
      mode: "append",
      role: "assistant",
      text: "You have ",
    });
  });

  it("streams transcript deltas and replaces them with the final transcript", () => {
    const partial = applyRealtimeTranscriptUpdate(
      [],
      {
        id: "voice-assistant-item-1",
        mode: "append",
        role: "assistant",
        text: "You have ",
      },
      "typed-1",
      1,
    );
    const streamed = applyRealtimeTranscriptUpdate(
      partial,
      {
        id: "voice-assistant-item-1",
        mode: "append",
        role: "assistant",
        text: "three tasks.",
      },
      "typed-1",
      1,
    );
    const completed = applyRealtimeTranscriptUpdate(
      streamed,
      {
        id: "voice-assistant-item-1",
        mode: "replace",
        role: "assistant",
        text: "You have three tasks.",
      },
      "typed-1",
      1,
    );

    expect(completed).toEqual([
      {
        id: "voice-assistant-item-1",
        metadata: {
          source: "voice",
          voiceAnchorMessageId: "typed-1",
          voiceOrder: 1,
        },
        parts: [{ type: "text", text: "You have three tasks." }],
        role: "assistant",
      },
    ]);
  });

  it("keeps voice turns anchored between earlier and later typed messages", () => {
    const typedBefore = textMessage("typed-1", "user", "What is due?");
    const typedAfter = textMessage("typed-2", "user", "Create a follow-up.");
    const assistantVoice = {
      ...textMessage(
        "voice-assistant-item-1",
        "assistant",
        "You have one item due today.",
      ),
      metadata: {
        source: "voice" as const,
        voiceAnchorMessageId: "typed-1",
        voiceOrder: 1,
      },
    };
    const userVoice = {
      ...textMessage("voice-user-item-1", "user", "What is due today?"),
      metadata: {
        source: "voice" as const,
        voiceAnchorMessageId: "typed-1",
        voiceOrder: 0,
      },
    };

    expect(
      mergeRealtimeVoiceMessages(
        [typedBefore, typedAfter],
        [assistantVoice, userVoice],
      ).map((message) => message.id),
    ).toEqual([
      "typed-1",
      "voice-user-item-1",
      "voice-assistant-item-1",
      "typed-2",
    ]);
  });
});
