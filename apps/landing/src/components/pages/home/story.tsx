"use client";
import { Box, Text } from "ui";
import { cn } from "lib";
import { Container } from "@/components/ui";
import { Blur } from "@/components/ui/blur";

export const Story = () => {
  const features: { heading: string; description: string }[] = [
    {
      heading: "OKR Management",
      description:
        "Set and track Objectives and Key Results (OKRs) with powerful tools for measuring progress and aligning teams.",
    },
    {
      heading: "Strategic Roadmapping",
      description:
        "Plan and visualize your organization's journey with interactive roadmaps and milestone tracking.",
    },
    {
      heading: "Team Collaboration",
      description:
        "Foster cross-functional teamwork with shared objectives, real-time updates, and integrated team workspaces.",
    },
    {
      heading: "Progress Tracking",
      description:
        "Monitor objective health, track key results, and get insights into team performance with comprehensive analytics.",
    },
  ];

  return (
    <Container className="mt-16 max-w-5xl md:mt-28">
      <Box className="relative">
        <Text
          align="center"
          as="h1"
          className="mb-8 h-max max-w-4xl pb-2 text-5xl font-semibold md:mb-10 md:mt-6 md:text-7xl"
          color="gradient"
        >
          Better Tools, Bigger Impact
        </Text>
        <Text
          align="center"
          className="mb-12 max-w-6xl text-xl leading-snug opacity-80 md:mb-20 md:text-2xl"
          fontWeight="normal"
        >
          Your strategy deserves better than spreadsheets and slides. Complexus
          delivers a complete platform for OKR management, strategic planning,
          and team alignment that turns great plans into greater achievements.
        </Text>
        <Box className="mb-16 grid grid-cols-1 gap-10 md:mb-32 md:grid-cols-2 md:gap-x-20 md:gap-y-16">
          {features.map(({ heading, description }, idx) => (
            <Box
              className={cn(
                "border-t border-gray-200/10 pt-8 text-center md:pt-12",
                {
                  "border-b pb-12 md:border-b-0 md:pb-0":
                    idx === features.length - 1,
                },
              )}
              key={heading}
            >
              <Text className="mb-4 text-6xl opacity-30 md:text-8xl">
                {idx + 1}.
              </Text>
              <Text fontSize="xl" fontWeight="semibold" transform="uppercase">
                {heading}
              </Text>
              <Text
                className="mt-4 opacity-80"
                fontSize="lg"
                fontWeight="normal"
              >
                {description}
              </Text>
            </Box>
          ))}
        </Box>
        <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[700px] w-[700px] -translate-x-1/2 -translate-y-1/2 bg-warning/10" />
      </Box>
    </Container>
  );
};
