"use client";

import type { StaticImageData } from "next/image";
import type { ComponentPropsWithoutRef, ComponentType } from "react";
import type { MotionProps } from "framer-motion";
import { useState } from "react";
import Image from "next/image";
import { MotionConfig, motion } from "framer-motion";
import { PlusIcon } from "icons";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import finLogo from "../../../public/images/brands/fin.png";
import miningoLogo from "../../../public/images/brands/miningo.svg";
import zimboriginalLogo from "../../../public/images/brands/zimboriginal.png";
import finImage from "../../../public/images/testimonials/fin-city.webp";
import miningoImage from "../../../public/images/testimonials/miningo-geospatial.webp";
import zimboriginalImage from "../../../public/images/testimonials/zimboriginal-studio.webp";

type CustomerStory = {
  author: string;
  company: string;
  id: string;
  image: StaticImageData;
  imageClassName?: string;
  logo: StaticImageData | string;
  logoClassName: string;
  quote: string;
  role: string;
};

const STORY_TRANSITION = {
  damping: 60,
  mass: 1,
  stiffness: 500,
  type: "spring",
} as const;

// Framer Motion 11 resolves intrinsic element props as `unknown` under the
// landing app's React 19 types. Keep the compatibility casts local until the
// animation dependency is upgraded.
const MotionButton = motion.button as ComponentType<
  ComponentPropsWithoutRef<"button"> & MotionProps
>;
const MotionDiv = motion.div as ComponentType<
  ComponentPropsWithoutRef<"div"> & MotionProps
>;
const MotionFooter = motion.footer as ComponentType<
  ComponentPropsWithoutRef<"footer"> & MotionProps
>;

const customerStories: readonly CustomerStory[] = [
  {
    author: "Dominic Chingoma",
    company: "Fin",
    id: "fin",
    image: finImage,
    imageClassName: "object-[52%_center]",
    logo: finLogo,
    logoClassName: "h-9 w-auto max-w-24",
    quote:
      "Planning feels much lighter now. I can see what the team is working on, where we’re stuck, and what needs my attention.",
    role: "Head of Engineering & CTO",
  },
  {
    author: "Tshaxedue Gondo",
    company: "Miningo Technologies",
    id: "miningo",
    image: miningoImage,
    imageClassName: "object-[54%_center]",
    logo: miningoLogo,
    logoClassName: "h-11 w-auto max-w-32",
    quote:
      "Our feedback and priorities finally live in one place. I can see what customers need and what the team should focus on next.",
    role: "Founder & Spatial Data Scientist",
  },
  {
    author: "Shungu C. Chidovi",
    company: "Zimboriginal",
    id: "zimboriginal",
    image: zimboriginalImage,
    imageClassName: "object-center",
    logo: zimboriginalLogo,
    logoClassName: "h-7 w-auto max-w-40",
    quote:
      "I can capture an idea while it’s still rough, shape it clearly, and keep it moving without losing anything important.",
    role: "Founder, Writer & Teacher",
  },
];

type CustomerStoryCardProps = {
  active?: boolean;
  mobile?: boolean;
  onSelect?: () => void;
  story: CustomerStory;
};

