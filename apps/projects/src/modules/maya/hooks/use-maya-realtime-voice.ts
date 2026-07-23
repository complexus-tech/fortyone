"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getApiUrl } from "@/lib/api-url";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import {
  applyRealtimeTranscriptUpdate,
  getMayaMessageText,
  getRealtimeTranscriptUpdate,
  mergeRealtimeVoiceMessages,
} from "../utils/realtime-voice-messages";

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
  maxSessionSeconds: number;
  model: string;
  monthlyLimitSeconds: number;
  remainingSeconds: number;
  sessionId: string;
  voice: string;
};

type RealtimeFunctionCall = {
  arguments?: string;
  call_id?: string;
  name?: string;
  type?: string;
};

type RealtimeServerEvent = {
  delta?: string;
  error?: {
    message?: string;
  };
  item?: {
    id?: string;
    role?: string;
    type?: string;
  };
  item_id?: string;
  response?: {
    output?: RealtimeFunctionCall[];
  };
  response_id?: string;
  transcript?: string;
  type?: string;
};

type RealtimeToolOutput = {
  success?: boolean;
  error?: string;
};

type UseMayaRealtimeVoiceOptions = {
  conversationMessages: MayaUIMessage[];
  currentPath: string;
};

const REALTIME_CALLS_URL = "https://api.openai.com/v1/realtime/calls";
const FALLBACK_REALTIME_MAX_SESSION_SECONDS = 5 * 60;
const REALTIME_IDLE_TIMEOUT_MS = 60_000;
const GOODBYE_DISCONNECT_DELAY_MS = 800;
const MAX_REALTIME_CONTEXT_MESSAGES = 24;
const REALTIME_ACTIVITY_EVENTS = new Set([
  "conversation.item.added",
  "conversation.item.created",
  "conversation.item.input_audio_transcription.completed",
  "conversation.item.input_audio_transcription.delta",
  "input_audio_buffer.speech_started",
  "input_audio_buffer.speech_stopped",
  "response.audio.delta",
  "response.audio.done",
  "response.audio_transcript.delta",
  "response.audio_transcript.done",
  "response.created",
  "response.done",
  "response.output_audio.delta",
  "response.output_audio.done",
  "response.output_audio_transcript.delta",
  "response.output_audio_transcript.done",
  "response.output_item.added",
]);

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

