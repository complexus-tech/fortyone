import { Text, Box, Button } from "ui";
import { Container, Blur } from "@/components/ui";

export const Hero = () => {
  return (
    <Box>
      <Blur className="absolute -top-[65vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 dark:bg-primary/[0.03]" />
      <Container className="relative max-w-4xl pb-12 pt-36 md:pt-40">
        <Button
          className="px-3 text-sm md:text-base"
          color="tertiary"
          rounded="full"
          size="sm"
        >
          Our product
        </Button>
        <Text
          as="h1"
          color="gradient"
          fontWeight="medium"
          className="my-8 pb-2 text-5xl md:text-7xl"
        >
          Simplify project management
        </Text>
        <Text
          fontSize="2xl"
          fontWeight="normal"
          className="mb-8 max-w-6xl leading-snug opacity-80"
        >
          We're dedicated to streamlining project management for teams of any
          scale. Our platform is crafted to enable you to efficiently organize
          tasks, collaborate seamlessly, and effortlessly reach your project
          objectives.
        </Text>
      </Container>
    </Box>
  );
};
