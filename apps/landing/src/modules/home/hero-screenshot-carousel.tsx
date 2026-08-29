"use client";

import type { FocusEvent } from "react";
import { useEffect, useState } from "react";
import { cn } from "lib";
import feedbackPortalDark from "../../../public/images/product/feedback-portal-dark.webp";
import feedbackPortalLight from "../../../public/images/product/feedback-portal-light.webp";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import strategyMapDark from "../../../public/images/product/strategy-map-dark.webp";
import strategyMapLight from "../../../public/images/product/strategy-map-light.webp";
import { HeroCarouselIndicator } from "./hero-carousel-indicator";
import { ProductScreenshot } from "./product-screenshot";
import styles from "./hero-screenshot-carousel.module.css";

const HERO_SLIDES = [
  {
    alt: "FortyOne Strategy Map connecting company goals to strategic pillars and objectives",
    darkImage: strategyMapDark,
    id: "strategy",
    label: "Strategy map",
    lightImage: strategyMapLight,
    url: "https://fortyone.app/strategy-map",
  },
  {
    alt: "FortyOne feedback portal showing customer requests, votes, and product feedback",
    darkImage: feedbackPortalDark,
    id: "feedback",
    label: "Customer feedback",
    lightImage: feedbackPortalLight,
    url: "https://fortyone.app/feedback",
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
