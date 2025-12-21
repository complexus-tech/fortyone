"use client";
import React from "react";
import { Container } from "@/components/ui";
import { Flex, Text } from "ui";
import digitankLogo from "../../../public/images/brands/digitank.png";
import miningoLogo from "../../../public/images/brands/miningo.svg";
import zimboriginalLogo from "../../../public/images/brands/zimboriginal.png";
import mdsLogo from "../../../public/images/brands/wastemate.png";
import Image from "next/image";

export const SampleClients = () => {
  const brands = [
    "/images/brands/digitank.png",
    "/images/brands/miningo.svg",
    "/images/brands/mds.svg",
    "/images/brands/nesbil.png",
    "/images/brands/zimboriginal.png",
  ];

  return (
    <Container className="relative z-10 mt-12 hidden md:block">
      <Flex className="gap-10" align="center">
        <Text color="muted" fontSize="sm" className="shrink-0">
          Trusted by leading <br />
          product teams at
        </Text>
        <Flex align="center" className="gap-12" wrap>
          <Image
            src={digitankLogo}
            alt="Digitank logo"
            className="h-8 w-auto grayscale dark:invert"
          />
          <Image
            src={miningoLogo}
            alt="Miningo logo"
            className="h-10 w-auto grayscale dark:invert"
          />
          <Image
            src={zimboriginalLogo}
            alt="Zimboriginal logo"
            className="h-10 w-auto grayscale dark:invert"
          />
          <Image
            src={mdsLogo}
            alt="MDS logo"
            className="h-10 w-auto grayscale dark:invert"
          />
        </Flex>
      </Flex>
    </Container>
  );
};
