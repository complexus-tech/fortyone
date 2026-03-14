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
import objectiveImg from "../../../public/images/product/objective.webp";
import kanbanImg from "../../../public/images/product/kanban.webp";
import kanbanImgLight from "../../../public/images/product/kanban-light.webp";
import objectiveImgLight from "../../../public/images/product/objective-light.webp";

export const HeroCards = () => {
  const cursor = useCursor();
  const { resolvedTheme } = useTheme();

  const cards = [
    {
      id: 1,
      title: "Blockers don't hide in kanban. They surface.",
      description:
        "Maya watches ownership and activity. If something starts to stall, it surfaces early instead of waiting for the end-of-sprint retro.",
      image: {
        src: kanbanImg,
        srcLight: kanbanImgLight,
        alt: "Project management kanban board - task workflow",
      },
    },
    {
      id: 2,
      title: "Done tasks move the quarter, not just the board.",
      description:
        "As tasks close, objective progress updates in real time. Leaders see the quarter move without asking for updates.",
      image: {
        src: objectiveImg,
        srcLight: objectiveImgLight,
        alt: "OKR objective tracking in project management platform",
      },
    },
  ];

  return (
    <Box>
      <Box className="relative">
        <Container className="relative mt-12 overflow-visible">
          <Blur className="bg-warning/12 absolute -top-[14%] right-5/12 h-[100px] -translate-x-1/2 md:h-[800px] md:w-[800px]" />
          <Blur className="bg-danger/12 absolute -top-[14%] -right-8/12 h-[100px] -translate-x-1/2 md:h-[800px] md:w-[800px]" />
          <Box
            className="relative -mr-5 w-[calc(100%+1.25rem)] overflow-hidden md:mr-0 md:w-auto md:overflow-visible"
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
                  className="border-border bg-background/50 dark:bg-background/5 relative rounded-l-lg rounded-r-none border p-0.5 backdrop-blur md:rounded-xl md:p-[0.35rem]"
                  key={card.id}
                >
                  <Flex
                    align="center"
                    className="mt-1 mb-2 justify-start px-1.5 md:justify-between"
                  >
                    <Flex className="gap-1.5">
                      <Dot className="text-primary size-2.5" />
                      <Dot className="text-warning size-2.5" />
                      <Dot className="text-success size-2.5" />
                    </Flex>
                    <ArrowDown2Icon
                      className="hidden h-3.5 md:block"
                      strokeWidth={2.5}
                    />
                  </Flex>
                  <Box className="relative overflow-hidden rounded-l-md rounded-r-none md:rounded-lg">
                    <Image
                      alt={card.title}
                      className="border-border/70 relative hidden h-88 w-auto max-w-none rounded-l-md rounded-r-none border md:h-auto md:w-full md:max-w-full md:rounded-lg dark:block"
                      placeholder="blur"
                      priority={card.id === 2}
                      src={card.image.src}
                    />
                    <Image
                      alt={card.title}
                      className="border-border relative h-88 w-auto max-w-none rounded-l-md rounded-r-none border md:h-auto md:w-full md:max-w-full md:rounded-lg dark:hidden"
                      placeholder="blur"
                      priority={card.id === 2}
                      src={card.image.srcLight}
                    />
                  </Box>
                </SwiperSlide>
              ))}
            </Swiper>
          </Box>
        </Container>
        <Box className="pointer-events-none absolute right-0 -bottom-1 left-0 z-10 hidden h-120 bg-linear-to-t from-white md:block dark:h-140 dark:from-black dark:via-black/80 dark:via-30%" />
      </Box>
    </Box>
  );
};
