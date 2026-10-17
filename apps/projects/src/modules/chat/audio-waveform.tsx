import { useId, useMemo } from "react";

const WAVEFORM_WIDTH = 216;
const WAVEFORM_HEIGHT = 36;
const EMPTY_PEAKS = Array.from({ length: 48 }, () => 0);

export const AudioWaveform = ({
  peaks,
  progress,
}: {
  peaks?: readonly number[];
  progress: number;
}) => {
  const clipId = useId();
  const path = useMemo(() => {
    const values = peaks?.length ? peaks : EMPTY_PEAKS;
    const step = WAVEFORM_WIDTH / values.length;
    return values
      .map((peak, index) => {
        const amplitude = Number.isFinite(peak)
          ? Math.max(0, Math.min(1, peak))
          : 0;
        const height = 2 + amplitude * (WAVEFORM_HEIGHT - 6);
        const x = (index + 0.5) * step;
        return `M${x} ${(WAVEFORM_HEIGHT - height) / 2}v${height}`;
      })
      .join(" ");
  }, [peaks]);
  return (
    <svg
      aria-hidden
      className="pointer-events-none h-9 w-full overflow-visible"
      data-waveform={peaks?.length ? "ready" : "placeholder"}
      fill="none"
      preserveAspectRatio="none"
      viewBox={`0 0 ${WAVEFORM_WIDTH} ${WAVEFORM_HEIGHT}`}
    >
      <defs>
        <clipPath id={clipId}>
          <rect
            height={WAVEFORM_HEIGHT}
            width={WAVEFORM_WIDTH * Math.max(0, Math.min(1, progress))}
          />
        </clipPath>
      </defs>
      <path
        className="text-primary/35 dark:text-primary/40"
        d={path}
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth={2.5}
      />
      <path
        className="text-primary"
        clipPath={`url(#${clipId})`}
        d={path}
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth={2.5}
      />
    </svg>
  );
};
