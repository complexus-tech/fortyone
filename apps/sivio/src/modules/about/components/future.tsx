import React from "react";
import { Box, BlurImage, Text, Flex } from "ui";

export const Future = () => {
  return (
    <Box className="relative">
      <BlurImage
        className="aspect-[16/4]"
        imageClassName=" object-top"
        quality={100}
        src="/images/about/3.jpg"
      />
      <Flex
        align="center"
        className="absolute inset-0 bg-black/50"
        justify="center"
      >
        <Text className="max-w-5xl text-center text-4xl font-semibold italic text-white">
          &ldquo;The future of Africa will be shaped by those who invest in its
          people today.&rdquo;
        </Text>
      </Flex>
    </Box>
  );
};
