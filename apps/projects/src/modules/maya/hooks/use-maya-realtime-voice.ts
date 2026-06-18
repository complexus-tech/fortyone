"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getApiUrl } from "@/lib/api-url";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";

type RealtimeVoiceStatus =
  | "idle"
  | "connecting"
  | "connected"
  | "disconnecting";

type ApiResponse<T> = {
  data?: T;
  error?: {
    message?: string;
  };
};

type RealtimeSessionResponse = {
  clientSecret: string;
  expiresAt?: number;
  model: string;
  voice: string;
};

type RealtimeFunctionCall = {
  arguments?: string;
  call_id?: string;
  name?: string;
  type?: string;
};

type RealtimeServerEvent = {
  response?: {
    output?: RealtimeFunctionCall[];
  };
  type?: string;
};

type RealtimeToolOutput = {
  success?: boolean;
  error?: string;
};

const REALTIME_CALLS_URL = "https://api.openai.com/v1/realtime/calls";

const isBrowserRealtimeSupported = () => {
  if (
    typeof window === "undefined" ||
    typeof RTCPeerConnection === "undefined" ||
    !("mediaDevices" in navigator)
  ) {
    return false;
  }

  return "getUserMedia" in navigator.mediaDevices;
};

const errorMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }
  return fallback;
};

const parseRealtimeToolOutput = async (response: Response) => {
  const payload = (await response
    .json()
    .catch(() => null)) as ApiResponse<RealtimeToolOutput> | null;

  if (!response.ok) {
    return {
      success: false,
      error: payload?.error?.message ?? "Tool execution failed.",
    };
  }

  return (
    payload?.data ?? {
      success: false,
      error: "Tool returned an unreadable response.",
    }
  );
};

