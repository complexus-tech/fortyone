"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getApiUrl } from "@/lib/api-url";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { feedbackKeys, notificationKeys, sprintKeys } from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import {
  applyRealtimeTranscriptUpdate,
  getMayaMessageText,
  getRealtimeTranscriptUpdate,
  mergeRealtimeVoiceMessages,
} from "../utils/realtime-voice-messages";
import {
  createRealtimePeerConnection,
  disposeRealtimePeerConnection,
} from "../utils/realtime-peer-connection";
import { extractRealtimeClientAction } from "../utils/realtime-client-actions";
import type { RealtimeClientAction } from "../utils/realtime-client-actions";
import {
  FALLBACK_REALTIME_MAX_SESSION_SECONDS,
  GOODBYE_DISCONNECT_DELAY_MS,
  getRealtimeErrorMessage,
  isBrowserRealtimeSupported,
  MAX_REALTIME_CONTEXT_MESSAGES,
  parseRealtimeCallAnswer,
  parseRealtimeSessionResponse,
  parseRealtimeToolOutput,
  REALTIME_ACTIVITY_EVENTS,
  REALTIME_CALLS_URL,
  REALTIME_IDLE_TIMEOUT_MS,
  type RealtimeFunctionCall,
  type RealtimeServerEvent,
  type RealtimeToolOutput,
  type RealtimeVoiceStatus,
  type UseMayaRealtimeVoiceOptions,
} from "./maya-realtime-voice-contract";
import { useRealtimeVoiceMessageOrder } from "./use-realtime-voice-message-order";

export const useMayaRealtimeVoice = ({
  conversationMessages,
  currentPath,
  navigate,
  setApplicationTheme,
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
  const pendingClientActionRef = useRef<RealtimeClientAction | null>(null);
  const navigateRef = useRef(navigate);
  const setApplicationThemeRef = useRef(setApplicationTheme);
  const conversationMessagesRef = useRef(conversationMessages);
  const messagesRef = useRef<MayaUIMessage[]>([]);
  const voiceAnchorMessageIdRef = useRef<string | null>(null);
  const { getMessageOrder, rememberEventItemOrder, resetMessageOrders } =
    useRealtimeVoiceMessageOrder();
  const [status, setStatus] = useState<RealtimeVoiceStatus>("idle");
  const [isSpeaking, setIsSpeaking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [messages, setMessages] = useState<MayaUIMessage[]>([]);
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);

  useEffect(() => {
    conversationMessagesRef.current = conversationMessages;
  }, [conversationMessages]);

  useEffect(() => {
    navigateRef.current = navigate;
    setApplicationThemeRef.current = setApplicationTheme;
  }, [navigate, setApplicationTheme]);

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
    pendingClientActionRef.current = null;

    const peerConnection = peerConnectionRef.current;
    peerConnectionRef.current = null;
    disposeRealtimePeerConnection(peerConnection);

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

  const runPendingClientAction = useCallback(() => {
    const action = pendingClientActionRef.current;
    pendingClientActionRef.current = null;
    if (!action) {
      return;
    }
    if (action.type === "navigate") {
      if (action.path.startsWith("/") && !action.path.startsWith("//")) {
        navigateRef.current(action.path);
      }
      return;
    }
    setApplicationThemeRef.current(action.theme);
  }, []);

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

      const sessionId = activeSessionIdRef.current;
      const output: RealtimeToolOutput = sessionId
        ? await fetch(
            `${getApiUrl()}/workspaces/${workspaceSlug}/maya/realtime-tool`,
            {
              method: "POST",
              credentials: "include",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({
                arguments: toolArguments,
                callId,
                name,
                sessionId,
              }),
            },
          )
            .then(parseRealtimeToolOutput)
            .catch((toolError: unknown) => ({
              success: false,
              error: getRealtimeErrorMessage(
                toolError,
                "Tool execution failed.",
              ),
            }))
        : {
            success: false,
            error: "The voice session is no longer active.",
          };
      if (
        (name === "create_task" ||
          name === "update_story" ||
          name === "story_comments") &&
        output.success === true
      ) {
        queryClient.invalidateQueries({
          queryKey: storyKeys.all(workspaceSlug),
        });
      }
      if (name === "notifications" && output.success === true) {
        queryClient.invalidateQueries({
          queryKey: notificationKeys.all(workspaceSlug),
        });
      }
      if (name === "sprints" && output.success === true) {
        queryClient.invalidateQueries({
          queryKey: sprintKeys.all(workspaceSlug),
        });
      }
      if (name === "customer_feedback" && output.success === true) {
        queryClient.invalidateQueries({
          queryKey: feedbackKeys.all(workspaceSlug),
        });
      }
      const { action, modelOutput } = extractRealtimeClientAction(output);
      if (action && output.success === true) {
        pendingClientActionRef.current = action;
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
            output: JSON.stringify(modelOutput),
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
        const order = getMessageOrder(transcriptUpdate.id);
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
          runPendingClientAction();
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
      getMessageOrder,
      handleFunctionCalls,
      markSpeaking,
      rememberEventItemOrder,
      resetIdleTimer,
      runPendingClientAction,
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
      return parseRealtimeSessionResponse(response);
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

      const peerConnection = createRealtimePeerConnection();
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
      if (connectionAttemptRef.current !== attemptId) {
        return;
      }

      const answerSdp = await parseRealtimeCallAnswer(answerResponse);
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
      setError(
        getRealtimeErrorMessage(connectError, "Failed to start voice session."),
      );
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
    resetMessageOrders();
    setMessages([]);
    setError(null);
  }, [resetMessageOrders]);

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
