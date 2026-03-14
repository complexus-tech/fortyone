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
    <Container className="relative z-10 mt-16 hidden md:block">
      <Flex align="center" className="justify-center gap-12" wrap>
        <Image
          src={miningoLogo}
          alt="Miningo logo"
          className="h-11 w-auto grayscale dark:invert"
        />
        <Image
          src={digitankLogo}
          alt="Digitank logo"
          className="h-7 w-auto grayscale dark:invert"
        />

        <Image
          src={artCircles}
          alt="Art Circles logo"
          className="mb-1 h-5.5 w-auto opacity-80 dark:invert"
        />
        <Image
          src={zimboriginalLogo}
          alt="Zimboriginal logo"
          className="h-9 w-auto grayscale dark:invert"
        />
      </Flex>
    </Container>
  );
};
