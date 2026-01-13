"use client";
import React from "react";
import { Container } from "@/components/ui";
import { Flex, Text } from "ui";
import digitankLogo from "../../../public/images/brands/digitank.png";
import miningoLogo from "../../../public/images/brands/miningo.svg";
import zimboriginalLogo from "../../../public/images/brands/zimboriginal.png";
import artCircles from "../../../public/images/brands/artcircles.png";
import Image from "next/image";

export const SampleClients = () => {
  return (
    <Container className="relative z-10 mt-12 hidden md:block">
      <Flex className="gap-10" align="center">
        <Text color="muted" fontSize="sm" className="shrink-0">
          Trusted by leading <br />
          product teams at
        </Text>
        <Flex align="center" className="gap-12" wrap>
          <Image
            src={miningoLogo}
            alt="Miningo logo"
            className="h-11 w-auto grayscale dark:invert"
          />
          <Image
            src={digitankLogo}
            alt="Digitank logo"
            className="h-8 w-auto grayscale dark:invert"
          />

          <Image
            src={artCircles}
            alt="Art Circles logo"
            className="mb-1.5 h-6 w-auto opacity-80 dark:invert"
          />
          <Image
            src={zimboriginalLogo}
            alt="Zimboriginal logo"
            className="h-10 w-auto grayscale dark:invert"
          />
        </Flex>
      </Flex>
    </Container>
  );
};
