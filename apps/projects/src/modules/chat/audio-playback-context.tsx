"use client";

import type { ReactNode } from "react";
import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { AudioPlaybackController } from "./audio-playback-controller";
import {
  createAudioPlaybackController,
  IDLE_AUDIO_PLAYBACK,
} from "./audio-playback-controller";
import { AudioWaveformProvider } from "./audio-waveform-context";

const PlaybackContext = createContext<AudioPlaybackController | null>(null);
const NavigationContext = createContext<
  ((messageId: string) => void) | undefined
>(undefined);

export const AudioPlaybackProvider = ({
  children,
  onViewMessage,
}: {
  children: ReactNode;
  onViewMessage?: (messageId: string) => void;
}) => {
  // eslint-disable-next-line react/hook-use-state -- Playback owns one stable controller for this provider lifetime.
  const [controller] = useState(createAudioPlaybackController);
  useEffect(
    () => () => {
      controller.destroy();
    },
    [controller],
  );
  return (
    <PlaybackContext.Provider value={controller}>
      <NavigationContext.Provider value={onViewMessage}>
        <AudioWaveformProvider>{children}</AudioWaveformProvider>
      </NavigationContext.Provider>
    </PlaybackContext.Provider>
  );
};

export const useAudioPlaybackController = () => {
  const controller = useContext(PlaybackContext);
  if (!controller)
    throw new Error("Audio playback requires AudioPlaybackProvider");
  return controller;
};

export const useAudioPlaybackState = (attachmentId?: string) => {
  const controller = useAudioPlaybackController();
  return useSyncExternalStore(
    controller.subscribe,
    () => {
      const state = controller.getSnapshot();
      return !attachmentId || state.track?.id === attachmentId
        ? state
        : IDLE_AUDIO_PLAYBACK;
    },
    () => IDLE_AUDIO_PLAYBACK,
  );
};

export const useAudioPlaybackNavigation = () => useContext(NavigationContext);

export const useAudioSourceVisibility = (attachmentId: string) => {
  const controller = useAudioPlaybackController();
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const element = ref.current;
    if (!element) return;
    const token = Symbol(attachmentId);
    if (typeof IntersectionObserver === "undefined") {
      controller.setRowVisible(attachmentId, token, true);
      return () => {
        controller.setRowVisible(attachmentId, token, false);
      };
    }
    const observer = new IntersectionObserver((entries) => {
      controller.setRowVisible(
        attachmentId,
        token,
        entries.some((entry) => entry.isIntersecting),
      );
    });
    observer.observe(element);
    return () => {
      observer.disconnect();
      controller.setRowVisible(attachmentId, token, false);
    };
  }, [attachmentId, controller]);
  return ref;
};
