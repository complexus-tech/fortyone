import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { RealtimeTheme } from "../utils/realtime-client-actions";

export type RealtimeVoiceStatus =
  | "idle"
  | "connecting"
  | "connected"
  | "disconnecting";

export type ApiResponse<T> = {
  data?: T;
  error?: {
    message?: string;
  };
};

export type RealtimeSessionResponse = {
  clientSecret: string;
  expiresAt?: number;
  maxSessionSeconds: number;
  model: string;
  monthlyLimitSeconds: number;
  remainingSeconds: number;
  sessionId: string;
  voice: string;
};

export type RealtimeFunctionCall = {
  arguments?: string;
  call_id?: string;
  name?: string;
  type?: string;
};

export type RealtimeServerEvent = {
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

export type RealtimeToolOutput = {
  clientAction?: unknown;
  success?: boolean;
  error?: string;
};

export type UseMayaRealtimeVoiceOptions = {
  conversationMessages: MayaUIMessage[];
  currentPath: string;
  navigate: (path: string) => void;
  setApplicationTheme: (theme: RealtimeTheme) => void;
};

export const REALTIME_CALLS_URL = "https://api.openai.com/v1/realtime/calls";
export const FALLBACK_REALTIME_MAX_SESSION_SECONDS = 5 * 60;
export const REALTIME_IDLE_TIMEOUT_MS = 60_000;
export const GOODBYE_DISCONNECT_DELAY_MS = 800;
export const MAX_REALTIME_CONTEXT_MESSAGES = 24;
export const REALTIME_ACTIVITY_EVENTS = new Set([
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

export const isBrowserRealtimeSupported = () => {
  if (
    typeof window === "undefined" ||
    typeof RTCPeerConnection === "undefined" ||
    !("mediaDevices" in navigator)
  ) {
    return false;
  }

  return "getUserMedia" in navigator.mediaDevices;
};

export const getRealtimeErrorMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }
  return fallback;
};

const parseRealtimeApiResponse = async <T>(response: Response) =>
  (await response.json().catch(() => null)) as ApiResponse<T> | null;

export const parseRealtimeSessionResponse = async (response: Response) => {
  if (!response.ok) {
    const payload =
      await parseRealtimeApiResponse<RealtimeSessionResponse>(response);
    throw new Error(
      payload?.error?.message ?? "Failed to create voice session.",
    );
  }

  const payload =
    await parseRealtimeApiResponse<RealtimeSessionResponse>(response);
  if (!payload?.data?.clientSecret || !payload.data.sessionId) {
    throw new Error("Voice session did not include a client secret.");
  }

  return payload.data;
};

export const parseRealtimeCallAnswer = async (response: Response) => {
  if (!response.ok) {
    throw new Error("Failed to connect voice session.");
  }

  return response.text();
};

export const parseRealtimeToolOutput = async (response: Response) => {
  if (!response.ok) {
    const payload =
      await parseRealtimeApiResponse<RealtimeToolOutput>(response);
    return {
      success: false,
      error: payload?.error?.message ?? "Tool execution failed.",
    };
  }

  const payload = await parseRealtimeApiResponse<RealtimeToolOutput>(response);
  return (
    payload?.data ?? {
      success: false,
      error: "Tool returned an unreadable response.",
    }
  );
};
