import { Text, Box, Button } from "ui";
import { Container, Blur } from "@/components/ui";

export const Hero = () => {
  return (
    <Box>
      <Blur className="absolute -top-[65vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 dark:bg-primary/[0.03]" />
      <Container className="relative max-w-4xl pb-16 pt-36 md:pt-40">
        <Button
          className="px-3 text-sm md:text-base"
          color="tertiary"
          rounded="full"
          size="sm"
        >
          Get to know us
        </Button>
        <Text
          as="h1"
          color="gradient"
          fontWeight="medium"
          className="my-8 text-5xl leading-[1.05] md:text-7xl"
        >
          Achieve your goals with Complexus
        </Text>
        <Text
          fontWeight="normal"
          className="mb-6 max-w-6xl text-xl leading-snug opacity-80 md:mb-20 md:text-2xl"
        >
          At Complexus, we believe in simplifying project management for teams
          of all sizes. Our platform is designed to empower you to organize
          tasks, collaborate effectively, and achieve your project goals with
          ease.
        </Text>
      </Container>
    </Box>
  );
};
