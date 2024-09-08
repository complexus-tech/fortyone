"use client";
import { BlurImage, Box, Text } from "ui";
import { Container } from "@/components/ui";
import { Blur } from "@/components/ui/blur";
import { useCursor } from "@/hooks";

export const Story = () => {
  const cursor = useCursor();
  const features: { heading: string; description: string }[] = [
    {
      heading: "Automation",
      description: "Automate workflows to save time and increase productivity.",
    },
    {
      heading: "Customizability",
      description:
        "Tailor the platform to fit your team's unique requirements.",
    },
    {
      heading: "Whiteboards",
      description:
        "Visualize ideas and plans effortlessly with advanced whiteboard functionality.",
    },
    {
      heading: "Discussions",
      description:
        "Engage in meaningful discussions and keep everyone informed and aligned.",
    },
  ];

  return (
    <Container className="max-w-3xl">
      <Text
        className="mb-4 text-xl uppercase tracking-wide md:mb-6 md:md:w-1/2 md:text-2xl"
        color="gradient"
      >
        Our Mission
      </Text>
      <Text
        className="mb-20 max-w-6xl text-xl leading-snug opacity-80 md:mb-32 md:text-2xl"
        fontWeight="normal"
      >
        Our mission is to provide a seamless, efficient, and accessible online
        image conversion solution. We strive to empower users worldwide with the
        ability to transform their images effortlessly, supporting creativity
        and productivity across various platforms and devices.
      </Text>

      <Text
        className="mb-4 text-xl uppercase tracking-wide md:mb-6 md:w-1/2 md:text-2xl"
        color="gradient"
      >
        Meet the founder
      </Text>
      <Text
        className="mb-10 max-w-6xl text-xl leading-snug opacity-80 md:text-2xl"
        fontWeight="normal"
      >
        Greetings! I&rsquo;m Joseph, the creator of ImageConveta. With a passion
        for simplifying digital workflows, I set out to build a tool that makes
        image conversion accessible to everyone.
      </Text>
      <Box
        onMouseEnter={() => {
          cursor.setText("Hello 😉");
        }}
        onMouseLeave={() => {
          cursor.removeText();
        }}
        className="mb-36"
      >
        <BlurImage
          className="pointer-events-none aspect-square rounded-2xl object-bottom grayscale"
          src="/joseph.webp"
        />
      </Box>
    </Container>
  );
};
