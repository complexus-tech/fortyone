import React from "react";
import { Box, BlurImage, Text, Flex } from "ui";

export const Compliance = () => {
  return (
    <Box className="relative">
      <BlurImage
        className="aspect-[16/5]"
        imageClassName=" object-top"
        quality={100}
        src="/images/about/1.jpg"
      />
      <Flex
        align="center"
        className="absolute inset-0 bg-black/60"
        justify="center"
      >
        <Text className="max-w-2xl text-center text-5xl font-semibold text-white">
          Understanding Our Compliance Score
        </Text>
      </Flex>
    </Box>
  );
};
