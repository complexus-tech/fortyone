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
    <Container className="relative z-10 mt-10 hidden md:block">
      <Flex className="gap-10" align="center">
        <Text
          color="muted"
          fontSize="sm"
          className="shrink-0 pl-0.5 leading-4.5"
        >
          Trusted by <br />
          teams at
        </Text>
        <Flex align="center" className="gap-12" wrap>
          <Image
            src={miningoLogo}
            alt="Miningo logo"
            className="h-9.5 w-auto grayscale dark:invert"
          />
          <Image
            src={digitankLogo}
            alt="Digitank logo"
            className="h-6 w-auto grayscale dark:invert"
          />

          <Image
            src={artCircles}
            alt="Art Circles logo"
            className="mb-1 h-4.5 w-auto opacity-80 dark:invert"
          />
          <Image
            src={zimboriginalLogo}
            alt="Zimboriginal logo"
            className="h-8 w-auto grayscale dark:invert"
          />
        </Flex>
      </Flex>
    </Container>
  );
};
