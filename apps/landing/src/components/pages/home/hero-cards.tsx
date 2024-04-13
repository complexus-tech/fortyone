"use client";

import { Box } from "ui";
import { EffectCards } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import "swiper/css";
import "swiper/css/effect-cards";
import { Container, Blur } from "@/components/ui";
import { useCursor } from "@/hooks";

export const HeroCards = () => {
  const cursor = useCursor();

  const cards = [
    {
      id: 1,
      title: "Stories",
      image: {
        src: "/stories.png",
        alt: "Stories",
        width: 1854,
        height: 2982,
      },
    },
    {
      id: 2,
      title: "Dashboard",
      image: {
        src: "/dashboard.png",
        alt: "Dashboard",
        width: 1846,
        height: 2980,
      },
    },
    {
      id: 3,
      title: "Story",
      image: {
        src: "/story.png",
        alt: "Story",
        width: 1854,
        height: 2982,
      },
    },
  ];
  return (
    <Container>
      <Box
        className="relative"
        onMouseEnter={() => {
          cursor.setText("←Drag→");
        }}
        onMouseLeave={() => {
          cursor.removeText();
        }}
      >
        <Blur className="absolute -top-28 left-1/2 right-1/2 h-[300px] w-[300px] -translate-x-1/2 bg-warning/10 md:h-[700px] md:w-[700px]" />
        <Blur className="absolute -bottom-28 -left-36 hidden h-[500px] w-[500px] bg-primary/10 md:block" />
        <Blur className="absolute -bottom-6 -right-20 hidden h-[400px] w-[400px] bg-warning/10 md:block" />
        <Swiper
          effect="cards"
          grabCursor
          initialSlide={1}
          modules={[EffectCards]}
        >
          {cards.map((card) => (
            <SwiperSlide
              className="relative rounded-lg border border-dark-50 bg-dark-200 p-0.5 md:rounded-2xl md:p-2"
              key={card.id}
            >
              <div className="pointer-events-none absolute inset-0 bg-[url(/noise.png)] opacity-80" />
              <Image
                alt={card.title}
                className="relative rounded md:rounded-lg"
                height={card.image.height}
                src={card.image.src}
                width={card.image.width}
              />
            </SwiperSlide>
          ))}
        </Swiper>
      </Box>
    </Container>
  );
};