export const useMayaRealtimeVoice = ({
  conversationMessages,
  currentPath,
}: UseMayaRealtimeVoiceOptions) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
  const connectionAttemptRef = useRef(0);
  const connectionAbortControllerRef = useRef<AbortController | null>(null);
  const dataChannelRef = useRef<RTCDataChannel | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const remoteAudioRef = useRef<HTMLAudioElement | null>(null);
  const handledFunctionCallsRef = useRef<Set<string>>(new Set());
  const speakingTimeoutRef = useRef<number | null>(null);
  const idleTimeoutRef = useRef<number | null>(null);
  const sessionTimeoutRef = useRef<number | null>(null);
  const countdownIntervalRef = useRef<number | null>(null);
  const goodbyeTimeoutRef = useRef<number | null>(null);
  const sessionEndsAtRef = useRef<number | null>(null);
  const activeSessionIdRef = useRef<string | null>(null);
  const conversationMessagesRef = useRef(conversationMessages);
  const messagesRef = useRef<MayaUIMessage[]>([]);
  const voiceAnchorMessageIdRef = useRef<string | null>(null);
  const voiceMessageOrdersRef = useRef<Map<string, number>>(new Map());
  const nextVoiceMessageOrderRef = useRef(0);
  const [status, setStatus] = useState<RealtimeVoiceStatus>("idle");
  const [isSpeaking, setIsSpeaking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [messages, setMessages] = useState<MayaUIMessage[]>([]);
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);

  useEffect(() => {
    conversationMessagesRef.current = conversationMessages;
  }, [conversationMessages]);

  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  const clearSpeakingTimer = useCallback(() => {
    if (speakingTimeoutRef.current) {
      window.clearTimeout(speakingTimeoutRef.current);
      speakingTimeoutRef.current = null;
    }
  }, []);

  const clearIdleTimer = useCallback(() => {
    if (idleTimeoutRef.current) {
      window.clearTimeout(idleTimeoutRef.current);
      idleTimeoutRef.current = null;
    }
  }, []);

  const clearSessionTimers = useCallback(() => {
    if (sessionTimeoutRef.current) {
      window.clearTimeout(sessionTimeoutRef.current);
      sessionTimeoutRef.current = null;
    }
    if (countdownIntervalRef.current) {
      window.clearInterval(countdownIntervalRef.current);
      countdownIntervalRef.current = null;
    }
    sessionEndsAtRef.current = null;
    setRemainingSeconds(null);
  }, []);

  const endRealtimeSession = useCallback(
    (sessionId: string) => {
      void fetch(
        `${getApiUrl()}/workspaces/${workspaceSlug}/maya/realtime-session/end`,
        {
          body: JSON.stringify({ sessionId }),
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          keepalive: true,
          method: "POST",
        },
      ).catch(() => {
        // Session end reporting is best-effort; the backend caps open sessions.
      });
    },
    [workspaceSlug],
  );

  const closeConnection = useCallback(() => {
    connectionAttemptRef.current += 1;
    connectionAbortControllerRef.current?.abort();
    connectionAbortControllerRef.current = null;
    clearSpeakingTimer();
    clearIdleTimer();
    clearSessionTimers();
    setIsSpeaking(false);

    if (goodbyeTimeoutRef.current) {
      window.clearTimeout(goodbyeTimeoutRef.current);
      goodbyeTimeoutRef.current = null;
    }

    const activeSessionId = activeSessionIdRef.current;
    activeSessionIdRef.current = null;
    if (activeSessionId) {
      endRealtimeSession(activeSessionId);
    }

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
      remoteAudioRef.current.pause();
      remoteAudioRef.current.srcObject = null;
      remoteAudioRef.current.remove();
      remoteAudioRef.current = null;
    }
  }, [
    clearIdleTimer,
    clearSessionTimers,
    clearSpeakingTimer,
    endRealtimeSession,
  ]);

  const resetIdleTimer = useCallback(() => {
    clearIdleTimer();
    idleTimeoutRef.current = window.setTimeout(() => {
      closeConnection();
      setStatus("idle");
    }, REALTIME_IDLE_TIMEOUT_MS);
  }, [clearIdleTimer, closeConnection]);

  const startSessionTimer = useCallback(
    (maxSessionSeconds: number) => {
      clearSessionTimers();
      const seconds =
        maxSessionSeconds > 0
          ? maxSessionSeconds
          : FALLBACK_REALTIME_MAX_SESSION_SECONDS;
      sessionEndsAtRef.current = Date.now() + seconds * 1000;
      setRemainingSeconds(seconds);

      countdownIntervalRef.current = window.setInterval(() => {
        const endsAt = sessionEndsAtRef.current;
        if (!endsAt) {
          return;
        }
        setRemainingSeconds(
          Math.max(0, Math.ceil((endsAt - Date.now()) / 1000)),
        );
      }, 1_000);

      sessionTimeoutRef.current = window.setTimeout(() => {
        closeConnection();
        setStatus("idle");
      }, seconds * 1000);
    },
    [clearSessionTimers, closeConnection],
  );

  const markSpeaking = useCallback(() => {
    setIsSpeaking(true);
    clearSpeakingTimer();
    speakingTimeoutRef.current = window.setTimeout(() => {
      setIsSpeaking(false);
      speakingTimeoutRef.current = null;
    }, 900);
  }, [clearSpeakingTimer]);

  const getVoiceMessageOrder = useCallback((messageId: string) => {
    const existingOrder = voiceMessageOrdersRef.current.get(messageId);
    if (existingOrder !== undefined) {
      return existingOrder;
    }

    const nextOrder = nextVoiceMessageOrderRef.current;
    nextVoiceMessageOrderRef.current += 1;
    voiceMessageOrdersRef.current.set(messageId, nextOrder);
    return nextOrder;
  }, []);

  const rememberEventItemOrder = useCallback(
    (event: RealtimeServerEvent) => {
      if (event.type === "input_audio_buffer.speech_started" && event.item_id) {
        getVoiceMessageOrder(`voice-user-${event.item_id}`);
        return;
      }

      if (
        (event.type === "conversation.item.added" ||
          event.type === "conversation.item.created") &&
        event.item?.id
      ) {
        const role = event.item.role === "assistant" ? "assistant" : "user";
        getVoiceMessageOrder(`voice-${role}-${event.item.id}`);
        return;
      }

      if (event.type === "response.output_item.added" && event.item?.id) {
        getVoiceMessageOrder(`voice-assistant-${event.item.id}`);
      }
    },
    [getVoiceMessageOrder],
  );

  const runRealtimeTool = useCallback(
    async (functionCall: RealtimeFunctionCall) => {
      resetIdleTimer();

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
      resetIdleTimer();

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
      if (name === "end_conversation") {
        setStatus("disconnecting");
        goodbyeTimeoutRef.current = window.setTimeout(() => {
          closeConnection();
          setStatus("idle");
        }, GOODBYE_DISCONNECT_DELAY_MS);
        return;
      }
      dataChannel.send(JSON.stringify({ type: "response.create" }));
    },
    [closeConnection, queryClient, resetIdleTimer, workspaceSlug],
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
      let event: RealtimeServerEvent;
      try {
        event = JSON.parse(data) as RealtimeServerEvent;
      } catch {
        return;
      }

      if (event.type && REALTIME_ACTIVITY_EVENTS.has(event.type)) {
        resetIdleTimer();
      }
      rememberEventItemOrder(event);

      const transcriptUpdate = getRealtimeTranscriptUpdate(event);
      if (transcriptUpdate) {
        const order = getVoiceMessageOrder(transcriptUpdate.id);
        setMessages((currentMessages) =>
          applyRealtimeTranscriptUpdate(
            currentMessages,
            transcriptUpdate,
            voiceAnchorMessageIdRef.current,
            order,
          ),
        );
      }

      switch (event.type) {
        case "response.output_audio.delta":
        case "response.audio.delta":
        case "response.output_audio_transcript.delta":
        case "response.audio_transcript.delta":
          markSpeaking();
          break;
        case "response.output_audio.done":
        case "response.audio.done":
          clearSpeakingTimer();
          setIsSpeaking(false);
          break;
        case "response.done":
          clearSpeakingTimer();
          setIsSpeaking(false);
          handleFunctionCalls(event);
          break;
        case "error":
          setError(
            event.error?.message ?? "The voice session encountered an error.",
          );
          break;
        default:
          break;
      }
    },
    [
      clearSpeakingTimer,
      getVoiceMessageOrder,
      handleFunctionCalls,
      markSpeaking,
      rememberEventItemOrder,
      resetIdleTimer,
    ],
  );

  const createRealtimeSession = useCallback(
    async (signal: AbortSignal) => {
      const contextMessages = mergeRealtimeVoiceMessages(
        conversationMessagesRef.current,
        messagesRef.current,
      )
        .map((message) => ({
          role: message.role,
          text: getMayaMessageText(message).trim(),
        }))
        .filter(
          (
            message,
          ): message is {
            role: "assistant" | "user";
            text: string;
          } =>
            (message.role === "assistant" || message.role === "user") &&
            Boolean(message.text),
        )
        .slice(-MAX_REALTIME_CONTEXT_MESSAGES);

      const response = await fetch(
        `${getApiUrl()}/workspaces/${workspaceSlug}/maya/realtime-session`,
        {
          body: JSON.stringify({
            currentPath,
            messages: contextMessages,
          }),
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          method: "POST",
          signal,
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
      if (!payload?.data?.clientSecret || !payload.data.sessionId) {
        throw new Error("Voice session did not include a client secret.");
      }
      return payload.data;
    },
    [currentPath, workspaceSlug],
  );

  const disconnect = useCallback(() => {
    if (status === "idle") {
      return;
    }
    setStatus("disconnecting");
    closeConnection();
    setStatus("idle");
  }, [closeConnection, status]);

  const connect = useCallback(async () => {
    if (status !== "idle") {
      return;
    }
    setError(null);

    if (!isBrowserRealtimeSupported()) {
      setError("Realtime voice is not supported in this browser.");
      return;
    }

    setStatus("connecting");
    voiceAnchorMessageIdRef.current =
      conversationMessagesRef.current.at(-1)?.id ?? null;
    const attemptId = connectionAttemptRef.current + 1;
    connectionAttemptRef.current = attemptId;
    const abortController = new AbortController();
    connectionAbortControllerRef.current = abortController;

    try {
      const localStream = await navigator.mediaDevices.getUserMedia({
        audio: {
          autoGainControl: true,
          echoCancellation: true,
          noiseSuppression: true,
        },
      });
      if (connectionAttemptRef.current !== attemptId) {
        localStream.getTracks().forEach((track) => {
          track.stop();
        });
        return;
      }
      localStreamRef.current = localStream;

      const session = await createRealtimeSession(abortController.signal);
      if (connectionAttemptRef.current !== attemptId) {
        return;
      }
      activeSessionIdRef.current = session.sessionId;

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
      dataChannel.onopen = () => {
        if (
          connectionAttemptRef.current !== attemptId ||
          dataChannel.readyState !== "open"
        ) {
          return;
        }
        setStatus("connected");
        startSessionTimer(session.maxSessionSeconds);
        resetIdleTimer();
        dataChannel.send(
          JSON.stringify({
            type: "response.create",
            response: {
              instructions:
                "Begin with one warm, concise sentence. Introduce yourself as Maya and ask what the user would like help with in FortyOne.",
            },
          }),
        );
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
        signal: abortController.signal,
      });
      if (!answerResponse.ok) {
        throw new Error("Failed to connect voice session.");
      }
      if (connectionAttemptRef.current !== attemptId) {
        return;
      }

      const answerSdp = await answerResponse.text();
      if (connectionAttemptRef.current !== attemptId) {
        return;
      }
      await peerConnection.setRemoteDescription({
        type: "answer",
        sdp: answerSdp,
      });
    } catch (connectError) {
      if (abortController.signal.aborted) {
        return;
      }
      closeConnection();
      setStatus("idle");
      setError(errorMessage(connectError, "Failed to start voice session."));
    }
  }, [
    closeConnection,
    createRealtimeSession,
    handleRealtimeEvent,
    resetIdleTimer,
    startSessionTimer,
    status,
  ]);

  const clearMessages = useCallback(() => {
    messagesRef.current = [];
    voiceMessageOrdersRef.current.clear();
    nextVoiceMessageOrderRef.current = 0;
    setMessages([]);
    setError(null);
  }, []);

  useEffect(() => {
    const handlePageHide = () => {
      closeConnection();
    };
    window.addEventListener("pagehide", handlePageHide);

    return () => {
      window.removeEventListener("pagehide", handlePageHide);
      closeConnection();
    };
  }, [closeConnection]);

  return {
    clearMessages,
    connect,
    disconnect,
    error,
    isListening: status === "connected" && !isSpeaking,
    isSpeaking,
    messages,
    remainingSeconds,
    status,
  };
};
