import React from "react";
import { Box, BlurImage, Text, Flex } from "ui";

export const Hero = () => {
  return (
    <Box className="relative">
      <BlurImage
        className="aspect-[16/7]"
        quality={100}
        src="/images/home/hero.webp"
      />
      <Flex
        align="center"
        className="absolute inset-0 bg-black/40"
        justify="center"
      >
        <Text className="max-w-6xl text-center text-8xl font-semibold text-white">
          Join the AfricaGiving platform
        </Text>
      </Flex>
    </Box>
  );
};
