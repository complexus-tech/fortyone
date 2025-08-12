"use client";

import { Box } from "ui";
import { EffectCards } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import "swiper/css";
import "swiper/css/effect-cards";
import { Blur, Container } from "@/components/ui";
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
    <Box className="relative">
      <Container className="relative mt-12">
        <Blur className="absolute -top-[12%] left-1/2 right-1/2 h-[100px] -translate-x-1/2 dark:bg-white/15 md:h-[600px] md:w-[800px]" />
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
              delay: 6000,
              disableOnInteraction: false,
            }}
            cardsEffect={{
              slideShadows: false,
            }}
            effect="cards"
            grabCursor
            initialSlide={1}
            loop
            modules={[EffectCards]}
          >
            {cards.map((card) => (
              <SwiperSlide
                className="relative rounded-lg border border-dark/20 bg-white/50 p-0.5 backdrop-blur dark:border-dark-50/70 dark:bg-dark-200/40 md:rounded-3xl md:p-[0.35rem]"
                key={card.id}
              >
                <Box className="relative">
                  <Image
                    alt={card.title}
                    className="relative hidden rounded-[0.4rem] border border-dark-50/70 dark:block md:rounded-[1.2rem]"
                    placeholder="blur"
                    priority
                    src={card.image.src}
                  />
                  <Image
                    alt={card.title}
                    className="relative rounded-[0.4rem] border border-dark/20 dark:hidden md:rounded-[1.2rem]"
                    placeholder="blur"
                    priority
                    src={card.image.srcLight}
                  />
                  <Box className="absolute inset-0 hidden bg-gradient-to-t from-white via-white/5 dark:from-black dark:via-black/50 md:block" />
                </Box>
              </SwiperSlide>
            ))}
          </Swiper>
        </Box>
      </Container>
      <Box className="pointer-events-none absolute inset-0 z-10 bg-gradient-to-t from-white via-white/70 dark:from-black dark:via-black/80" />
    </Box>
  );
};
