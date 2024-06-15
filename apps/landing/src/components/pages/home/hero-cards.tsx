"use client";

import { Box } from "ui";
import { EffectCards } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import "swiper/css";
import "swiper/css/effect-cards";
import { Container, Blur } from "@/components/ui";
import { useCursor } from "@/hooks";
import storiesImg from "../../../../public/stories.webp";
import dashboardImg from "../../../../public/dashboard.webp";
import storyImg from "../../../../public/story.png";

export const HeroCards = () => {
  const cursor = useCursor();

  const cards = [
    {
      id: 1,
      title: "Dashboard",
      image: {
        src: dashboardImg,
        alt: "Dashboard",
      },
    },
    {
      id: 2,
      title: "Stories",
      image: {
        src: storiesImg,
        alt: "Stories",
      },
    },
    {
      id: 3,
      title: "Story",
      image: {
        src: storyImg,
        alt: "Story",
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
        <Blur className="absolute -top-28 left-1/2 right-1/2 h-[300px] w-[300px] -translate-x-1/2 bg-warning/15 md:h-[700px] md:w-[800px]" />
        <Blur className="absolute -bottom-28 -left-36 hidden h-[500px] w-[500px] bg-warning/10 md:block" />
        <Blur className="absolute -bottom-6 -right-20 hidden h-[400px] w-[400px] bg-warning/10 md:block" />
        <Swiper
          effect="cards"
          grabCursor
          initialSlide={1}
          modules={[EffectCards]}
        >
          {cards.map((card) => (
            <SwiperSlide
              className="relative rounded-lg border border-gray/60 bg-dark-100/40 p-0.5 backdrop-blur md:rounded-2xl md:p-2"
              key={card.id}
            >
              <Image
                alt={card.title}
                className="relative rounded border-[0.5px] border-dark-100 md:rounded-lg"
                placeholder="blur"
                src={card.image.src}
              />
            </SwiperSlide>
          ))}
        </Swiper>
      </Box>
    </Container>
  );
};
