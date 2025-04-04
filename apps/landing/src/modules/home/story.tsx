"use client";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const Story = () => {
  const features: { heading: string; description: string }[] = [
    {
      heading: "Intuitive User Experience",
      description:
        "Beautiful interface that can be learned in a day. No extensive training required.",
    },
    {
      heading: "Customizable Terminology",
      description:
        "Make the platform speak your language with custom terminology.",
    },
    {
      heading: "Flexible Workflows",
      description:
        "Create custom workflows for each team while maintaining cross-organization visibility.",
    },
    {
      heading: "Visual Task Management",
      description:
        "Intuitive Kanban boards for real-time work visualization and resource optimization.",
    },
    {
      heading: "OKR Management",
      description:
        "Set and track Objectives and Key Results (OKRs) to align teams with your vision.",
    },
    {
      heading: "Team Collaboration",
      description:
        "Foster teamwork with shared objectives and integrated workspaces that break down silos.",
    },
  ];

  return (
    <Container className="mt-16 md:mt-28">
      <Box className="relative">
        <Text
          as="h1"
          className="mb-8 h-max max-w-5xl pb-2 text-5xl font-semibold md:mb-10 md:mt-6 md:text-7xl"
          color="gradient"
        >
          Why Choose Complexus
        </Text>
        <Text
          className=" mb-12 max-w-5xl text-xl leading-snug opacity-80 md:mb-20 md:text-2xl"
          fontWeight="normal"
        >
          Simplify complexity without sacrifice. Complexus adapts to your
          team&apos;s needs while keeping everyone aligned with strategic goals.
        </Text>
        <Box className="mb-16 grid grid-cols-1 divide-x divide-y divide-dashed divide-gray-200/10 border border-gray-200/10 md:mb-32 md:grid-cols-3">
          {features.map(({ heading, description }, idx) => (
            <Box className="p-8 md:px-12 md:py-16" key={heading}>
              <Text className="mb-4 text-6xl opacity-20 md:text-8xl">
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
      </Box>
    </Container>
  );
};
