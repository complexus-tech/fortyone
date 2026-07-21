"use client";

import type { FocusEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { Box, Text } from "ui";
import { cn } from "lib";
import { Container } from "@/components/ui";

const AUTO_ADVANCE_MS = 8_000;

const testimonialButtonClassName =
  "border-border text-text-muted hover:border-border-strong hover:bg-state-hover hover:text-foreground focus-visible:outline-foreground min-h-10 cursor-pointer rounded-none border bg-transparent font-normal transition-[background-color,border-color,color,transform] duration-150 ease-out focus-visible:outline-2 focus-visible:outline-offset-[3px] active:scale-[0.96]";

const testimonials = [
  {
    quote:
      "Customer feedback now comes into one place, where we can prioritise requests and turn the right ideas into planned work. Our public roadmap also shows customers what the team is working on and what is coming next.",
    author: "Tshaxedue Gondo",
    role: "Founder & Spatial Data Scientist",
    company: "Miningo Technologies",
  },
  {
    quote:
      "We plan product work in one shared view, with priorities, owners, and delivery risks clear to the whole team. GitHub keeps engineering updates connected to each task, so we spend less time rebuilding context and more time delivering.",
    author: "Dominic Chingoma",
    role: "Head of Engineering & CTO",
    company: "Fin",
  },
];

const TestimonialArrow = ({ direction }: { direction: "left" | "right" }) => {
  const isLeft = direction === "left";

  return (
    <svg
      aria-hidden="true"
      className="h-[18px] w-[18px] fill-none stroke-current [stroke-width:1.3] [stroke-linecap:round] [stroke-linejoin:round]"
      viewBox="0 0 18 18"
    >
      <path
        d={
          isLeft
            ? "M10.75 4.75 6.5 9l4.25 4.25"
            : "M7.25 4.75 11.5 9l-4.25 4.25"
        }
      />
      <path d={isLeft ? "M6.75 9h5.75" : "M11.25 9H5.5"} />
    </svg>
  );
};

export const Testimonials = () => {
  const [activeIndex, setActiveIndex] = useState(0);
  const [autoplayEnabled, setAutoplayEnabled] = useState(true);
  const [isInteracting, setIsInteracting] = useState(false);
  const [isDocumentVisible, setIsDocumentVisible] = useState(true);
  const [shouldReduceMotion, setShouldReduceMotion] = useState(false);
  const slideRef = useRef<HTMLQuoteElement>(null);
  const hasMounted = useRef(false);

  const move = (direction: -1 | 1) => {
    setActiveIndex((currentIndex) => {
      return (
        (currentIndex + direction + testimonials.length) % testimonials.length
      );
    });
  };

  useEffect(() => {
    const motionPreference = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    );
    const updateMotionPreference = () => {
      setShouldReduceMotion(motionPreference.matches);

      if (motionPreference.matches) setAutoplayEnabled(false);
    };

    updateMotionPreference();
    motionPreference.addEventListener("change", updateMotionPreference);

    return () => {
      motionPreference.removeEventListener("change", updateMotionPreference);
    };
  }, []);

  useEffect(() => {
    if (!hasMounted.current) {
      hasMounted.current = true;
      return;
    }

    const slide = slideRef.current;
    if (!slide) return;

    const animation = slide.animate(
      shouldReduceMotion
        ? [{ opacity: 0.72 }, { opacity: 1 }]
        : [
            { opacity: 0, transform: "translateY(7px)" },
            { opacity: 1, transform: "translateY(0)" },
          ],
      {
        duration: shouldReduceMotion ? 160 : 240,
        easing: "cubic-bezier(0.23, 1, 0.32, 1)",
        fill: "both",
      },
    );

    return () => {
      animation.cancel();
    };
  }, [activeIndex, shouldReduceMotion]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      setIsDocumentVisible(document.visibilityState === "visible");
    };

    handleVisibilityChange();
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, []);

  useEffect(() => {
    if (!autoplayEnabled || isInteracting || !isDocumentVisible) return;

    const interval = window.setInterval(() => {
      setActiveIndex((currentIndex) => {
        return (currentIndex + 1) % testimonials.length;
      });
    }, AUTO_ADVANCE_MS);

    return () => {
      window.clearInterval(interval);
    };
  }, [activeIndex, autoplayEnabled, isDocumentVisible, isInteracting]);

  const handleBlur = (event: FocusEvent<HTMLDivElement>) => {
    const nextTarget = event.relatedTarget;
    if (
      nextTarget instanceof Node &&
      event.currentTarget.contains(nextTarget)
    ) {
      return;
    }

    setIsInteracting(false);
  };

  const testimonial = testimonials[activeIndex];
  const currentPosition = String(activeIndex + 1).padStart(2, "0");
  const total = String(testimonials.length).padStart(2, "0");

  return (
    <Container
      aria-labelledby="testimonials-title"
      as="section"
      className="grid grid-cols-1 gap-8 py-16 md:grid-cols-[auto_1fr] md:justify-between md:gap-16 md:py-28"
    >
      <Box data-landing-reveal>
        <Text
          as="h2"
          className="m-0 text-4xl md:text-5xl"
          id="testimonials-title"
        >
          How teams turn <br /> requests into progress.
        </Text>
      </Box>

      <Box
        className="border-border w-full justify-self-end border-b font-normal md:max-w-[720px]"
        data-landing-reveal
        onBlurCapture={handleBlur}
        onFocusCapture={() => {
          setIsInteracting(true);
        }}
        onPointerEnter={() => {
          setIsInteracting(true);
        }}
        onPointerLeave={() => {
          setIsInteracting(false);
        }}
      >
        <Box
          aria-atomic="true"
          aria-live={isInteracting ? "polite" : "off"}
          className="flex min-h-[330px] items-start pb-8 max-[760px]:min-h-[410px] max-[760px]:pb-[26px]"
        >
          <blockquote
            className="m-0 grid w-full grid-cols-[22px_minmax(0,1fr)] gap-[18px] max-[760px]:grid-cols-[18px_minmax(0,1fr)] max-[760px]:gap-3.5"
            ref={slideRef}
          >
            <span
              aria-hidden="true"
              className="text-primary font-serif text-[38px] leading-[0.95] max-[760px]:text-[32px]"
            >
              “
            </span>
            <Box>
              <Text className="m-0 font-serif text-[clamp(23px,2.5vw,30px)] leading-[1.42] font-normal tracking-[-0.025em] italic max-[760px]:text-[22px] max-[760px]:leading-[1.45]">
                {testimonial.quote}
              </Text>
              <Box as="footer" className="mt-[26px] flex flex-col gap-1">
                <Text as="cite" className="text-base font-medium not-italic">
                  {testimonial.author}
                </Text>
                <Text
                  as="span"
                  className="text-text-muted text-sm leading-relaxed"
                >
                  {testimonial.role} · {testimonial.company}
                </Text>
              </Box>
            </Box>
          </blockquote>
        </Box>

        <Box className="border-border flex items-center justify-between gap-6 border-t py-[18px] max-[760px]:py-3.5">
          <Text
            aria-hidden="true"
            as="span"
            className="text-text-muted text-sm tracking-[0.04em] [font-variant-numeric:tabular-nums]"
          >
            {currentPosition} / {total}
          </Text>

          <Box className="flex items-center gap-2">
            <button
              aria-label={
                autoplayEnabled
                  ? "Pause automatic testimonial rotation"
                  : "Play automatic testimonial rotation"
              }
              className={cn(testimonialButtonClassName, "px-3.5 text-sm")}
              onClick={() => {
                setAutoplayEnabled((enabled) => !enabled);
              }}
              type="button"
            >
              {autoplayEnabled ? "Pause" : "Play"}
            </button>
            <button
              aria-label="Previous testimonial"
              className={cn(
                testimonialButtonClassName,
                "inline-flex w-10 items-center justify-center p-0",
              )}
              onClick={() => {
                move(-1);
              }}
              type="button"
            >
              <TestimonialArrow direction="left" />
            </button>
            <button
              aria-label="Next testimonial"
              className={cn(
                testimonialButtonClassName,
                "inline-flex w-10 items-center justify-center p-0",
              )}
              onClick={() => {
                move(1);
              }}
              type="button"
            >
              <TestimonialArrow direction="right" />
            </button>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