const CustomerStoryCard = ({
  active = false,
  mobile = false,
  onSelect,
  story,
}: CustomerStoryCardProps) => {
  const showStory = active || mobile;
  let contentPositionClassName =
    "bottom-[-60px] left-[36px] h-[360px] w-[350px] gap-[230px] overflow-hidden";

  if (mobile) {
    contentPositionClassName = "inset-[28px] w-auto justify-between";
  } else if (showStory) {
    contentPositionClassName =
      "top-[36px] bottom-[36px] left-[36px] w-[350px] justify-between overflow-visible";
  }

  return (
    <article
      aria-describedby={
        showStory ? `customer-story-${story.id}-quote` : undefined
      }
      aria-label={`${story.company} customer story`}
      className={cn(
        "bg-surface-elevated focus-visible:outline-primary relative isolate w-full overflow-hidden rounded-lg text-white shadow-lg shadow-black/10 focus-visible:outline-2 focus-visible:outline-offset-4",
        mobile ? "h-[416px]" : "h-[540px]",
      )}
      onPointerEnter={
        mobile
          ? undefined
          : (event) => {
              if (event.pointerType === "mouse") {
                onSelect?.();
              }
            }
      }
      tabIndex={mobile ? undefined : -1}
    >
      <Box className="absolute inset-0 overflow-hidden">
        <Box
          className={cn(
            "absolute overflow-hidden",
            mobile
              ? "inset-0"
              : "top-1/2 left-1/2 -mt-[394px] -ml-[365px] h-[788px] w-[730px]",
          )}
        >
          <Image
            alt=""
            aria-hidden="true"
            className={cn("object-cover", story.imageClassName)}
            fill
            sizes={
              mobile
                ? "(max-width: 809px) 19.375rem, (max-width: 1199px) 41.25rem, 1px"
                : "(min-width: 1200px) 45.625rem, 1px"
            }
            src={story.image}
          />
        </Box>
      </Box>

      <Box aria-hidden="true" className="absolute inset-0 bg-black/10" />
      <Box
        aria-hidden="true"
        className={cn(
          "absolute inset-0",
          mobile
            ? "bg-[linear-gradient(180deg,rgba(0,0,0,0)_0%,rgba(0,0,0,0.7)_100%)]"
            : "bg-[linear-gradient(180deg,rgba(0,0,0,0)_0%,rgba(0,0,0,0.5)_100%)]",
        )}
      />
      {!mobile && (
        <MotionDiv
          animate={{ opacity: showStory ? 1 : 0 }}
          aria-hidden="true"
          className="absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0)_0%,rgba(0,0,0,0.6)_100%)]"
          initial={false}
          transition={STORY_TRANSITION}
        />
      )}

      <MotionDiv
        className={cn(
          "absolute z-10",
          mobile ? "top-[28px] left-[28px]" : "top-[38px] left-[36px]",
        )}
      >
        <Image
          alt=""
          aria-hidden="true"
          className={cn(
            "object-contain object-left brightness-0 invert",
            story.logoClassName,
          )}
          src={story.logo}
        />
      </MotionDiv>

      <MotionDiv
        className={cn("absolute z-10 flex flex-col", contentPositionClassName)}
      >
        <MotionDiv
          animate={{
            opacity: showStory ? 1 : 0,
          }}
          aria-hidden={!showStory}
          className={cn(
            "w-full origin-top overflow-hidden",
            showStory && "order-1",
          )}
          initial={false}
          layout={mobile ? false : "position"}
          transition={STORY_TRANSITION}
        >
          <blockquote
            className="m-0 w-full"
            id={`customer-story-${story.id}-quote`}
          >
            <Text className="font-serif text-[24px] leading-[1.2] tracking-[-0.02em] text-pretty">
              “{story.quote}”
            </Text>
          </blockquote>
        </MotionDiv>

        <MotionFooter
          animate={{
            opacity: showStory ? 1 : 0,
          }}
          aria-hidden={!showStory}
          className={cn(
            "flex min-w-0 flex-col gap-1 border-l border-white/35 pl-[18px]",
            showStory && "order-2",
          )}
          initial={false}
          layout={mobile ? false : "position"}
          transition={STORY_TRANSITION}
        >
          <Text as="cite" className="text-sm font-normal text-white not-italic">
            {story.author}
          </Text>
          <Text className="text-[11px] tracking-[0.06em] text-white/68 uppercase">
            {story.role} · {story.company}
          </Text>
        </MotionFooter>

        <MotionDiv
          aria-hidden="true"
          className="h-[10px] shrink-0"
          layout={mobile ? false : "position"}
          transition={STORY_TRANSITION}
        />
      </MotionDiv>

      {!mobile && (
        <MotionButton
          animate={{
            opacity: showStory ? 0 : 1,
          }}
          aria-controls={`customer-story-${story.id}-quote`}
          aria-expanded={showStory}
          aria-hidden={showStory}
          aria-label={`Read ${story.company}'s story`}
          className="focus-visible:outline-primary absolute bottom-[36px] left-[36px] z-20 grid size-[36px] cursor-pointer place-items-center rounded-md bg-black/25 text-white shadow-sm shadow-black/10 backdrop-blur-[5px] hover:bg-black/35 focus-visible:outline-2 focus-visible:outline-offset-2 dark:bg-white/18 dark:hover:bg-white/28"
          inert={showStory}
          initial={false}
          onClick={(event) => {
            const card = event.currentTarget.closest("article");
            const wasKeyboardActivation = event.detail === 0;

            onSelect?.();

            if (wasKeyboardActivation) {
              requestAnimationFrame(() => {
                if (card instanceof HTMLElement) {
                  card.focus({ preventScroll: true });
                }
              });
            }
          }}
          style={{ pointerEvents: showStory ? "none" : "auto" }}
          tabIndex={showStory ? -1 : 0}
          transition={STORY_TRANSITION}
          type="button"
        >
          <PlusIcon
            aria-hidden="true"
            className="size-[24px] text-white"
            strokeWidth={1.5}
          />
        </MotionButton>
      )}
    </article>
  );
};

export const CustomerStories = () => {
  const [activeStoryId, setActiveStoryId] = useState("fin");

  return (
    <Container
      aria-labelledby="customer-stories-title"
      as="section"
      className="scroll-mt-24 pb-16 md:pb-28"
      id="customer-stories"
    >
      <Box
        aria-hidden="true"
        className="border-border/40 mx-auto w-4/5 border-t"
      />
      <Box
        className="mt-16 mb-10 max-w-3xl text-left md:mt-28 md:mb-12"
        data-landing-reveal
      >
        <Text
          as="h2"
          className="text-3xl md:text-5xl"
          id="customer-stories-title"
        >
          What our customers
          <span className="block">say.</span>
        </Text>
        <Text className="text-text-description mt-6 max-w-lg text-base leading-relaxed text-pretty">
          We love building FortyOne with our customers. Here’s some of the love
          they’ve shared with us.
        </Text>
      </Box>

      <Box className="grid gap-[48px] min-[1200px]:hidden" data-landing-reveal>
        {customerStories.map((story) => (
          <Box
            className="mx-auto w-full max-w-[310px] min-[810px]:max-w-[660px]"
            key={story.id}
          >
            <CustomerStoryCard mobile story={story} />
          </Box>
        ))}
      </Box>

      <MotionConfig reducedMotion="user" transition={STORY_TRANSITION}>
        <MotionDiv
          className="hidden items-stretch gap-[10px] min-[1200px]:flex"
          data-landing-reveal
          onMouseLeave={() => {
            setActiveStoryId("fin");
          }}
          style={{ transitionDelay: "70ms" }}
        >
          {customerStories.map((story) => {
            const isActive = story.id === activeStoryId;

            return (
              <MotionDiv
                animate={{ flexGrow: isActive ? 2 : 1 }}
                className="min-w-0"
                initial={false}
                key={story.id}
                style={{ flexBasis: 0, flexShrink: 0 }}
                transition={STORY_TRANSITION}
              >
                <CustomerStoryCard
                  active={isActive}
                  onSelect={() => {
                    setActiveStoryId(story.id);
                  }}
                  story={story}
                />
              </MotionDiv>
            );
          })}
        </MotionDiv>
      </MotionConfig>
    </Container>
  );
};
