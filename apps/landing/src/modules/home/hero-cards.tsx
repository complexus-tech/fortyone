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
        <Blur className="absolute -bottom-28 -left-36 isolate hidden h-[600px] w-[600px] bg-warning/15 dark:md:block" />
        <Blur className="absolute -bottom-6 -right-20 hidden h-[500px] w-[500px] bg-warning/20 dark:md:block" />
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
              className="relative rounded-lg border bg-gray-100/60 p-0.5 backdrop-blur dark:border-gray/60 dark:bg-dark-100/40 md:rounded-xl md:p-1.5"
              key={card.id}
            >
              <Image
                alt={card.title}
                className="relative rounded border dark:border-dark-100 md:rounded-lg"
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
