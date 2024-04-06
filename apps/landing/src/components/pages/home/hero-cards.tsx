"use client";

import { Box } from "ui";
import { EffectCards, Autoplay } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import "swiper/css";
import "swiper/css/effect-cards";
import { Container, Blur } from "@/components/ui";

export const HeroCards = () => {
  return (
    <Container>
      <Box className="relative">
        <Blur className="absolute -top-10 left-1/2 right-1/2 h-[500px] w-[500px] -translate-x-1/2 bg-primary/40 dark:bg-primary/20" />
        <Blur className="absolute -bottom-28 -left-36 h-[500px] w-[500px] bg-warning/50 dark:bg-primary/10" />
        <Blur className="absolute -bottom-6 right-20 -z-10 h-[400px] w-[400px] bg-white dark:bg-warning/20" />
        <Swiper effect="cards" grabCursor modules={[EffectCards, Autoplay]}>
          <SwiperSlide className="relative rounded-2xl">
            <Image
              alt="Stories"
              className="animate-gradient rounded-2xl bg-gradient-to-br from-primary via-secondary to-warning/60 p-1 dark:p-0.5"
              height={2982}
              src="/stories.png"
              width={1854}
            />
            <div className="pointer-events-none absolute inset-0 z-10 bg-[url(/noise.png)] opacity-50" />
          </SwiperSlide>
          <SwiperSlide className="relative rounded-2xl">
            <Image
              alt="Stories"
              className="animate-gradient rounded-2xl bg-gradient-to-br from-primary via-secondary to-warning/60 p-1 dark:p-0.5"
              height={2982}
              src="/story.png"
              width={1854}
            />
            <div className="pointer-events-none absolute inset-0 z-10 bg-[url(/noise.png)] opacity-50" />
          </SwiperSlide>
          <SwiperSlide className="relative rounded-2xl">
            <Image
              alt="Dashboard"
              className="animate-gradient rounded-2xl bg-gradient-to-br from-primary via-secondary to-warning/60 p-1 dark:p-0.5"
              height={2980}
              src="/dashboard.png"
              width={1846}
            />
            <div className="pointer-events-none absolute inset-0 z-10 bg-[url(/noise.png)] opacity-50" />
          </SwiperSlide>
        </Swiper>
      </Box>
    </Container>
  );
};
