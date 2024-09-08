import { Text, Box, Button } from "ui";
import { Container, Blur } from "@/components/ui";

export const Hero = () => {
  return (
    <Box>
      <Blur className="absolute -top-[65vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 dark:bg-primary/[0.03]" />
      <Container className="relative max-w-3xl pb-16 pt-36 md:pt-40">
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
          className="my-8 font-satoshi text-5xl font-bold leading-[1.05] md:text-6xl"
          color="gradient"
          fontWeight="medium"
        >
          Convert your images to any format
        </Text>
        <Text
          className="mb-6 max-w-6xl text-xl leading-snug opacity-80 md:mb-20 md:text-2xl"
          fontWeight="normal"
        >
          ImageConveta is an online image converter that allows you to convert
          your images to any format. It is free, fast, and easy to use.
        </Text>
      </Container>
    </Box>
  );
};
