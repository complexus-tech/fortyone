"use client";

import { Text, Box } from "ui";
import { Container } from "@/components/ui";

export const Maya = () => {
  return (
    <Container className="dark:py-10">
      <Text
        className="mb-16 max-w-5xl text-5xl font-semibold leading-[1.15]"
        color="gradientDark"
      >
        Maya, your AI assistant, helps your team plan sprints, track objectives,
        and catch bottlenecks before they slow you down.
      </Text>

      <Box className="rounded-[0.6rem] border border-[#8080802a] bg-white/50 p-0.5 backdrop-blur dark:border-dark-50/70 dark:bg-dark-200/40 md:rounded-3xl md:p-1.5">
        <video
          autoPlay
          className="aspect-[16/10] h-full w-full object-cover md:rounded-[1.1rem]"
          loop
          muted
          src="/videos/test.mp4"
        />
      </Box>
    </Container>
  );
};
