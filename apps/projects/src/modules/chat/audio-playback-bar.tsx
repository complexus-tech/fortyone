"use client";

import { CloseIcon, ReplyIcon } from "icons";
import { Button, Tooltip } from "ui";
import { AudioPlayerControls } from "./audio-player-controls";
import {
  useAudioPlaybackController,
  useAudioPlaybackNavigation,
  useAudioPlaybackState,
} from "./audio-playback-context";

export const AudioPlaybackBar = () => {
  const controller = useAudioPlaybackController();
  const state = useAudioPlaybackState();
  const onViewMessage = useAudioPlaybackNavigation();
  if (!state.track || state.sourceVisible || state.status === "ended")
    return null;
  return (
    <section
      aria-label="Audio playing outside this view"
      className="border-border/50 bg-surface-elevated mx-4 mb-2 rounded-lg border px-3 py-2 shadow-sm"
    >
      <div className="mb-1 flex min-w-0 items-center justify-between gap-2">
        <span className="min-w-0 truncate font-medium">
          {state.track.fileName}
        </span>
        <div className="flex items-center gap-1">
          {onViewMessage ? (
            <Tooltip title="Return to message">
              <Button
                aria-label="Return to audio message"
                asIcon
                color="tertiary"
                onClick={() => {
                  if (state.track) onViewMessage(state.track.messageId);
                }}
                size="xs"
                type="button"
                variant="naked"
              >
                <ReplyIcon aria-hidden className="size-4" />
              </Button>
            </Tooltip>
          ) : null}
          <Tooltip title="Stop audio">
            <Button
              aria-label="Stop audio playback"
              asIcon
              color="tertiary"
              onClick={controller.stop}
              size="xs"
              type="button"
              variant="naked"
            >
              <CloseIcon aria-hidden className="size-4" />
            </Button>
          </Tooltip>
        </div>
      </div>
      <AudioPlayerControls
        controller={controller}
        state={state}
        track={state.track}
      />
      {state.error ? (
        <p className="text-danger mt-2" role="alert">
          {state.error}
        </p>
      ) : null}
    </section>
  );
};
