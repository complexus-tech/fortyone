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
    {
      id: 1,
      title: "List tasks",
      image: {
        src: listImg,
        srcLight: listImgLight,
        alt: "Project management list view - task tracking",
      },
    },
    {
      id: 2,
      title: "Kanban",
      image: {
        src: kanbanImg,
        srcLight: kanbanImgLight,
        alt: "Project management kanban board - task workflow",
      },
    },
    {
      id: 3,
      title: "Objective",
      image: {
        src: objectiveImg,
        srcLight: objectiveImgLight,
        alt: "OKR objective tracking in project management platform",
      },
    },
  ];

  return (
    <Box className="relative">
      <Container className="relative mt-12">
        <Blur className="absolute -top-[12%] right-1/2 left-1/2 h-[100px] -translate-x-1/2 md:h-[600px] md:w-[800px] dark:bg-white/15" />
        <Box
          className="relative"
          onMouseEnter={() => {
            if (resolvedTheme === "dark") {
              cursor.setText("←Drag→");
            }
          }}
          onMouseLeave={() => {
            if (resolvedTheme === "dark") {
              cursor.removeText();
            }
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
                className="border-border bg-background/5 d relative rounded-lg border p-0.5 backdrop-blur md:rounded-2xl md:p-[0.35rem]"
                key={card.id}
              >
                <Flex
                  align="center"
                  className="mt-1 mb-2 px-1.5"
                  justify="between"
                >
                  <Flex className="gap-1.5">
                    <Dot className="text-primary size-2.5" />
                    <Dot className="text-warning size-2.5" />
                    <Dot className="text-success size-2.5" />
                  </Flex>
                  <ArrowDown2Icon className="h-3.5" strokeWidth={2.5} />
                </Flex>
                <Box className="relative">
                  <Image
                    alt={card.title}
                    className="border-border/70 relative hidden rounded-[0.4rem] border md:rounded-[0.7rem] dark:block"
                    placeholder="blur"
                    priority
                    src={card.image.src}
                  />
                  <Image
                    alt={card.title}
                    className="border-border relative rounded-[0.4rem] border md:rounded-[0.7rem] dark:hidden"
                    placeholder="blur"
                    priority
                    src={card.image.srcLight}
                  />
                </Box>
              </SwiperSlide>
            ))}
          </Swiper>
        </Box>
      </Container>
      <Box className="pointer-events-none absolute inset-0 z-10 hidden bg-linear-to-t from-white via-white/70 dark:block dark:from-black dark:via-black/80 dark:via-30%" />
    </Box>
  );
};
