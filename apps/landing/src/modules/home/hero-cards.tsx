"use client";

import { Box } from "ui";
import { EffectCards } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import "swiper/css";
import "swiper/css/effect-cards";
import { Container, Blur } from "@/components/ui";
import { useCursor } from "@/hooks";
import listImg from "../../../public/images/product/list.webp";
import objectiveImg from "../../../public/images/product/objective.webp";
import kanbanImg from "../../../public/images/product/kanban.webp";

export const HeroCards = () => {
  const cursor = useCursor();

  const cards = [
    {
      id: 1,
      title: "List stories",
      image: {
        src: listImg,
        alt: "List stories",
      },
    },
    {
      id: 2,
      title: "Objective",
      image: {
        src: objectiveImg,
        alt: "Objective",
      },
    },
    {
      id: 3,
      title: "Kanban",
      image: {
        src: kanbanImg,
        alt: "Kanban",
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
        <Blur className="absolute -top-28 left-1/2 right-1/2 z-10 h-[300px] w-[300px] -translate-x-1/2 bg-warning/5 md:h-[700px] md:w-[65vw]" />
        <Blur className="absolute -bottom-28 -left-36 hidden h-[500px] w-[500px] bg-warning/10 md:block" />
        <Blur className="absolute -bottom-6 -right-20 hidden h-[400px] w-[400px] bg-warning/15 md:block" />
        <Swiper
          autoplay={{
            delay: 2000,
            disableOnInteraction: false,
          }}
          effect="cards"
          grabCursor
          initialSlide={1}
          modules={[EffectCards]}
        >
          {cards.map((card) => (
            <SwiperSlide
              className="relative rounded-lg border border-gray/60 bg-dark-100/40 p-0.5 backdrop-blur md:rounded-xl md:p-1.5"
              key={card.id}
            >
              <Image
                alt={card.title}
                className="relative rounded border border-dark-100 md:rounded-lg"
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
