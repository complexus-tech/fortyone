"use client";

import { useId, useState } from "react";
import {
  DownloadIcon,
  LoadingIcon,
  MoreVerticalIcon,
  TranscribeIcon,
} from "icons";
import { Button, Menu, Tooltip } from "ui";
import type { AudioTrack } from "./audio-playback-controller";
import { AudioPlayerControls } from "./audio-player-controls";
import {
  useAudioPlaybackController,
  useAudioPlaybackState,
  useAudioSourceVisibility,
} from "./audio-playback-context";

export type AudioTranscript = {
  status: "idle" | "pending" | "ready" | "error";
  text?: string;
  error?: string;
};

export const AudioAttachmentPlayer = ({
  attachment,
  transcript = { status: "idle" },
  onRequestTranscript,
}: {
  attachment: AudioTrack;
  transcript?: AudioTranscript;
  onRequestTranscript?: () => void;
}) => {
  const controller = useAudioPlaybackController();
  const state = useAudioPlaybackState(attachment.id);
  const ref = useAudioSourceVisibility(attachment.id);
  const transcriptId = useId();
  const [expanded, setExpanded] = useState(false);
  const pending = transcript.status === "pending";
  const hasTranscript = transcript.status === "ready";
  let transcriptLabel = "Transcribe";
  if (pending) transcriptLabel = "Transcribing…";
  else if (transcript.status === "error")
    transcriptLabel = "Retry transcription";
  else if (hasTranscript)
    transcriptLabel = expanded ? "Hide transcript" : "Show transcript";
  return (
    <div
      aria-label={attachment.fileName}
      className="border-border bg-surface w-[420px] max-w-full rounded-lg border px-3 py-2 dark:bg-[color-mix(in_oklab,var(--color-background)_65%,var(--color-surface))]"
      ref={ref}
      role="group"
    >
      <AudioPlayerControls
        controller={controller}
        moreActions={
          <Menu>
            <Tooltip title="Audio options">
              <Menu.Button>
                <Button
                  aria-label="Audio options"
                  asIcon
                  className="size-7 shrink-0 [&_svg]:text-current"
                  color="tertiary"
                  size="xs"
                  type="button"
                  variant="naked"
                >
                  <MoreVerticalIcon aria-hidden className="size-4" />
                </Button>
              </Menu.Button>
            </Tooltip>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item asChild>
                  <a
                    download={attachment.fileName}
                    href={attachment.url}
                    rel="noopener noreferrer"
                    target="_blank"
                  >
                    <DownloadIcon aria-hidden className="size-4" /> Download
                    audio
                  </a>
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        }
        state={state}
        track={attachment}
        transcriptControl={
          onRequestTranscript || transcript.status !== "idle" ? (
            <Tooltip title={transcriptLabel}>
              <Button
                aria-controls={transcriptId}
                aria-expanded={expanded}
                aria-label={transcriptLabel}
                asIcon
                className="text-text-muted hover:text-foreground size-7 shrink-0 [&_svg]:text-current"
                color="tertiary"
                disabled={pending || (!hasTranscript && !onRequestTranscript)}
                onClick={() => {
                  if (hasTranscript) setExpanded(!expanded);
                  else {
                    setExpanded(true);
                    onRequestTranscript?.();
                  }
                }}
                size="xs"
                type="button"
                variant="naked"
              >
                {pending ? (
                  <LoadingIcon aria-hidden className="size-4.5 animate-spin" />
                ) : (
                  <TranscribeIcon aria-hidden className="size-4.5" />
                )}
              </Button>
            </Tooltip>
          ) : undefined
        }
      />
      {state.error ? (
        <p className="text-danger mt-2" role="alert">
          {state.error}
        </p>
      ) : null}
      {onRequestTranscript || transcript.status !== "idle" ? (
        <div aria-live="polite" id={transcriptId}>
          {pending ? (
            <span className="sr-only">Transcribing recording…</span>
          ) : null}
          {expanded && hasTranscript ? (
            <p className="mt-2 leading-relaxed whitespace-pre-wrap">
              {transcript.text?.trim() ||
                "No speech was detected in this recording."}
            </p>
          ) : null}
          {transcript.status === "error" ? (
            <p className="text-text-muted mt-1">
              {transcript.error ??
                "The transcript could not be created. Try again."}
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  );
};
