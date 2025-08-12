"use client";

import { Text, Box } from "ui";
import { Container } from "@/components/ui";

export const Maya = () => {
  return (
    <Container className="dark:py-10">
      <Text
        className="mb-16 max-w-5xl text-6xl font-semibold leading-[1.15] md:mb-20"
        color="gradientDark"
      >
        Your AI assistant, helps your team plan sprints, track objectives, and
        catch bottlenecks before they slow you down.
      </Text>

      {/* <Box className="rounded-[0.6rem] border border-[#8080802a] bg-white/50 p-0.5 backdrop-blur dark:border-dark-50/70 dark:bg-dark-200/40 md:rounded-3xl md:p-1.5">
        <video
          autoPlay
          className="aspect-[16/10] h-full w-full object-cover md:rounded-[1.1rem]"
          loop
          muted
          src="/videos/test.mp4"
        />
      </Box> */}

      <video
        autoPlay
        className="aspect-[16/10] h-full w-full object-cover md:rounded-3xl"
        loop
        muted
        src="/videos/test.mp4"
      />

      <Box className="mt-20 grid grid-cols-2 gap-16">
        <Box>
          <Text className="mb-5 text-4xl font-semibold">
            AI that works like a teammate
          </Text>
          <Text className="max-w-lg text-lg" color="muted">
            Maya keeps projects moving, anticipating next steps and removing
            blockers before they slow you down.
          </Text>
          <Box className="mt-8 h-48 rounded-3xl bg-gray-100 dark:bg-dark-300/80" />
        </Box>

        <Box>
          <Text className="mb-5 text-4xl font-semibold">
            Plans that write themselves
          </Text>
          <Text className="max-w-lg text-lg" color="muted">
            Turn rough ideas into fully scoped sprints in seconds, with AI
            handling the details.
          </Text>
          <Box className="mt-8 h-48 rounded-3xl bg-gray-100 dark:bg-dark-300/80" />
        </Box>
      </Box>
    </Container>
  );
};
