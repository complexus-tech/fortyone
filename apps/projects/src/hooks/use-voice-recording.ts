import { useState, useRef, useCallback, useEffect } from "react";
import { toast } from "sonner";

type RecordingState = "idle" | "recording" | "processing";

const MAX_RECORDING_TIME = 60; // 60 seconds max

export const useVoiceRecording = (onAutoStop?: () => void) => {
  const [isRecording, setIsRecording] = useState(false);
  const [recordingState, setRecordingState] = useState<RecordingState>("idle");
  const [recordingDuration, setRecordingDuration] = useState(0);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const onAutoStopRef = useRef(onAutoStop);

  useEffect(() => {
    onAutoStopRef.current = onAutoStop;
  }, [onAutoStop]);

  // Helper function to format duration
  const formatDuration = useCallback((seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }, []);

  const stopTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const stopRecording = useCallback(() => {
    const mediaRecorder = mediaRecorderRef.current;
    if (!mediaRecorder || mediaRecorder.state === "inactive") return;

    mediaRecorder.stop();
    setIsRecording(false);
    setRecordingState("processing");
    stopTimer();
  }, [stopTimer]);

  const getAudioBlob = useCallback(() => {
    if (audioChunksRef.current.length === 0) return null;

    const audioBlob = new Blob(audioChunksRef.current, { type: "audio/webm" });
    return audioBlob;
  }, []);

  const resetRecording = useCallback(() => {
    audioChunksRef.current = [];
    setRecordingState("idle");
    setRecordingDuration(0);
    stopTimer();
  }, [stopTimer]);

  const startRecording = useCallback(async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

      const mediaRecorder = new MediaRecorder(stream, {
        mimeType: "audio/webm;codecs=opus",
      });

      mediaRecorderRef.current = mediaRecorder;
      audioChunksRef.current = [];

      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunksRef.current.push(event.data);
        }
      };

      mediaRecorder.onstop = () => {
        stream.getTracks().forEach((track) => {
          track.stop();
        });
      };

      mediaRecorder.start();
      setIsRecording(true);
      setRecordingState("recording");
      setRecordingDuration(0);
      stopTimer();
      timerRef.current = setInterval(() => {
        setRecordingDuration((prev) => {
          const nextDuration = prev + 1;

          if (nextDuration >= MAX_RECORDING_TIME) {
            stopRecording();
            onAutoStopRef.current?.();
          }

          return nextDuration;
        });
      }, 1000);
    } catch (error) {
      toast.error("Failed to start recording", {
        description: "Please allow microphone access and try again.",
      });
    }
  }, [stopRecording, stopTimer]);

  return {
    isRecording,
    recordingState,
    recordingDuration,
    formatDuration,
    MAX_RECORDING_TIME,
    startRecording,
    stopRecording,
    getAudioBlob,
    resetRecording,
  };
};
