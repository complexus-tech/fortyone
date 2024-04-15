"use client";
import { BlurImage, Box, Text, Wrapper } from "ui";
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
    <Container className="max-w-4xl">
      <Text
        className="mb-4 text-xl uppercase tracking-wide md:mb-6 md:md:w-1/2 md:text-2xl"
        color="gradient"
      >
        Our Mission
      </Text>
      <Text
        fontWeight="normal"
        className="mb-20 max-w-6xl text-xl leading-snug opacity-80 md:mb-32 md:text-2xl"
      >
        Our mission is to provide a user-friendly, feature-rich project
        management solution that streamlines workflows and enhances
        productivity. Whether you're a small startup, a growing business, or a
        seasoned enterprise, we're here to support your project management needs
        every step of the way.
      </Text>

      <Text
        className="mb-4 text-xl uppercase tracking-wide md:mb-6 md:w-1/2 md:text-2xl"
        color="gradient"
      >
        Meet the founder
      </Text>
      <Text
        fontWeight="normal"
        className="mb-10 max-w-6xl text-xl leading-snug opacity-80 md:text-2xl"
      >
        Greetings! I'm Joseph, the creator of Complexus. With a passion for
        simplifying project management and a commitment to empowering teams, I
        embarked on this journey to develop a platform that caters to the
        diverse needs of modern projects. Every aspect of our system, from its
        design to its functionality, is meticulously crafted to deliver an
        exceptional user experience.
      </Text>
      <Box
        onMouseEnter={() => {
          cursor.setText("Hello 😉");
        }}
        onMouseLeave={() => {
          cursor.removeText();
        }}
      >
        <BlurImage
          src="/joseph.webp"
          theme="dark"
          objectPosition="bottom"
          className="aspect-square rounded-3xl grayscale"
        />
      </Box>

      <Box className="relative">
        <Text
          className="mb-4 mt-16 text-xl uppercase tracking-wide md:mb-6 md:mt-28 md:text-2xl"
          color="gradient"
        >
          What makes Complexus unique
        </Text>
        <Text
          fontWeight="normal"
          className="mb-20 max-w-6xl text-xl leading-snug opacity-80 md:text-2xl"
        >
          Our platform is crafted with a focus on user experience. We understand
          that intuitive design is crucial for productivity, so we've ensured
          that every feature is seamlessly integrated, making it effortless for
          teams to collaborate and achieve their objectives.
        </Text>
        <Box className="mb-16 grid grid-cols-1 gap-10 md:mb-32 md:grid-cols-2 md:gap-x-20 md:gap-y-16">
          {features.map(({ heading, description }, idx) => (
            <Box
              key={heading}
              className="border-t border-gray-200/10 pt-8 md:pt-12"
            >
              <Text className="mb-4 text-6xl opacity-30 md:text-8xl">
                {idx + 1}.
              </Text>
              <Text fontSize="lg" transform="uppercase">
                {heading}
              </Text>
              <Text
                fontSize="lg"
                fontWeight="normal"
                className="mt-4 opacity-80"
              >
                {description}
              </Text>
            </Box>
          ))}
        </Box>
        <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[700px] w-[700px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
      </Box>
    </Container>
  );
};
