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
          Contact us
        </Button>
        <Text
          as="h1"
          color="gradient"
          fontWeight="medium"
          className="my-8 pb-2 text-5xl leading-none md:text-7xl"
        >
          What assistance do you require?
        </Text>
        <Text
          fontWeight="normal"
          className="max-w-6xl text-xl leading-snug opacity-80 md:mb-16 md:text-2xl"
        >
          Reach out to our sales and support teams for demonstrations,
          assistance with onboarding, or any questions regarding complexus.
        </Text>
      </Container>
    </Box>
  );
};
