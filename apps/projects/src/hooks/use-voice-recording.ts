import { useState, useRef, useCallback, useEffect } from "react";
import { toast } from "sonner";

type RecordingState = "idle" | "recording" | "processing";

const MAX_RECORDING_TIME = 60; // 60 seconds max

export const useVoiceRecording = (onRecordingStop?: () => void) => {
  const [isRecording, setIsRecording] = useState(false);
  const [recordingState, setRecordingState] = useState<RecordingState>("idle");
  const [recordingDuration, setRecordingDuration] = useState(0);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Helper function to format duration
  const formatDuration = useCallback((seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }, []);

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

      // Start timer
      setRecordingDuration(0);
      timerRef.current = setInterval(() => {
        setRecordingDuration((prev) => prev + 1);
      }, 1000);
    } catch (error) {
      toast.error("Failed to start recording", {
        description: "Please allow microphone access and try again.",
      });
    }
  }, []);

  const stopRecording = useCallback(() => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
      setRecordingState("processing");

      // Stop timer
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }

      // Call the callback to trigger transcription
      if (onRecordingStop) {
        onRecordingStop();
      }
    }
  }, [isRecording, onRecordingStop]);

  const getAudioBlob = useCallback(() => {
    if (audioChunksRef.current.length === 0) return null;

    const audioBlob = new Blob(audioChunksRef.current, { type: "audio/webm" });
    return audioBlob;
  }, []);

  const resetRecording = useCallback(() => {
    audioChunksRef.current = [];
    setRecordingState("idle");
    setRecordingDuration(0);

    // Clear timer if it's still running
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Auto-stop when max duration is reached
  useEffect(() => {
    if (recordingDuration >= MAX_RECORDING_TIME) {
      stopRecording();
    }
  }, [recordingDuration, stopRecording]);

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
