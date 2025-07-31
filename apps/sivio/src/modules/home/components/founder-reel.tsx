"use client";
import { useRef, useState, useEffect } from "react";
import { Box, Flex, Text } from "ui";
import Link from "next/link";
import { Container } from "@/components/ui";

const PlayIcon = () => (
  <Box className="flex items-center">
    <svg fill="none" height="64" viewBox="0 0 64 64" width="64">
      <polygon fill="white" points="0,0 32,32 0,64" />
      <polygon fill="white" points="16,0 48,32 16,64" />
      <polygon fill="white" points="32,0 64,32 32,64" />
    </svg>
  </Box>
);

const PreferToReadBar = () => (
  <Container className="absolute bottom-0 left-0 right-0 z-10">
    <Link
      className="flex w-full cursor-pointer items-center gap-6 border-t border-gray-200 bg-white px-8 py-5"
      href="/"
    >
      <Text className="text-2xl font-semibold text-black">Prefer to read</Text>
      <Box className="flex gap-2">
        <Box className="rounded-full bg-black p-2">
          <svg fill="none" height="20" viewBox="0 0 20 20" width="20">
            <path
              d="M7 5l5 5-5 5"
              stroke="white"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
            />
          </svg>
        </Box>
        <Box className="rounded-full bg-black p-2">
          <svg fill="none" height="20" viewBox="0 0 20 20" width="20">
            <path
              d="M7 5l5 5-5 5"
              stroke="white"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
            />
          </svg>
        </Box>
      </Box>
    </Link>
  </Container>
);

export const FounderReel = () => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;
    const handlePlay = () => {
      setIsPlaying(true);
    };
    const handlePause = () => {
      setIsPlaying(false);
    };
    const handleEnded = () => {
      setIsPlaying(false);
    };
    video.addEventListener("play", handlePlay);
    video.addEventListener("pause", handlePause);
    video.addEventListener("ended", handleEnded);
    return () => {
      video.removeEventListener("play", handlePlay);
      video.removeEventListener("pause", handlePause);
      video.removeEventListener("ended", handleEnded);
    };
  }, []);

  const handleOverlayPlay = () => {
    setIsPlaying(true);
    void videoRef.current?.play();
  };

  return (
    <Box>
      <Box className="relative aspect-video w-full overflow-hidden shadow-lg">
        <video
          className="h-full w-full object-cover"
          controls={isPlaying}
          muted
          ref={videoRef}
          src="/videos/founder-reel.mp4"
          tabIndex={-1}
        />
        {!isPlaying && (
          <Flex
            aria-label="Play founder reel"
            className="absolute inset-0 cursor-pointer items-center justify-center bg-black/50 transition"
            onClick={handleOverlayPlay}
            role="button"
            tabIndex={0}
          >
            <Flex className="items-center gap-8 bg-gray-100/40 px-8 py-6 backdrop-blur">
              <Flex align="center" className="mr-6" direction="column">
                <PlayIcon />
                <Text className="mt-2 text-lg font-semibold text-white">
                  Play reel
                </Text>
              </Flex>
              <Text className="text-5xl font-black uppercase tracking-wide text-white md:text-6xl">
                MEET OUR FOUNDER
              </Text>
            </Flex>
          </Flex>
        )}
        {!isPlaying ? <PreferToReadBar /> : null}
      </Box>
    </Box>
  );
};
