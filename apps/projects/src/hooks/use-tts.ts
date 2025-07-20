import { useState, useRef, useCallback, useEffect } from "react";
import { toast } from "sonner";
import ky from "ky";
import type { TTSVoice } from "@/types/tts";

type AudioState = "idle" | "loading" | "playing" | "paused" | "ended";

type AudioItem = {
  text: string;
  voice: TTSVoice;
  state: AudioState;
  audioUrl?: string;
};

export const useTTS = () => {
  const [currentAudio, setCurrentAudio] = useState<AudioItem | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const audioCache = useRef<Map<string, string>>(new Map());

  const generateAudio = useCallback(
    async (text: string, voice: TTSVoice = "alloy"): Promise<string> => {
      const cacheKey = `${text}-${voice}`;

      // Check cache first
      if (audioCache.current.has(cacheKey)) {
        return audioCache.current.get(cacheKey)!;
      }

      try {
        const response = await ky
          .post("/api/tts", {
            json: { text, voice },
          })
          .arrayBuffer();

        const blob = new Blob([response], { type: "audio/wav" });
        const audioUrl = URL.createObjectURL(blob);

        // Cache the audio URL
        audioCache.current.set(cacheKey, audioUrl);

        return audioUrl;
      } catch (error) {
        throw new Error("Failed to generate audio");
      }
    },
    [],
  );

  // Initialize audio element
  useEffect(() => {
    if (!audioRef.current) {
      audioRef.current = new Audio();
      audioRef.current.volume = 1.0;
      audioRef.current.preload = "auto";

      const handleEnded = () => {
        setCurrentAudio(null);
      };

      const handleError = (_event: Event) => {
        setCurrentAudio(null);
        setIsLoading(false);
        toast.warning("Audio playback failed", {
          description: "Please try again.",
          duration: 3000,
        });
      };

      const handleLoadStart = () => {
        setCurrentAudio((prev) =>
          prev ? { ...prev, state: "loading" } : null,
        );
      };

      const handleCanPlay = () => {
        setCurrentAudio((prev) =>
          prev ? { ...prev, state: "playing" } : null,
        );
      };

      audioRef.current.addEventListener("ended", handleEnded);
      audioRef.current.addEventListener("error", handleError);
      audioRef.current.addEventListener("loadstart", handleLoadStart);
      audioRef.current.addEventListener("canplay", handleCanPlay);
    }

    return () => {
      if (audioRef.current) {
        audioRef.current.pause();
      }
    };
  }, []);

  const playText = useCallback(
    async (text: string, voice: TTSVoice = "alloy") => {
      // Stop any current playback
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current.currentTime = 0;
      }

      setIsLoading(true);
      setCurrentAudio({
        text,
        voice,
        state: "loading",
      });

      // Safety timeout to clear loading state if something goes wrong
      const timeoutId = setTimeout(() => {
        setIsLoading(false);
        setCurrentAudio(null);
      }, 30000); // 30 seconds timeout

      try {
        const audioUrl = await generateAudio(text, voice);

        setCurrentAudio({
          text,
          voice,
          state: "loading",
          audioUrl,
        });

        if (audioRef.current) {
          audioRef.current.src = audioUrl;
          await audioRef.current.play();
        }
      } catch (error) {
        toast.warning("Failed to play audio", {
          description: "Please try again.",
          duration: 3000,
        });
        setCurrentAudio(null);
      } finally {
        clearTimeout(timeoutId);
        setIsLoading(false);
      }
    },
    [generateAudio],
  );

  const pause = useCallback(() => {
    if (audioRef.current && currentAudio?.state === "playing") {
      audioRef.current.pause();
      setCurrentAudio((prev) => (prev ? { ...prev, state: "paused" } : null));
    }
  }, [currentAudio?.state]);

  const resume = useCallback(() => {
    if (audioRef.current && currentAudio?.state === "paused") {
      audioRef.current.play();
      setCurrentAudio((prev) => (prev ? { ...prev, state: "playing" } : null));
    }
  }, [currentAudio?.state]);

  const stop = useCallback(() => {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current.currentTime = 0;
    }
    setCurrentAudio(null);
    setIsLoading(false);
  }, []);

  const clearCache = useCallback(() => {
    // Revoke all cached audio URLs to free memory
    audioCache.current.forEach((url) => {
      URL.revokeObjectURL(url);
    });
    audioCache.current.clear();
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      clearCache();
    };
  }, [clearCache]);

  return {
    // State
    currentAudio,
    isLoading,

    // Actions
    playText,
    pause,
    resume,
    stop,
    clearCache,

    // Computed
    isPlaying: currentAudio?.state === "playing",
    isPaused: currentAudio?.state === "paused",
  };
};
