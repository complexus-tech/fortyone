import { useState, useRef, useCallback, useEffect } from "react";
import { toast } from "sonner";
import { getMicrophoneErrorDescription } from "@/lib/microphone-error";

type RecordingState = "idle" | "requesting" | "recording" | "processing";

type RecordingSession = {
  recorder: MediaRecorder | null;
  stream: MediaStream | null;
  chunks: Blob[];
  completion: Promise<Blob | null>;
  resolve: (blob: Blob | null) => void;
  settled: boolean;
  stopping: boolean;
  autoStop: boolean;
};

const MAX_RECORDING_TIME = 60;
const RECORDING_MIME_TYPES = [
  "audio/webm;codecs=opus",
  "audio/webm",
  "audio/mp4",
];

const releaseStream = (stream: MediaStream | null) => {
  stream?.getTracks().forEach((track) => {
    track.stop();
  });
};

const disposeSession = (session: RecordingSession) => {
  const recorder = session.recorder;
  if (recorder) {
    recorder.ondataavailable = null;
    recorder.onstop = null;
    recorder.onerror = null;
    if (recorder.state !== "inactive") {
      try {
        recorder.stop();
      } catch {
        // Releasing the tracks below also ends capture if the recorder failed.
      }
    }
  }
  releaseStream(session.stream);
  session.stream = null;
  session.chunks = [];
};

export const useVoiceRecording = (onAutoStop?: (blob: Blob) => void) => {
  const [recordingState, setRecordingState] = useState<RecordingState>("idle");
  const [recordingDuration, setRecordingDuration] = useState(0);
  const stateRef = useRef<RecordingState>("idle");
  const sessionRef = useRef<RecordingSession | null>(null);
  const mountedRef = useRef(true);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const onAutoStopRef = useRef(onAutoStop);

  useEffect(() => {
    onAutoStopRef.current = onAutoStop;
  }, [onAutoStop]);

  const updateState = useCallback((state: RecordingState) => {
    stateRef.current = state;
    if (mountedRef.current) setRecordingState(state);
  }, []);

  const stopTimer = useCallback(() => {
    if (timerRef.current !== null) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const isCurrentSession = useCallback(
    (session: RecordingSession) =>
      mountedRef.current && sessionRef.current === session,
    [],
  );

  const finishSession = useCallback(
    (session: RecordingSession, blob: Blob | null) => {
      if (session.settled) return;
      session.settled = true;
      disposeSession(session);
      session.resolve(blob);

      if (sessionRef.current !== session) return;
      stopTimer();
      updateState(blob ? "processing" : "idle");
      if (!blob) sessionRef.current = null;
      if (blob && session.autoStop && mountedRef.current) {
        onAutoStopRef.current?.(blob);
      }
    },
    [stopTimer, updateState],
  );

  const resetRecording = useCallback(() => {
    const session = sessionRef.current;
    sessionRef.current = null;
    if (session) {
      if (session.settled) disposeSession(session);
      else finishSession(session, null);
    }
    stopTimer();
    updateState("idle");
    if (mountedRef.current) setRecordingDuration(0);
  }, [finishSession, stopTimer, updateState]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      resetRecording();
    };
  }, [resetRecording]);

  const stopRecording = useCallback((): Promise<Blob | null> => {
    const session = sessionRef.current;
    if (!session) return Promise.resolve(null);
    if (!session.recorder) {
      resetRecording();
      return session.completion;
    }
    if (session.settled || session.stopping) return session.completion;

    session.stopping = true;
    stopTimer();
    updateState("processing");
    try {
      if (session.recorder.state !== "inactive") session.recorder.stop();
    } catch (error) {
      finishSession(session, null);
      toast.error("Failed to finish recording", {
        description: getMicrophoneErrorDescription(error),
      });
    }
    // MediaRecorder emits its final dataavailable event before stop.
    return session.completion;
  }, [finishSession, resetRecording, stopTimer, updateState]);

  const startRecording = useCallback(async () => {
    if (!mountedRef.current || stateRef.current !== "idle") return;

    if (
      typeof navigator.mediaDevices === "undefined" ||
      typeof navigator.mediaDevices.getUserMedia !== "function" ||
      typeof globalThis.MediaRecorder !== "function"
    ) {
      toast.error("Recording is unavailable", {
        description: "Audio recording is not supported in this app or browser.",
      });
      return;
    }

    let resolveCompletion!: (blob: Blob | null) => void;
    const completion = new Promise<Blob | null>((resolve) => {
      resolveCompletion = resolve;
    });
    const session: RecordingSession = {
      recorder: null,
      stream: null,
      chunks: [],
      completion,
      resolve: resolveCompletion,
      settled: false,
      stopping: false,
      autoStop: false,
    };
    sessionRef.current = session;
    updateState("requesting");
    setRecordingDuration(0);

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      if (!isCurrentSession(session)) {
        releaseStream(stream);
        return;
      }
      session.stream = stream;
      const mimeType =
        typeof MediaRecorder.isTypeSupported === "function"
          ? RECORDING_MIME_TYPES.find((type) =>
              MediaRecorder.isTypeSupported(type),
            )
          : undefined;
      const recorder = mimeType
        ? new MediaRecorder(stream, { mimeType })
        : new MediaRecorder(stream);
      session.recorder = recorder;
      recorder.ondataavailable = (event) => {
        if (event.data.size > 0) session.chunks.push(event.data);
      };
      recorder.onstop = () => {
        const blob = session.chunks.length
          ? new Blob(session.chunks, {
              type: recorder.mimeType || session.chunks[0].type,
            })
          : null;
        // A disconnected microphone can end capture without an explicit stop.
        if (!session.stopping) session.autoStop = true;
        finishSession(session, blob);
        if (!blob && mountedRef.current) {
          toast.error("No audio was recorded", {
            description: "Check your microphone and try again.",
          });
        }
      };
      recorder.onerror = (event) => {
        finishSession(session, null);
        toast.error("Recording was interrupted", {
          description: getMicrophoneErrorDescription(event.error),
        });
      };
      recorder.start();
      updateState("recording");
      const startedAt = Date.now();
      timerRef.current = setInterval(() => {
        const duration = Math.min(
          Math.floor((Date.now() - startedAt) / 1000),
          MAX_RECORDING_TIME,
        );
        setRecordingDuration(duration);
        if (duration >= MAX_RECORDING_TIME) {
          session.autoStop = true;
          void stopRecording();
        }
      }, 1000);
    } catch (error) {
      if (!isCurrentSession(session)) return;
      finishSession(session, null);
      toast.error("Failed to start recording", {
        description: getMicrophoneErrorDescription(error),
      });
    }
  }, [finishSession, isCurrentSession, stopRecording, updateState]);

  const formatDuration = useCallback((seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  }, []);

  return {
    isRecording: recordingState === "recording",
    recordingState,
    recordingDuration,
    formatDuration,
    MAX_RECORDING_TIME,
    startRecording,
    stopRecording,
    resetRecording,
  };
};
