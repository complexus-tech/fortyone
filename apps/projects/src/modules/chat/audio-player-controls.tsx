import type { ReactNode } from "react";
import { LoadingIcon, PauseIcon } from "icons";
import { Button, Tooltip } from "ui";
import type {
  AudioPlaybackController,
  AudioPlaybackState,
  AudioTrack,
} from "./audio-playback-controller";
import { AudioWaveform } from "./audio-waveform";
import { useAudioWaveform } from "./audio-waveform-context";

export const formatAudioTime = (seconds: number) => {
  const time = Math.max(0, Math.floor(Number.isFinite(seconds) ? seconds : 0));
  return `${Math.floor(time / 60)}:${String(time % 60).padStart(2, "0")}`;
};

const PlayMark = () => (
  <svg aria-hidden className="size-4.5" fill="currentColor" viewBox="0 0 24 24">
    <path d="M7.5 4.9c-.9-.5-2 .1-2 1.1v12c0 1 1.1 1.6 2 1.1l10.4-6c.9-.5.9-1.8 0-2.3L7.5 4.9Z" />
  </svg>
);

export const AudioPlayerControls = ({
  track,
  state,
  controller,
  transcriptControl,
  moreActions,
}: {
  track: AudioTrack;
  state: AudioPlaybackState;
  controller: AudioPlaybackController;
  transcriptControl?: ReactNode;
  moreActions?: ReactNode;
}) => {
  const { ref, data: waveform } = useAudioWaveform(track);
  const active = state.track?.id === track.id;
  const playing =
    active && (state.status === "playing" || state.status === "loading");
  const duration =
    (active && state.duration) ||
    track.durationSeconds ||
    waveform?.duration ||
    0;
  const currentTime = active ? state.currentTime : 0;
  const rate = active ? state.rate : 1;
  const progress = duration > 0 ? currentTime / duration : 0;
  let timeLabel = duration > 0 ? formatAudioTime(duration) : "—:——";
  if (playing || (currentTime > 0 && state.status !== "ended"))
    timeLabel = formatAudioTime(currentTime);
  let playIcon = <PlayMark />;
  if (active && state.status === "loading")
    playIcon = <LoadingIcon aria-hidden className="size-4.5 animate-spin" />;
  else if (playing) playIcon = <PauseIcon aria-hidden className="size-4" />;
  return (
    <div className="flex min-w-0 items-center gap-2">
      <Tooltip title={playing ? "Pause" : "Play"}>
        <Button
          aria-label={`${playing ? "Pause" : "Play"} ${track.fileName}`}
          asIcon
          className="bg-primary text-primary-foreground hover:bg-primary/90 size-9 shrink-0 [&_svg]:text-current"
          color="tertiary"
          onClick={() => {
            if (playing) controller.pause();
            else void controller.play({ ...track, durationSeconds: duration });
          }}
          rounded="full"
          size="sm"
          type="button"
          variant="naked"
        >
          {playIcon}
        </Button>
      </Tooltip>
      <div className="min-w-0 flex-1" ref={ref}>
        <div className="has-focus-visible:ring-primary/40 relative flex h-10 items-center rounded-sm has-focus-visible:ring-2">
          <AudioWaveform peaks={waveform?.peaks} progress={progress} />
          <input
            aria-label={`Seek ${track.fileName}`}
            aria-valuetext={`${formatAudioTime(currentTime)} of ${formatAudioTime(duration)}`}
            className="absolute inset-0 m-0 h-full w-full cursor-pointer opacity-0 disabled:cursor-default"
            disabled={!active || !state.canSeek}
            max={duration || 1}
            min={0}
            onChange={(event) => {
              controller.seek(Number(event.target.value));
            }}
            step={0.1}
            type="range"
            value={currentTime}
          />
        </div>
      </div>
      <span
        aria-hidden
        className="text-text-muted min-w-9 shrink-0 text-right tabular-nums"
      >
        {timeLabel}
      </span>
      {transcriptControl}
      <Tooltip title="Playback speed">
        <Button
          aria-label={`Playback speed ${rate} times`}
          className="min-w-7 shrink-0 px-1 tabular-nums"
          color="tertiary"
          disabled={!active}
          onClick={() => {
            const rates = [1, 1.5, 2];
            controller.setRate(rates[(rates.indexOf(rate) + 1) % rates.length]);
          }}
          size="xs"
          type="button"
          variant="naked"
        >
          {rate}×
        </Button>
      </Tooltip>
      {moreActions}
    </div>
  );
};
