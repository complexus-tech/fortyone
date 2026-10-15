"use client";

import type { ChangeEvent } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import ky from "ky";
import { toast } from "sonner";
import { useVoiceRecording } from "@/hooks/use-voice-recording";
import {
  getNetworkOnline,
  subscribeToNetworkStatus,
  useNetworkOnline,
} from "@/shared/network/status";

type RecordedChatMessageOptions = {
  value: string;
  onChange: (event: ChangeEvent<HTMLTextAreaElement>) => void;
  conversationId?: string | null;
  disabled: boolean;
};

const recordingFilename = (mimeType: string) => {
  const extensions: Record<string, string> = {
    "audio/webm": "webm",
    "audio/mp4": "m4a",
    "audio/ogg": "ogg",
    "audio/wav": "wav",
    "audio/mpeg": "mp3",
  };
  const extension = extensions[mimeType.split(";")[0]];
  return extension ? `recording.${extension}` : "recording";
};

export const useRecordedChatMessage = ({
  value,
  onChange,
  conversationId,
  disabled,
}: RecordedChatMessageOptions) => {
  const pathname = usePathname();
  const isOnline = useNetworkOnline();
  const [isTranscribing, setIsTranscribing] = useState(false);
  const processRecordingRef = useRef<(blob: Blob) => void>(() => {});
  const transcriptionRef = useRef<AbortController | null>(null);
  const recordingAttemptRef = useRef(0);
  const draftRef = useRef({ value, onChange });
  useEffect(() => {
    draftRef.current = { value, onChange };
  }, [value, onChange]);
  const handleAutoStop = useCallback((blob: Blob) => {
    processRecordingRef.current(blob);
  }, []);
  const {
    isRecording,
    recordingState,
    recordingDuration,
    formatDuration,
    MAX_RECORDING_TIME,
    startRecording,
    stopRecording,
    resetRecording,
  } = useVoiceRecording(handleAutoStop);
  const recordingBusy = recordingState !== "idle" || isTranscribing;
  const canCancelRecording =
    recordingState === "recording" || recordingState === "requesting";
  const processRecording = useCallback(
    async (audioBlob: Blob | null) => {
      if (transcriptionRef.current) return;
      if (!audioBlob?.size) {
        toast.info("No audio recorded", {
          description: "Please try again.",
          duration: 2500,
          closeButton: false,
        });
        resetRecording();
        return;
      }
      if (!getNetworkOnline()) {
        toast.error("Transcription requires an internet connection.");
        resetRecording();
        return;
      }
      const controller = new AbortController();
      transcriptionRef.current = controller;
      setIsTranscribing(true);
      const formData = new FormData();
      formData.append("audio", audioBlob, recordingFilename(audioBlob.type));
      try {
        const response = await ky
          .post("/api/transcribe", {
            body: formData,
            signal: controller.signal,
            retry: 0,
          })
          .json<{ text: string }>();
        if (controller.signal.aborted) return;
        const text = response.text.trim();
        if (text) {
          const draft = draftRef.current;
          draft.onChange({
            target: { value: [draft.value, text].filter(Boolean).join(" ") },
          } as ChangeEvent<HTMLTextAreaElement>);
        }
      } catch {
        if (!controller.signal.aborted) {
          toast.error("Failed to transcribe audio", {
            description: "Please try again.",
            duration: 2500,
            closeButton: false,
          });
        }
      }
      if (transcriptionRef.current === controller) {
        transcriptionRef.current = null;
        setIsTranscribing(false);
        resetRecording();
      }
    },
    [resetRecording],
  );

  const handleVoiceRecording = async () => {
    if (disabled || !getNetworkOnline()) return;
    if (isRecording) {
      const attempt = recordingAttemptRef.current;
      const blob = await stopRecording();
      if (recordingAttemptRef.current === attempt) await processRecording(blob);
    } else if (!recordingBusy) {
      await startRecording();
    }
  };

  const cancelRecording = useCallback(() => {
    recordingAttemptRef.current += 1;
    transcriptionRef.current?.abort();
    transcriptionRef.current = null;
    setIsTranscribing(false);
    resetRecording();
  }, [resetRecording]);

  useEffect(() => {
    window.addEventListener("pagehide", cancelRecording);
    return () => {
      window.removeEventListener("pagehide", cancelRecording);
      cancelRecording();
    };
  }, [cancelRecording, pathname, conversationId]);

  useEffect(() => {
    return subscribeToNetworkStatus(() => {
      if (getNetworkOnline()) return;
      cancelRecording();
      if (recordingBusy) {
        toast.error("Recording stopped", {
          description:
            "Reconnect to the internet and record your message again.",
        });
      }
    });
  }, [recordingBusy, cancelRecording]);

  useEffect(() => {
    processRecordingRef.current = (blob) => {
      void processRecording(blob);
    };
  }, [processRecording]);

  return {
    isRecording,
    recordingState,
    recordingDuration,
    formatDuration,
    MAX_RECORDING_TIME,
    recordingBusy,
    canCancelRecording,
    isTranscribing,
    isOnline,
    handleVoiceRecording,
    cancelRecording,
  };
};
