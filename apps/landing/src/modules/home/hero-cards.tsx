"use client";

import { Box, Flex, Text } from "ui";
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
      title: "Rough idea in. Structured task out.",
      description:
        "Type what needs doing in plain language. Maya turns it into a structured task with context, ownership, and a goal attached.",
      image: {
        src: listImg,
        srcLight: listImgLight,
        alt: "Project management list view - task tracking",
      },
    },
    {
      id: 2,
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
      id: 3,
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
    <Box className="relative">
      <Container className="relative mt-12">
        <Blur className="absolute -top-[12%] right-1/2 left-1/2 h-[100px] -translate-x-1/2 md:h-[600px] md:w-[800px] dark:bg-white/10" />
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
                    className="border-border/70 relative hidden rounded-md border md:rounded-xl dark:block"
                    placeholder="blur"
                    priority={card.id === 2}
                    src={card.image.src}
                  />
                  <Image
                    alt={card.title}
                    className="border-border relative rounded-md border md:rounded-xl dark:hidden"
                    placeholder="blur"
                    priority={card.id === 2}
                    src={card.image.srcLight}
                  />
                </Box>
              </SwiperSlide>
            ))}
          </Swiper>
        </Box>
        <Box className="mt-8 grid grid-cols-1 gap-4 md:mt-10 md:grid-cols-3">
          {cards.map((card) => (
            <Box
              className="border-border bg-surface/70 rounded-2xl border px-5 py-5 text-left"
              key={card.id}
            >
              <Text as="h3" className="mb-2 text-lg font-semibold md:text-xl">
                {card.title}
              </Text>
              <Text className="text-sm leading-relaxed opacity-70 md:text-base">
                {card.description}
              </Text>
            </Box>
          ))}
        </Box>
      </Container>
      <Box className="pointer-events-none absolute inset-0 z-10 hidden bg-linear-to-t from-white via-white/70 dark:block dark:from-black dark:via-black/80 dark:via-30%" />
    </Box>
  );
};
