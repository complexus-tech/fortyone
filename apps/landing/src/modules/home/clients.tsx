"use client";
import React from "react";
import Marquee from "react-fast-marquee";
import { Text } from "ui";
import { useTheme } from "next-themes";
import { Container } from "@/components/ui";

const Brand = ({ logo }: { logo: string }) => {
  return (
    <img
      alt="brand logo"
      className="3xl:h-20 mr-8 block h-8 w-auto grayscale dark:invert md:mr-16 md:h-9 md:justify-self-center"
      key={logo}
      loading="lazy"
      src={logo}
    />
  );
};

export const SampleClients = () => {
  const { resolvedTheme } = useTheme();
  const brands = [
    "/images/brands/miningo.svg",
    "/images/brands/mds.svg",
    "/images/brands/nesbil.png",
    "/images/brands/zimboriginal.png",
    "/images/brands/digitank.png",
    "/images/brands/wastemate.png",
  ];

  return (
    <Container className="relative -top-4 z-10 mb-20 max-w-6xl md:-top-20">
      <Text
        as="h2"
        className="mb-10 text-center font-semibold opacity-80 md:mb-20"
      >
        Trusted by
      </Text>
      <Marquee
        gradient
        gradientColor={resolvedTheme === "light" ? "#ffffff" : "#08090a"}
        pauseOnHover
        speed={30}
      >
        {brands.map((logo) => (
          <Brand key={logo} logo={logo} />
        ))}
      </Marquee>
    </Container>
  );
};
