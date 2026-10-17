"use client";

import { useEffect, useRef, useState } from "react";
import { MicrophoneIcon, StopIcon } from "icons";
import { Button, Popover, Tooltip } from "ui";
import { useAudioPlaybackController } from "./audio-playback-context";
import { useChatVoiceRecording } from "./use-chat-voice-recording";

const durationLabel = (seconds: number) =>
  `${Math.floor(seconds / 60)}:${Math.floor(seconds % 60)
    .toString()
    .padStart(2, "0")}`;

export const VoiceRecorder = ({
  disabled,
  onAttach,
  onActiveChange,
}: {
  disabled: boolean;
  onAttach: (file: File, durationSeconds: number) => Promise<boolean>;
  onActiveChange: (active: boolean) => void;
}) => {
  const [open, setOpen] = useState(false);
  const [attaching, setAttaching] = useState(false);
  const [attachError, setAttachError] = useState<string | null>(null);
  const attempt = useRef(0);
  const attachingRef = useRef(false);
  const audio = useRef<HTMLAudioElement>(null);
  const playback = useAudioPlaybackController();
  const { phase, elapsedSeconds, preview, error, start, stop, discard } =
    useChatVoiceRecording();
  const setVisible = (next: boolean) => {
    if (attachingRef.current) return;
    if (!next) {
      audio.current?.pause();
      discard();
    }
    setOpen(next);
    setAttachError(null);
    onActiveChange(next);
  };
  const record = () => {
    if (attachingRef.current) return;
    audio.current?.pause();
    playback.pause();
    setAttachError(null);
    void start();
  };
  const attach = async () => {
    if (!preview || attachingRef.current) return;
    attachingRef.current = true;
    const current = ++attempt.current;
    setAttaching(true);
    setAttachError(null);
    audio.current?.pause();
    return Promise.resolve()
      .then(() => onAttach(preview.file, preview.durationSeconds))
      .then((attached) => {
        if (attempt.current !== current) return;
        if (attached) {
          discard();
          setOpen(false);
          onActiveChange(false);
        }
      })
      .catch(() => {
        if (attempt.current === current)
          setAttachError("Could not attach this voice note. Please try again.");
      })
      .finally(() => {
        if (attempt.current === current) {
          attachingRef.current = false;
          setAttaching(false);
        }
      });
  };
  useEffect(
    () => () => {
      attempt.current += 1;
      onActiveChange(false);
    },
    [onActiveChange],
  );
  useEffect(
    () =>
      playback.registerExternalPause(() => {
        audio.current?.pause();
      }),
    [playback],
  );

  return (
    <Popover onOpenChange={setVisible} open={open}>
      <Tooltip title="Record a voice note">
        <Popover.Trigger asChild>
          <Button
            aria-label="Record a voice note"
            asIcon
            className="text-text-muted size-7 rounded-sm [&_svg]:text-current"
            color="tertiary"
            disabled={disabled || attaching}
            onClick={() => {
              if (!open) record();
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            <MicrophoneIcon className="size-4" />
          </Button>
        </Popover.Trigger>
      </Tooltip>
      <Popover.Content
        align="start"
        aria-label="Voice note recorder"
        className="w-[360px] max-w-[calc(100vw-2rem)] p-4"
        onEscapeKeyDown={(event) => {
          if (attaching) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          event.preventDefault();
        }}
      >
        <div className="flex items-center justify-between gap-3">
          <p className="font-semibold">Voice note</p>
          <span className="text-text-muted tabular-nums">
            {durationLabel(elapsedSeconds)} / 5:00
          </span>
        </div>
        {phase === "requesting" ? (
          <p className="text-text-muted mt-3" role="status">
            Waiting for microphone access…
          </p>
        ) : null}
        {phase === "recording" ? (
          <div className="mt-4 flex items-center gap-2" role="status">
            <span
              aria-hidden
              className="bg-danger size-2 animate-pulse rounded-full motion-reduce:animate-none"
            />
            <span>Recording…</span>
            <Button
              className="ml-auto"
              color="tertiary"
              onClick={stop}
              size="sm"
              type="button"
              variant="outline"
            >
              <StopIcon className="mr-1.5 size-3" /> Stop
            </Button>
          </div>
        ) : null}
        {phase === "stopping" ? (
          <p className="text-text-muted mt-3" role="status">
            Finishing your recording…
          </p>
        ) : null}
        {preview ? (
          <div className="mt-3 space-y-2">
            <p className="text-text-muted">
              Listen before attaching it to your message.
            </p>
            {/* eslint-disable-next-line jsx-a11y/media-has-caption -- Unpublished local recordings have no transcript yet. */}
            <audio
              aria-label="Voice note preview"
              className="h-10 w-full"
              controls
              controlsList="nodownload"
              onPlay={() => {
                playback.pause();
              }}
              preload="metadata"
              ref={audio}
              src={preview.url}
            />
          </div>
        ) : null}
        {error || attachError ? (
          <p className="text-danger mt-3" role="alert">
            {error || attachError}
          </p>
        ) : null}
        <div className="mt-4 flex flex-wrap items-center justify-end gap-2">
          <Button
            color="tertiary"
            disabled={attaching}
            onClick={() => {
              setVisible(false);
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            {preview ? "Discard" : "Cancel"}
          </Button>
          {preview || error ? (
            <Button
              color="tertiary"
              disabled={attaching}
              onClick={record}
              size="sm"
              type="button"
              variant="outline"
            >
              Record again
            </Button>
          ) : null}
          {preview ? (
            <Button
              disabled={attaching || disabled}
              onClick={() => {
                void attach();
              }}
              size="sm"
              type="button"
            >
              {attaching ? "Attaching…" : "Attach voice note"}
            </Button>
          ) : null}
        </div>
      </Popover.Content>
    </Popover>
  );
};
