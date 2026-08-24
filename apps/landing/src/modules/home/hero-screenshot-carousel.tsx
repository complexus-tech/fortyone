"use client";

import type { FocusEvent } from "react";
import { useEffect, useState } from "react";
import { cn } from "lib";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import myWorkBoardDark from "../../../public/images/product/my-work-board-dark.webp";
import myWorkBoardLight from "../../../public/images/product/my-work-board-light.webp";
import myWorkListDark from "../../../public/images/product/my-work-list-dark.webp";
import myWorkListLight from "../../../public/images/product/my-work-list-light.webp";
import { HeroCarouselIndicator } from "./hero-carousel-indicator";
import { ProductScreenshot } from "./product-screenshot";
import styles from "./hero-screenshot-carousel.module.css";

const HERO_SLIDES = [
  {
    alt: "FortyOne My Work board showing tasks grouped by Backlog, To Do, In Progress, and Done",
    darkImage: myWorkBoardDark,
    id: "board",
    label: "Board",
    lightImage: myWorkBoardLight,
    url: "https://fortyone.app/my-work",
  },
  {
    alt: "FortyOne My Work list showing tasks with status, priority, owners, and delivery details",
    darkImage: myWorkListDark,
    id: "list",
    label: "List",
    lightImage: myWorkListLight,
    url: "https://fortyone.app/my-work",
  },
  {
    alt: "FortyOne Maya AI delivery brief showing objective counts, completion metrics, and delivery trends",
    darkImage: mayaDeliveryBriefDark,
    id: "maya",
    label: "Maya delivery brief",
    lightImage: mayaDeliveryBriefLight,
    url: "https://fortyone.app/maya",
  },
] as const;

export const HeroScreenshotCarousel = () => {
  const [activeIndex, setActiveIndex] = useState(0);
  const [previousIndex, setPreviousIndex] = useState<number | null>(null);
  const [isInteractionPaused, setIsInteractionPaused] = useState(false);
  const [isPageHidden, setIsPageHidden] = useState(false);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  useEffect(() => {
    const reducedMotionQuery = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    );
    const handleReducedMotionChange = () => {
      setPrefersReducedMotion(reducedMotionQuery.matches);
    };
    const handleVisibilityChange = () => {
      setIsPageHidden(document.visibilityState === "hidden");
    };

    handleReducedMotionChange();
    handleVisibilityChange();
    reducedMotionQuery.addEventListener("change", handleReducedMotionChange);
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      reducedMotionQuery.removeEventListener(
        "change",
        handleReducedMotionChange,
      );
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, []);

  const handleAdvance = () => {
    if (prefersReducedMotion || isInteractionPaused || isPageHidden) return;

    setPreviousIndex(activeIndex);
    setActiveIndex((activeIndex + 1) % HERO_SLIDES.length);
  };

  const handleSelect = (nextIndex: number) => {
    if (nextIndex === activeIndex) return;

    setPreviousIndex(activeIndex);
    setActiveIndex(nextIndex);
  };

  const handleBlur = (event: FocusEvent<HTMLElement>) => {
    if (!event.currentTarget.contains(event.relatedTarget)) {
      setIsInteractionPaused(false);
    }
  };

  return (
    <section
      aria-label="FortyOne product views"
      aria-roledescription="carousel"
      onBlur={handleBlur}
      onFocus={() => {
        setIsInteractionPaused(true);
      }}
      onPointerEnter={() => {
        setIsInteractionPaused(true);
      }}
      onPointerLeave={() => {
        setIsInteractionPaused(false);
      }}
    >
      <HeroCarouselIndicator
        activeIndex={activeIndex}
        isPaused={isInteractionPaused || isPageHidden || prefersReducedMotion}
        onAdvance={handleAdvance}
        onSelect={handleSelect}
        slideCount={HERO_SLIDES.length}
        slideLabels={HERO_SLIDES.map((slide) => slide.label)}
      />
      <div
        aria-atomic="true"
        aria-live="polite"
        className={styles.viewport}
        data-landing-reveal
      >
        {HERO_SLIDES.map((slide, index) => {
          const isActive = index === activeIndex;
          let slideStateClassName = styles.queuedSlide;

          if (isActive) {
            slideStateClassName = styles.activeSlide;
          } else if (index === previousIndex) {
            slideStateClassName = styles.exitingSlide;
          }

          return (
            <div
              aria-hidden={isActive ? undefined : "true"}
              className={cn(styles.slide, slideStateClassName)}
              data-active={isActive ? "true" : "false"}
              data-carousel-slide={slide.id}
              key={slide.id}
            >
              <ProductScreenshot
                alt={slide.alt}
                containerClassName="mt-4 md:mt-5"
                cropBrowserOnMobile
                darkImage={slide.darkImage}
                lightImage={slide.lightImage}
                priority={index === 0}
                reveal={false}
                url={slide.url}
              />
            </div>
          );
        })}
      </div>
    </section>
  );
};