export const useMayaRealtimeVoice = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
  const dataChannelRef = useRef<RTCDataChannel | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const remoteAudioRef = useRef<HTMLAudioElement | null>(null);
  const handledFunctionCallsRef = useRef<Set<string>>(new Set());
  const speakingTimeoutRef = useRef<number | null>(null);
  const [status, setStatus] = useState<RealtimeVoiceStatus>("idle");
  const [isSpeaking, setIsSpeaking] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const clearSpeakingTimer = useCallback(() => {
    if (speakingTimeoutRef.current) {
      window.clearTimeout(speakingTimeoutRef.current);
      speakingTimeoutRef.current = null;
    }
  }, []);

  const closeConnection = useCallback(() => {
    clearSpeakingTimer();
    setIsSpeaking(false);

    dataChannelRef.current?.close();
    dataChannelRef.current = null;
    handledFunctionCallsRef.current.clear();

    peerConnectionRef.current?.close();
    peerConnectionRef.current = null;

    localStreamRef.current?.getTracks().forEach((track) => {
      track.stop();
    });
    localStreamRef.current = null;

    if (remoteAudioRef.current) {
      remoteAudioRef.current.srcObject = null;
      remoteAudioRef.current.remove();
      remoteAudioRef.current = null;
    }
  }, [clearSpeakingTimer]);

  const markSpeaking = useCallback(() => {
    setIsSpeaking(true);
    clearSpeakingTimer();
    speakingTimeoutRef.current = window.setTimeout(() => {
      setIsSpeaking(false);
      speakingTimeoutRef.current = null;
    }, 900);
  }, [clearSpeakingTimer]);

  const runRealtimeTool = useCallback(
    async (functionCall: RealtimeFunctionCall) => {
      const callId = functionCall.call_id;
      const name = functionCall.name;
      if (!callId || !name || handledFunctionCallsRef.current.has(callId)) {
        return;
      }
      handledFunctionCallsRef.current.add(callId);

      let toolArguments: unknown = {};
      if (functionCall.arguments?.trim()) {
        try {
          toolArguments = JSON.parse(functionCall.arguments) as unknown;
        } catch {
          toolArguments = {};
        }
      }

      const output = await fetch(
        `${getApiUrl()}/workspaces/${workspaceSlug}/maya/realtime-tool`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            arguments: toolArguments,
            name,
          }),
        },
      )
        .then(parseRealtimeToolOutput)
        .catch((toolError: unknown) => ({
          success: false,
          error: errorMessage(toolError, "Tool execution failed."),
        }));
      if (name === "create_task" && output.success === true) {
        queryClient.invalidateQueries({
          queryKey: storyKeys.all(workspaceSlug),
        });
      }

      const dataChannel = dataChannelRef.current;
      if (!dataChannel || dataChannel.readyState !== "open") {
        return;
      }

      dataChannel.send(
        JSON.stringify({
          type: "conversation.item.create",
          item: {
            type: "function_call_output",
            call_id: callId,
            output: JSON.stringify(output),
          },
        }),
      );
      dataChannel.send(JSON.stringify({ type: "response.create" }));
    },
    [queryClient, workspaceSlug],
  );

  const handleFunctionCalls = useCallback(
    (event: RealtimeServerEvent) => {
      const calls =
        event.response?.output?.filter(
          (item) => item.type === "function_call",
        ) ?? [];

      calls.forEach((functionCall) => {
        void runRealtimeTool(functionCall);
      });
    },
    [runRealtimeTool],
  );

  const handleRealtimeEvent = useCallback(
    (data: string) => {
      try {
        const event = JSON.parse(data) as RealtimeServerEvent;
        switch (event.type) {
          case "response.audio.delta":
          case "response.audio_transcript.delta":
            markSpeaking();
            break;
          case "response.audio.done":
            clearSpeakingTimer();
            setIsSpeaking(false);
            break;
          case "response.done":
            clearSpeakingTimer();
            setIsSpeaking(false);
            handleFunctionCalls(event);
            break;
          default:
            break;
        }
      } catch {
        // Realtime data channel messages are best-effort UI signals here.
      }
    },
    [clearSpeakingTimer, handleFunctionCalls, markSpeaking],
  );

  const createRealtimeSession = useCallback(async () => {
    const response = await fetch(
      `${getApiUrl()}/workspaces/${workspaceSlug}/maya/realtime-session`,
      {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
      },
    );
    const payload = (await response
      .json()
      .catch(() => null)) as ApiResponse<RealtimeSessionResponse> | null;

    if (!response.ok) {
      throw new Error(
        payload?.error?.message ?? "Failed to create voice session.",
      );
    }
    if (!payload?.data?.clientSecret) {
      throw new Error("Voice session did not include a client secret.");
    }
    return payload.data;
  }, [workspaceSlug]);

  const disconnect = useCallback(() => {
    if (status === "idle") return;
    setStatus("disconnecting");
    closeConnection();
    setStatus("idle");
  }, [closeConnection, status]);

  const connect = useCallback(async () => {
    if (status === "connecting" || status === "connected") return;
    setError(null);

    if (!isBrowserRealtimeSupported()) {
      setError("Realtime voice is not supported in this browser.");
      return;
    }

    setStatus("connecting");

    try {
      const session = await createRealtimeSession();
      const peerConnection = new RTCPeerConnection();
      peerConnectionRef.current = peerConnection;

      const remoteAudio = document.createElement("audio");
      remoteAudio.autoplay = true;
      remoteAudio.setAttribute("playsinline", "true");
      remoteAudioRef.current = remoteAudio;

      peerConnection.ontrack = (event) => {
        remoteAudio.srcObject = event.streams[0] ?? null;
      };
      peerConnection.onconnectionstatechange = () => {
        if (
          peerConnection.connectionState === "failed" ||
          peerConnection.connectionState === "closed"
        ) {
          closeConnection();
          setStatus("idle");
        }
      };

      const localStream = await navigator.mediaDevices.getUserMedia({
        audio: {
          autoGainControl: true,
          echoCancellation: true,
          noiseSuppression: true,
        },
      });
      localStreamRef.current = localStream;
      localStream.getAudioTracks().forEach((track) => {
        peerConnection.addTrack(track, localStream);
      });

      const dataChannel = peerConnection.createDataChannel("oai-events");
      dataChannelRef.current = dataChannel;
      dataChannel.onmessage = (event) => {
        if (typeof event.data === "string") {
          handleRealtimeEvent(event.data);
        }
      };

      const offer = await peerConnection.createOffer();
      await peerConnection.setLocalDescription(offer);

      const answerResponse = await fetch(REALTIME_CALLS_URL, {
        method: "POST",
        body: offer.sdp ?? "",
        headers: {
          Authorization: `Bearer ${session.clientSecret}`,
          "Content-Type": "application/sdp",
        },
      });
      if (!answerResponse.ok) {
        throw new Error("Failed to connect voice session.");
      }

      await peerConnection.setRemoteDescription({
        type: "answer",
        sdp: await answerResponse.text(),
      });

      setStatus("connected");
    } catch (connectError) {
      closeConnection();
      setStatus("idle");
      setError(errorMessage(connectError, "Failed to start voice session."));
    }
  }, [closeConnection, createRealtimeSession, handleRealtimeEvent, status]);

  useEffect(() => {
    return () => {
      closeConnection();
    };
  }, [closeConnection]);

  return {
    connect,
    disconnect,
    error,
    isListening: status === "connected" && !isSpeaking,
    isSpeaking,
    status,
  };
};
