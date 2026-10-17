"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { getMicrophoneErrorDescription } from "@/lib/microphone-error";

export const CHAT_VOICE_MAX_SECONDS = 300;
export const CHAT_VOICE_MAX_BYTES = 20 * 1024 * 1024;
const RECORDING_TYPES = ["audio/webm;codecs=opus", "audio/mp4", "audio/webm"];

type VoicePreview = { file: File; url: string; durationSeconds: number };
type RecordingPhase =
  | "idle"
  | "requesting"
  | "recording"
  | "stopping"
  | "ready";
type RecordingState = {
  phase: RecordingPhase;
  elapsedSeconds: number;
  preview: VoicePreview | null;
  error: string | null;
};
type RecordingSession = {
  recorder: MediaRecorder | null;
  stream: MediaStream | null;
  chunks: Blob[];
  bytes: number;
  startedAt: number;
  durationSeconds: number;
  stopping: boolean;
  interval?: ReturnType<typeof setInterval>;
  limit?: ReturnType<typeof setTimeout>;
  finishing?: ReturnType<typeof setTimeout>;
};
const INITIAL_STATE: RecordingState = {
  phase: "idle",
  elapsedSeconds: 0,
  preview: null,
  error: null,
};
const releaseStream = (stream: MediaStream | null) => {
  stream?.getTracks().forEach((track) => {
    track.stop();
  });
};
const disposeSession = (session: RecordingSession) => {
  clearInterval(session.interval);
  clearTimeout(session.limit);
  clearTimeout(session.finishing);
  if (session.recorder) {
    session.recorder.ondataavailable = null;
    session.recorder.onstop = null;
    session.recorder.onerror = null;
    if (session.recorder.state !== "inactive") {
      try {
        session.recorder.stop();
      } catch {
        /* Releasing tracks also ends capture. */
      }
    }
  }
  releaseStream(session.stream);
  session.chunks = [];
};

/** A recording stays local until the user reviews and explicitly attaches it. */
export const useChatVoiceRecording = () => {
  const [state, setState] = useState<RecordingState>(INITIAL_STATE);
  const sessionRef = useRef<RecordingSession | null>(null);
  const previewUrl = useRef<string | null>(null);
  const mounted = useRef(true);
  const isCurrentSession = useCallback(
    (session: RecordingSession) =>
      mounted.current && sessionRef.current === session,
    [],
  );

  const discard = useCallback(() => {
    const current = sessionRef.current;
    sessionRef.current = null;
    if (current) disposeSession(current);
    if (previewUrl.current) URL.revokeObjectURL(previewUrl.current);
    previewUrl.current = null;
    if (mounted.current) setState(INITIAL_STATE);
  }, []);
  const fail = useCallback((session: RecordingSession, error: string) => {
    if (sessionRef.current !== session) return;
    sessionRef.current = null;
    disposeSession(session);
    if (mounted.current) setState({ ...INITIAL_STATE, error });
  }, []);
  const stop = useCallback(() => {
    const session = sessionRef.current;
    if (!session?.recorder || session.stopping) return;
    session.stopping = true;
    session.durationSeconds = Math.min(
      (Date.now() - session.startedAt) / 1000,
      CHAT_VOICE_MAX_SECONDS,
    );
    clearInterval(session.interval);
    clearTimeout(session.limit);
    setState((current) => ({
      ...current,
      phase: "stopping",
      elapsedSeconds: Math.floor(session.durationSeconds),
    }));
    session.finishing = setTimeout(() => {
      fail(session, "The recording could not be finished. Please try again.");
    }, 5000);
    try {
      if (session.recorder.state !== "inactive") session.recorder.stop();
    } catch (error) {
      fail(session, getMicrophoneErrorDescription(error));
    }
  }, [fail]);

  const start = useCallback(async () => {
    if (!mounted.current || sessionRef.current) return;
    discard();
    if (
      typeof navigator.mediaDevices === "undefined" ||
      typeof navigator.mediaDevices.getUserMedia !== "function" ||
      typeof MediaRecorder !== "function" ||
      typeof MediaRecorder.isTypeSupported !== "function"
    ) {
      setState({
        ...INITIAL_STATE,
        error: "Audio recording is not supported in this browser.",
      });
      return;
    }
    const mimeType = RECORDING_TYPES.find((type) =>
      MediaRecorder.isTypeSupported(type),
    );
    if (!mimeType) {
      setState({
        ...INITIAL_STATE,
        error: "This browser cannot record a supported audio format.",
      });
      return;
    }
    const session: RecordingSession = {
      recorder: null,
      stream: null,
      chunks: [],
      bytes: 0,
      startedAt: 0,
      durationSeconds: 0,
      stopping: false,
    };
    sessionRef.current = session;
    setState({ ...INITIAL_STATE, phase: "requesting" });
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      if (!isCurrentSession(session)) {
        releaseStream(stream);
        return;
      }
      session.stream = stream;
      const recorder = new MediaRecorder(stream, {
        mimeType,
        audioBitsPerSecond: 64_000,
      });
      session.recorder = recorder;
      recorder.ondataavailable = (event) => {
        if (sessionRef.current !== session || !event.data.size) return;
        if (session.bytes + event.data.size > CHAT_VOICE_MAX_BYTES) {
          fail(
            session,
            "Voice notes must be 20 MB or smaller. Try a shorter recording.",
          );
          return;
        }
        session.chunks.push(event.data);
        session.bytes += event.data.size;
      };
      recorder.onstop = () => {
        if (sessionRef.current !== session) return;
        if (!session.chunks.length) {
          fail(
            session,
            "No audio was recorded. Check your microphone and try again.",
          );
          return;
        }
        const durationSeconds = session.stopping
          ? session.durationSeconds
          : Math.min(
              (Date.now() - session.startedAt) / 1000,
              CHAT_VOICE_MAX_SECONDS,
            );
        const type = recorder.mimeType || mimeType;
        const extension = type.startsWith("audio/mp4") ? "m4a" : "webm";
        const file = new File(
          session.chunks,
          `Voice note ${new Date().toISOString().replace(/[:.]/g, "-")}.${extension}`,
          { type },
        );
        sessionRef.current = null;
        disposeSession(session);
        const url = URL.createObjectURL(file);
        previewUrl.current = url;
        if (mounted.current)
          setState({
            phase: "ready",
            preview: { file, url, durationSeconds },
            elapsedSeconds: Math.floor(durationSeconds),
            error: null,
          });
      };
      recorder.onerror = (event) => {
        fail(session, getMicrophoneErrorDescription(event.error));
      };
      recorder.start(1000);
      session.startedAt = Date.now();
      setState({ ...INITIAL_STATE, phase: "recording" });
      session.interval = setInterval(() => {
        const elapsedSeconds = Math.min(
          Math.floor((Date.now() - session.startedAt) / 1000),
          CHAT_VOICE_MAX_SECONDS,
        );
        setState((current) => ({ ...current, elapsedSeconds }));
      }, 1000);
      session.limit = setTimeout(stop, CHAT_VOICE_MAX_SECONDS * 1000);
    } catch (error) {
      fail(session, getMicrophoneErrorDescription(error));
    }
  }, [discard, fail, isCurrentSession, stop]);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      discard();
    };
  }, [discard]);
  return { ...state, start, stop, discard };
};
