"use client";

import type { ComponentPropsWithoutRef, ComponentType } from "react";
import type { MotionProps } from "framer-motion";
import { cn } from "lib";
import { LayoutGroup, motion, useReducedMotion } from "framer-motion";
import { Container } from "@/components/ui";
import styles from "./hero-carousel-indicator.module.css";

// Framer Motion 11 resolves intrinsic element props as `unknown` under the
// landing app's React 19 types. Keep the compatibility cast local until the
// animation dependency is upgraded.
const MotionButton = motion.button as ComponentType<
  ComponentPropsWithoutRef<"button"> & MotionProps
>;
const MotionSpan = motion.span as ComponentType<
  ComponentPropsWithoutRef<"span"> & MotionProps
>;

const CONTROLLER_TRANSITION = {
  bounce: 0,
  duration: 0.3,
  type: "spring",
} as const;
const REDUCED_MOTION_TRANSITION = { duration: 0 } as const;

type HeroCarouselIndicatorProps = {
  activeIndex: number;
  isPaused: boolean;
  onAdvance: () => void;
  onSelect: (index: number) => void;
  slideLabels: readonly string[];
  slideCount: number;
};

export const HeroCarouselIndicator = ({
  activeIndex,
  isPaused,
  onAdvance,
  onSelect,
  slideLabels,
  slideCount,
}: HeroCarouselIndicatorProps) => {
  const shouldReduceMotion = useReducedMotion();
  const controllerTransition = shouldReduceMotion
    ? REDUCED_MOTION_TRANSITION
    : CONTROLLER_TRANSITION;

  return (
    <LayoutGroup id="hero-carousel-controller">
      <Container className="mt-8 md:mt-10">
        <div
          aria-label="Choose a product view"
          className={cn(
            "flex h-2 items-center justify-start gap-1.5",
            styles.controls,
          )}
          data-paused={isPaused ? "true" : "false"}
          role="group"
        >
          {Array.from({ length: slideCount }, (_, index) => {
            const isActive = index === activeIndex;

            return (
              <MotionButton
                aria-current={isActive ? "true" : undefined}
                aria-label={`Show ${slideLabels[index]} view, ${index + 1} of ${slideCount}`}
                className={cn(
                  "focus-visible:outline-info relative h-2 cursor-pointer rounded-full focus-visible:outline-2 focus-visible:outline-offset-2",
                  isActive
                    ? "w-12"
                    : "bg-foreground/8 w-2 shadow-inner shadow-black/5 dark:bg-white/8 dark:shadow-black/20",
                )}
                key={index}
                layout
                onClick={() => {
                  onSelect(index);
                }}
                transition={controllerTransition}
                type="button"
              >
                {isActive ? (
                  <MotionSpan
                    className={styles.activeTrack}
                    layoutId="hero-carousel-active-track"
                    transition={controllerTransition}
                  >
                    <span
                      className={cn(
                        styles.activeIndicator,
                        styles.progressFill,
                      )}
                      onAnimationEnd={onAdvance}
                    />
                  </MotionSpan>
                ) : null}
              </MotionButton>
            );
          })}
        </div>
      </Container>
    </LayoutGroup>
  );
};
