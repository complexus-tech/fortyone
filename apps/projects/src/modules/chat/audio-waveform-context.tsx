"use client";

import type { ReactNode } from "react";
import { createContext, useContext, useEffect, useRef, useState } from "react";
import type { AudioTrack } from "./audio-playback-controller";
import type {
  AudioWaveformData,
  AudioWaveformStore,
} from "./audio-waveform-data";
import { createAudioWaveformStore } from "./audio-waveform-data";

const WaveformContext = createContext<AudioWaveformStore | null>(null);

export const AudioWaveformProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  // eslint-disable-next-line react/hook-use-state -- One stable cache for the provider lifetime; it is never replaced.
  const [store] = useState(createAudioWaveformStore);
  useEffect(
    () => () => {
      store.destroy();
    },
    [store],
  );
  return (
    <WaveformContext.Provider value={store}>
      {children}
    </WaveformContext.Provider>
  );
};

export const useAudioWaveform = ({
  id,
  url: playbackUrl,
  waveformUrl,
  durationSeconds,
}: AudioTrack) => {
  const url = waveformUrl ?? playbackUrl;
  const store = useContext(WaveformContext);
  const ref = useRef<HTMLDivElement>(null);
  const [result, setResult] = useState<{
    id: string;
    url: string;
    data: AudioWaveformData | null;
  } | null>(null);

  useEffect(() => {
    const element = ref.current;
    if (!store || !element) return;
    let request: AbortController | undefined;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let finished = false;
    const cancel = () => {
      clearTimeout(timer);
      request?.abort();
      request = undefined;
    };
    const load = () => {
      if (request || finished) return;
      const current = new AbortController();
      request = current;
      // A short dwell avoids downloading notes passed during a fast scroll.
      timer = setTimeout(() => {
        void store
          .load({ id, url, durationSeconds }, current.signal)
          .then((data) => {
            if (current.signal.aborted) return;
            finished = true;
            setResult({ id, url, data });
          });
      }, 200);
    };
    const observer =
      typeof IntersectionObserver === "undefined"
        ? null
        : new IntersectionObserver((entries) => {
            if (entries.some((entry) => entry.isIntersecting)) load();
            else cancel();
          });
    if (observer) observer.observe(element);
    else load();
    return () => {
      clearTimeout(timer);
      request?.abort();
      observer?.disconnect();
    };
  }, [durationSeconds, id, store, url]);

  return {
    ref,
    data: result?.id === id && result.url === url ? result.data : null,
  };
};
