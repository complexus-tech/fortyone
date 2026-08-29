import type { ReactNode } from "react";
import Image from "next/image";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import meshImage from "../../../public/images/meshing.webp";

export const FEATURE_STORY_TEXT_CLASS = "text-[0.9rem] leading-[1.35]";
export const FEATURE_STORY_META_TEXT_CLASS = "text-[0.82rem] leading-[1.25]";
export const FEATURE_STORY_SURFACE_CLASS =
  "bg-surface-elevated rounded-xl border border-border/80 shadow-lg shadow-shadow";

type FeatureStoryCardProps = {
  children: ReactNode;
  delay?: number;
  description: string;
  title: string;
};

export const FeatureStoryCard = ({
  children,
  delay = 0,
  description,
  title,
}: FeatureStoryCardProps) => {
  return (
    <Box
      className="h-full"
      data-landing-reveal
      style={{ transitionDelay: `${delay * 1000}ms` }}
    >
      <Box className="flex h-full flex-col">
        <Box className="relative flex shrink-0 items-end overflow-hidden rounded-2xl md:h-84">
          <Image
            alt=""
            className="object-cover grayscale-100 dark:opacity-40"
            fill
            sizes="(max-width: 767px) 100vw, (max-width: 1279px) 50vw, 25vw"
            src={meshImage}
          />
          <Box className="relative z-10 w-full p-5">{children}</Box>
        </Box>
        <Box className="mt-5 flex flex-col">
          <Text className="text-foreground mb-2 text-lg font-semibold">
            {title}
          </Text>
          <Text className="text-text-muted">{description}</Text>
        </Box>
      </Box>
    </Box>
  );
};

type FeatureStorySectionProps = {
  children: ReactNode;
  heading: ReactNode;
  id?: string;
};

export const FeatureStorySection = ({
  children,
  heading,
  id,
}: FeatureStorySectionProps) => {
  return (
    <Container className="scroll-mt-24 py-16 md:pt-36" id={id}>
      <Box data-landing-reveal>
        <Text as="h2" className="mb-14 max-w-3xl pb-2 text-3xl md:text-5xl">
          {heading}
        </Text>
      </Box>

      <Box className="grid grid-cols-1 gap-6 md:auto-rows-fr md:grid-cols-3">
        {children}
      </Box>
    </Container>
  );
};
