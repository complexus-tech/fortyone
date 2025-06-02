"use client";

import { Box } from "ui";
import { EffectCards } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import "swiper/css";
import "swiper/css/effect-cards";
import { Container } from "@/components/ui";
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
      title: "Kanban",
      image: {
        src: kanbanImg,
        alt: "Kanban",
      },
    },
    {
      id: 3,
      title: "Objective",
      image: {
        src: objectiveImg,
        alt: "Objective",
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
              className="relative rounded-lg border border-dark-100 bg-dark-100/40 p-0.5 backdrop-blur md:rounded-2xl md:p-1.5"
              key={card.id}
            >
              <Image
                alt={card.title}
                className="relative rounded border dark:border-dark-100 md:rounded-[0.8rem]"
                placeholder="blur"
                priority
                src={card.image.src}
              />
            </SwiperSlide>
          ))}
        </Swiper>
      </Box>
    </Container>
  );
};
