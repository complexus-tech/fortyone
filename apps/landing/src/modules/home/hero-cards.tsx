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
import listImgLight from "../../../public/images/product/list-light.webp";
import kanbanImgLight from "../../../public/images/product/kanban-light.webp";
import objectiveImgLight from "../../../public/images/product/objective-light.webp";

export const HeroCards = () => {
  const cursor = useCursor();

  const cards = [
    {
      id: 1,
      title: "List stories",
      image: {
        src: listImg,
        srcLight: listImgLight,
        alt: "List stories",
      },
    },
    {
      id: 2,
      title: "Kanban",
      image: {
        src: kanbanImg,
        srcLight: kanbanImgLight,
        alt: "Kanban",
      },
    },
    {
      id: 3,
      title: "Objective",
      image: {
        src: objectiveImg,
        srcLight: objectiveImgLight,
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
              className="relative rounded-[0.6rem] border border-gray-200/80 bg-white/50 p-0.5 backdrop-blur dark:border-dark-100 dark:bg-dark-100/40 md:rounded-3xl md:p-1.5"
              key={card.id}
            >
              <Image
                alt={card.title}
                className="relative hidden rounded border border-gray-200/80 dark:block dark:border-dark-100 md:rounded-[1.2rem]"
                placeholder="blur"
                priority
                src={card.image.src}
              />
              <Image
                alt={card.title}
                className="relative rounded border border-gray-200/80 dark:hidden dark:border-dark-100 md:rounded-[1.2rem]"
                placeholder="blur"
                priority
                src={card.image.srcLight}
              />
            </SwiperSlide>
          ))}
        </Swiper>
      </Box>
    </Container>
  );
};
