export type AudioTrack = {
  id: string;
  messageId: string;
  fileName: string;
  url: string;
  waveformUrl?: string;
  durationSeconds?: number;
};

export type AudioPlaybackState = {
  track: AudioTrack | null;
  status: "idle" | "loading" | "playing" | "paused" | "ended" | "error";
  currentTime: number;
  duration: number;
  rate: number;
  canSeek: boolean;
  error: string | null;
  sourceVisible: boolean;
};

export const IDLE_AUDIO_PLAYBACK: AudioPlaybackState = {
  track: null,
  status: "idle",
  currentTime: 0,
  duration: 0,
  rate: 1,
  canSeek: false,
  error: null,
  sourceVisible: false,
};

const validDuration = (value?: number) =>
  value && Number.isFinite(value) && value > 0 ? value : 0;

export const createAudioPlaybackController = (
  createAudio: () => HTMLAudioElement = () => new Audio(),
) => {
  let audio: HTMLAudioElement | null = null;
  let state = IDLE_AUDIO_PLAYBACK;
  let playRequest = 0;
  let removeListeners: (() => void) | null = null;
  const listeners = new Set<() => void>();
  const externalPauses = new Set<() => void>();
  const visibleRows = new Map<string, Set<symbol>>();

  const publish = (patch: Partial<AudioPlaybackState>) => {
    const next = { ...state, ...patch };
    if (
      Object.keys(patch).every(
        (key) =>
          next[key as keyof AudioPlaybackState] ===
          state[key as keyof AudioPlaybackState],
      )
    )
      return;
    state = next;
    listeners.forEach((listener) => {
      listener();
    });
  };
  const updateTime = () => {
    if (!audio || !state.track) return;
    const duration =
      validDuration(audio.duration) ||
      validDuration(state.track.durationSeconds);
    publish({
      duration,
      currentTime: Math.max(
        0,
        Number.isFinite(audio.currentTime) ? audio.currentTime : 0,
      ),
      canSeek: audio.readyState >= 1 && duration > 0,
    });
  };
  const ensureAudio = () => {
    if (audio) return audio;
    const element = createAudio();
    element.preload = "none";
    audio = element;
    const playing = () => {
      if (state.track && !element.paused)
        publish({ status: "playing", error: null });
    };
    const waiting = () => {
      if (!element.paused) publish({ status: "loading" });
    };
    const paused = () => {
      if (state.track && state.status !== "error" && !element.ended)
        publish({ status: "paused" });
    };
    const ended = () => {
      if (!element.ended) return;
      updateTime();
      publish({ status: "ended" });
    };
    const failed = () => {
      if (!state.track) return;
      playRequest++;
      publish({
        status: "error",
        canSeek: false,
        error:
          element.error?.code === 4
            ? "This browser cannot play this recording. Download the audio to listen."
            : "Audio could not be loaded. Try playing it again.",
      });
    };
    const events: [string, () => void][] = [
      ["playing", playing],
      ["waiting", waiting],
      ["pause", paused],
      ["ended", ended],
      ["error", failed],
      ["loadedmetadata", updateTime],
      ["durationchange", updateTime],
      ["timeupdate", updateTime],
    ];
    events.forEach(([name, handler]) => {
      element.addEventListener(name, handler);
    });
    removeListeners = () => {
      events.forEach(([name, handler]) => {
        element.removeEventListener(name, handler);
      });
    };
    return element;
  };
  const pause = () => {
    playRequest++;
    audio?.pause();
    if (state.track && state.status !== "error" && state.status !== "ended")
      publish({ status: "paused" });
  };
  const stop = () => {
    playRequest++;
    if (audio) {
      audio.pause();
      audio.removeAttribute("src");
      audio.load();
    }
    publish({ ...IDLE_AUDIO_PLAYBACK, rate: state.rate });
  };

  return {
    getSnapshot: () => state,
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    play: async (track: AudioTrack) => {
      externalPauses.forEach((pauseExternal) => {
        pauseExternal();
      });
      const element = ensureAudio();
      const request = ++playRequest;
      const sameTrack =
        state.track?.id === track.id && state.track.url === track.url;
      if (!sameTrack || state.status === "error") {
        element.pause();
        element.src = track.url;
        element.load();
        publish({
          track,
          status: "loading",
          currentTime: 0,
          duration: validDuration(track.durationSeconds),
          canSeek: false,
          error: null,
          sourceVisible: Boolean(visibleRows.get(track.id)?.size),
        });
      } else {
        if (state.status === "ended") element.currentTime = 0;
        publish({ status: "loading", error: null });
      }
      element.playbackRate = state.rate;
      try {
        await element.play();
        if (request === playRequest)
          publish({ status: element.paused ? "paused" : "playing" });
      } catch {
        // A superseded play/pause or source switch commonly rejects play().
        if (request === playRequest)
          publish({
            status: "error",
            error: "Playback could not start. Try playing it again.",
          });
      }
    },
    pause,
    stop,
    registerExternalPause: (pauseExternal: () => void) => {
      externalPauses.add(pauseExternal);
      return () => {
        externalPauses.delete(pauseExternal);
      };
    },
    seek: (seconds: number) => {
      if (!audio || !state.canSeek || !Number.isFinite(seconds)) return;
      const time = Math.max(0, Math.min(state.duration, seconds));
      audio.currentTime = time;
      publish({
        currentTime: time,
        ...(state.status === "ended" ? { status: "paused" } : {}),
      });
    },
    setRate: (rate: number) => {
      if (![1, 1.5, 2].includes(rate)) return;
      if (audio) audio.playbackRate = rate;
      publish({ rate });
    },
    setRowVisible: (id: string, token: symbol, visible: boolean) => {
      const rows = visibleRows.get(id) ?? new Set<symbol>();
      if (visible) rows.add(token);
      else rows.delete(token);
      if (rows.size) visibleRows.set(id, rows);
      else visibleRows.delete(id);
      if (state.track?.id === id) publish({ sourceVisible: rows.size > 0 });
    },
    invalidateMessage: (messageId: string) => {
      if (state.track?.messageId === messageId) stop();
    },
    destroy: () => {
      removeListeners?.();
      removeListeners = null;
      stop();
      audio = null;
      visibleRows.clear();
    },
  };
};

export type AudioPlaybackController = ReturnType<
  typeof createAudioPlaybackController
>;
