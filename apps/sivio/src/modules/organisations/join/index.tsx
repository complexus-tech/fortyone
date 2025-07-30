import React from "react";
import { Box, BlurImage, Text, Flex } from "ui";
import { Hero } from "./compoments/hero";
import { Partner } from "./compoments/partner";
import { Why } from "./compoments/why";
import { Role } from "./compoments/role";
import { Apply } from "./compoments/apply";

export const JoinPage = () => {
  return (
    <Box>
      <Hero />
      <Partner />
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
          <Text className="max-w-4xl text-center text-5xl font-semibold text-white">
            Raise funds with purpose, transparency, and trust.
          </Text>
        </Flex>
      </Box>
      <Why />
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
          <Text className="max-w-4xl text-center text-5xl font-semibold text-white">
            AfricaGiving provides the platform — you build the relationships.
          </Text>
        </Flex>
      </Box>
      <Role />
      <Apply />
    </Box>
  );
};
