import { useState, useRef, useCallback } from "react";
import { toast } from "sonner";

type RecordingState = "idle" | "recording" | "processing";

export const useVoiceRecording = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [recordingState, setRecordingState] = useState<RecordingState>("idle");
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);

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
    }
  }, [isRecording]);

  const getAudioBlob = useCallback(() => {
    if (audioChunksRef.current.length === 0) return null;

    const audioBlob = new Blob(audioChunksRef.current, { type: "audio/webm" });
    return audioBlob;
  }, []);

  const resetRecording = useCallback(() => {
    audioChunksRef.current = [];
    setRecordingState("idle");
  }, []);

  return {
    isRecording,
    recordingState,
    startRecording,
    stopRecording,
    getAudioBlob,
    resetRecording,
  };
};
