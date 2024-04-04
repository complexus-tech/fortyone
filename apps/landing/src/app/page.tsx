"use client";
import { Box } from "ui";
import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
} from "@/components/pages/home";
import { Blur } from "@/components/ui";

export default function Page(): JSX.Element {
  return (
    <Box className="relative">
      <Blur className="absolute -top-[70vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 dark:bg-primary/5" />
      <Hero />
      <HeroCards />
      <SampleClients />
      <Features />
      <Integrations />
    </Box>
  );
}
