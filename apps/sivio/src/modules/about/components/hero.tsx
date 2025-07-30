import React from "react";
import { Box, BlurImage, Text, Flex } from "ui";

export const Hero = () => {
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
        justify="center"
      >
        <Text className="text-center text-8xl font-semibold uppercase text-white">
          About Us
        </Text>
      </Flex>
    </Box>
  );
};
