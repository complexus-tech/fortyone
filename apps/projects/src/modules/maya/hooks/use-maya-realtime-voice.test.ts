/* global describe, expect, it -- Jest provides these test globals. */

import { act, cleanup, renderHook } from "@testing-library/react";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { storyKeys } from "@/modules/stories/constants";
import {
  applyRealtimeTranscriptUpdate,
  getRealtimeTranscriptUpdate,
  mergeRealtimeVoiceMessages,
} from "../utils/realtime-voice-messages";
import { createRealtimePeerConnection } from "../utils/realtime-peer-connection";
import { useMayaRealtimeVoice } from "./use-maya-realtime-voice";
import {
  GOODBYE_DISCONNECT_DELAY_MS,
  REALTIME_CONNECTION_TIMEOUT_MS,
} from "./maya-realtime-voice-contract";

const mockInvalidateQueries = jest.fn();
jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ invalidateQueries: mockInvalidateQueries }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/api-url", () => ({ getApiUrl: () => "https://api.test" }));
jest.mock("../utils/realtime-peer-connection", () => ({
  createRealtimePeerConnection: jest.fn(),
  disposeRealtimePeerConnection: (peer: RTCPeerConnection | null) => {
    peer?.close();
  },
}));

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

describe("Maya realtime voice connection lifecycle", () => {
  const getUserMedia = jest.fn();
  const fetchMock = jest.fn();
  const stopTrack = jest.fn();
  const closePeer = jest.fn();
  const closeChannel = jest.fn();
  const playAudio = jest.fn<Promise<void>, []>();
  const navigate = jest.fn();
  let track: MediaStreamTrack;
  let stream: MediaStream;
  let channel: RTCDataChannel;
  let peer: RTCPeerConnection;
  const show = () =>
    renderHook(() =>
      useMayaRealtimeVoice({
        conversationMessages: [],
        currentPath: "/acme/maya",
        navigate,
        setApplicationTheme: jest.fn(),
      }),
    );

  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
    Object.defineProperty(navigator, "mediaDevices", {
      configurable: true,
      value: { getUserMedia },
    });
    Object.defineProperty(globalThis, "RTCPeerConnection", {
      configurable: true,
      value: jest.fn(),
      writable: true,
    });
    Object.defineProperty(globalThis, "fetch", {
      configurable: true,
      value: fetchMock,
      writable: true,
    });
    track = { onended: null, stop: stopTrack } as unknown as MediaStreamTrack;
    stream = {
      getTracks: () => [track],
      getAudioTracks: () => [track],
    } as unknown as MediaStream;
    channel = {
      close: closeChannel,
      readyState: "open",
      send: jest.fn(),
    } as unknown as RTCDataChannel;
    peer = {
      addTrack: jest.fn(),
      close: closePeer,
      connectionState: "new",
      createDataChannel: () => channel,
      createOffer: async () => ({ sdp: "offer" }),
      setLocalDescription: async () => undefined,
      setRemoteDescription: async () => undefined,
    } as unknown as RTCPeerConnection;
    jest.mocked(createRealtimePeerConnection).mockReturnValue(peer);
    getUserMedia.mockResolvedValue(stream);
    fetchMock.mockImplementation(async (url: string) => ({
      ok: true,
      json: async () => ({
        data: {
          clientSecret: "test-secret",
          sessionId: "session-1",
          maxSessionSeconds: 300,
        },
      }),
      text: async () => (url.includes("realtime/calls") ? "answer" : ""),
    }));
    playAudio.mockResolvedValue(undefined);
    jest
      .spyOn(HTMLMediaElement.prototype, "play")
      .mockImplementation(playAudio);
    jest
      .spyOn(HTMLMediaElement.prototype, "pause")
      .mockImplementation(() => undefined);
  });

  afterEach(() => {
    cleanup();
    delete window.__FORTYONE_OFFLINE_PAGE__;
    jest.restoreAllMocks();
    jest.useRealTimers();
  });

  it("does not request the microphone or a session on a restored offline page", async () => {
    window.__FORTYONE_OFFLINE_PAGE__ = true;
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    expect(result.current.error).toContain("internet connection");
    expect(result.current.status).toBe("idle");
    expect(getUserMedia).not.toHaveBeenCalled();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("starts one microphone request when connect is called twice before a render", async () => {
    const { result } = show();
    await act(async () => {
      await Promise.all([result.current.connect(), result.current.connect()]);
    });
    expect(getUserMedia).toHaveBeenCalledTimes(1);
    expect(createRealtimePeerConnection).toHaveBeenCalledTimes(1);
    act(() => {
      channel.onopen?.(new Event("open"));
    });
    expect(result.current.status).toBe("connected");
  });

  it("releases a delayed microphone grant after cancelling before a render", async () => {
    let resolveStream!: (value: MediaStream) => void;
    getUserMedia.mockReturnValue(
      new Promise((resolve) => {
        resolveStream = resolve;
      }),
    );
    const { result } = show();
    await act(async () => {
      const pending = result.current.connect();
      result.current.disconnect();
      resolveStream(stream);
      await pending;
    });
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(fetchMock).not.toHaveBeenCalled();
    expect(result.current.status).toBe("idle");
  });

  it("explains microphone permission errors and allows retry", async () => {
    getUserMedia.mockRejectedValueOnce(
      new DOMException("Permission denied", "NotAllowedError"),
    );
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    expect(result.current.error).toMatch(/microphone/i);
    expect(result.current.error).toMatch(/settings/i);
    expect(result.current.status).toBe("idle");
    await act(async () => {
      await result.current.connect();
    });
    expect(result.current.error).toBeNull();
    expect(result.current.status).toBe("connecting");
  });

  it("releases the microphone and explains a failed peer connection", async () => {
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    Object.defineProperty(peer, "connectionState", { value: "failed" });
    act(() => {
      peer.onconnectionstatechange?.(new Event("connectionstatechange"));
    });
    expect(result.current.error).toContain("disconnected");
    expect(result.current.status).toBe("idle");
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(closePeer).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("realtime-session/end"),
      expect.anything(),
    );
  });

  it("ends the session when desktop service connectivity goes offline", async () => {
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    act(() => {
      window.__FORTYONE_OFFLINE_PAGE__ = true;
      window.dispatchEvent(new Event("fortyone:desktop-connectivity"));
    });
    expect(result.current.error).toContain("offline");
    expect(result.current.status).toBe("idle");
    expect(stopTrack).toHaveBeenCalledTimes(1);
  });

  it("ends a call with feedback if its microphone disconnects", async () => {
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    act(() => {
      track.onended?.(new Event("ended"));
    });
    expect(result.current.error).toContain("microphone disconnected");
    expect(result.current.status).toBe("idle");
    expect(closeChannel).toHaveBeenCalledTimes(1);
    expect(track.onended).toBeNull();
  });

  it("releases the microphone when remote audio cannot play", async () => {
    playAudio.mockRejectedValueOnce(
      new DOMException("Playback denied", "NotAllowedError"),
    );
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    await act(async () => {
      peer.ontrack?.({
        streams: [stream],
      } as unknown as RTCTrackEvent);
    });
    expect(result.current.error).toContain("audio could not play");
    expect(result.current.status).toBe("idle");
    expect(stopTrack).toHaveBeenCalledTimes(1);
  });

  it("times out a connection that never opens and releases the microphone", async () => {
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    act(() => {
      jest.advanceTimersByTime(REALTIME_CONNECTION_TIMEOUT_MS);
    });
    expect(result.current.error).toContain("could not connect");
    expect(result.current.status).toBe("idle");
    expect(stopTrack).toHaveBeenCalledTimes(1);
  });

  it("keeps an opened session active after the connection timeout", async () => {
    const { result } = show();
    await act(async () => {
      await result.current.connect();
    });
    act(() => {
      channel.onopen?.(new Event("open"));
      jest.advanceTimersByTime(REALTIME_CONNECTION_TIMEOUT_MS);
    });
    expect(result.current.status).toBe("connected");
    expect(result.current.error).toBeNull();
    expect(stopTrack).not.toHaveBeenCalled();
  });

  it.each(["update_story", "end_conversation"])(
    "ignores a late %s result after reconnecting while refreshing completed writes",
    async (toolName) => {
      const { result } = show();
      await act(async () => {
        await result.current.connect();
      });
      let finishTool!: (response: unknown) => void;
      fetchMock.mockReturnValueOnce(
        new Promise((resolve) => {
          finishTool = resolve;
        }),
      );
      act(() => {
        channel.onopen?.(new Event("open"));
        channel.onmessage?.(
          new MessageEvent("message", {
            data: JSON.stringify({
              type: "response.done",
              response: {
                output: [
                  {
                    type: "function_call",
                    call_id: "old-call",
                    name: toolName,
                  },
                ],
              },
            }),
          }),
        );
        result.current.disconnect();
      });

      const nextSessionSend = jest.fn();
      channel = { ...channel, send: nextSessionSend };
      await act(async () => {
        await result.current.connect();
        channel.onopen?.(new Event("open"));
      });
      await act(async () => {
        finishTool({
          ok: true,
          json: async () => ({
            data: {
              success: true,
              clientAction: { type: "navigate", path: "/acme/old-task" },
            },
          }),
        });
      });
      act(() => {
        channel.onmessage?.(
          new MessageEvent("message", {
            data: JSON.stringify({ type: "response.output_audio.done" }),
          }),
        );
        jest.advanceTimersByTime(GOODBYE_DISCONNECT_DELAY_MS);
      });

      expect(nextSessionSend).toHaveBeenCalledTimes(1);
      expect(navigate).not.toHaveBeenCalled();
      expect(result.current.status).toBe("connected");
      if (toolName === "update_story") {
        expect(mockInvalidateQueries).toHaveBeenCalledWith({
          queryKey: storyKeys.all("acme"),
        });
      }
    },
  );
});
