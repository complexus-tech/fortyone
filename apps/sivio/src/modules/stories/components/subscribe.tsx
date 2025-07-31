import React from "react";
import { Box, BlurImage, Text, Flex, Button, Input } from "ui";

export const Subscribe = () => {
  return (
    <Box className="relative">
      <BlurImage
        className="aspect-[16/6]"
        quality={100}
        src="/images/about/4.jpg"
      />
      <Flex
        align="center"
        className="absolute inset-0 bg-black/50"
        direction="column"
        justify="center"
      >
        <Text className="mb-4 text-center text-6xl font-semibold text-white">
          Subscribe
        </Text>
        <Text className="text-center text-xl text-white">
          Stay Up to date with our impact stories
        </Text>
        <form className="mt-6 w-full max-w-2xl space-y-4">
          <Input
            className="w-full bg-white"
            placeholder="John Doe"
            rounded="none"
            size="lg"
          />
          <Input
            className="w-full bg-white"
            placeholder="john@doe.com"
            rounded="none"
            size="lg"
          />
          <Button
            className="ml-auto"
            color="secondary"
            rounded="none"
            size="lg"
          >
            Subscribe
          </Button>
        </form>
      </Flex>
    </Box>
  );
};
