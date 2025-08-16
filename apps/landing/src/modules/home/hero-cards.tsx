"use client";

import { Box, Flex } from "ui";
import { EffectCards } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import "swiper/css";
import "swiper/css/effect-cards";
import { useTheme } from "next-themes";
import { ArrowDown2Icon } from "icons";
import { Blur, Container, Dot } from "@/components/ui";
import { useCursor } from "@/hooks";
import listImg from "../../../public/images/product/list.webp";
import objectiveImg from "../../../public/images/product/objective.webp";
import kanbanImg from "../../../public/images/product/kanban.webp";
import listImgLight from "../../../public/images/product/list-light.webp";
import kanbanImgLight from "../../../public/images/product/kanban-light.webp";
import objectiveImgLight from "../../../public/images/product/objective-light.webp";

export const HeroCards = () => {
  const cursor = useCursor();
  const { resolvedTheme } = useTheme();

  const cards = [
    ...(resolvedTheme === "dark"
      ? [
          {
            id: 1,
            title: "List stories",
            image: {
              src: listImg,
              srcLight: listImgLight,
              alt: "List stories",
            },
          },
        ]
      : []),
    {
      id: 2,
      title: "Kanban",
      image: {
        src: kanbanImg,
        srcLight: kanbanImgLight,
        alt: "Kanban",
      },
    },
    ...(resolvedTheme === "dark"
      ? [
          {
            id: 3,
            title: "Objective",
            image: {
              src: objectiveImg,
              srcLight: objectiveImgLight,
              alt: "Objective",
            },
          },
        ]
      : []),
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
                className="relative rounded-lg border border-gray-100 bg-white/50 p-0.5 shadow-xl shadow-gray-100 backdrop-blur dark:border-dark-50/70 dark:bg-dark-200/40 dark:shadow-none md:rounded-2xl md:p-[0.35rem]"
                key={card.id}
              >
                <Flex
                  align="center"
                  className="mb-2 mt-1 px-1.5"
                  justify="between"
                >
                  <Flex className="gap-1.5">
                    <Dot className="size-2.5 text-primary" />
                    <Dot className="size-2.5 text-warning" />
                    <Dot className="size-2.5 text-success" />
                  </Flex>
                  <ArrowDown2Icon className="h-3.5" strokeWidth={2.5} />
                </Flex>
                <Box className="relative">
                  <Image
                    alt={card.title}
                    className="relative hidden rounded-[0.4rem] border border-dark-50/70 dark:block md:rounded-[0.7rem]"
                    placeholder="blur"
                    priority
                    src={card.image.src}
                  />
                  <Image
                    alt={card.title}
                    className="relative rounded-[0.4rem] border border-gray-100 dark:hidden md:rounded-[0.7rem]"
                    placeholder="blur"
                    priority
                    src={card.image.srcLight}
                  />
                  <Box className="absolute inset-0 hidden bg-gradient-to-t from-white via-white/5 dark:block dark:from-black dark:via-black/10" />
                </Box>
              </SwiperSlide>
            ))}
          </Swiper>
        </Box>
      </Container>
      <Box className="pointer-events-none absolute inset-0 z-10 hidden bg-gradient-to-t from-white via-white/70 dark:block dark:from-black dark:via-black/80" />
    </Box>
  );
};
