"use client";

import type { ImageProps } from "next/image";
import type { KeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { ProductScreenshot } from "@/modules/home/product-screenshot";
import styles from "./ai-planning-workflow.module.css";

const AUTO_ADVANCE_DELAY_MS = 7000;

export type FeatureWorkflowItem = {
  alt: string;
  darkImage: ImageProps["src"];
  description: string;
  label: string;
  lightImage: ImageProps["src"];
  title: string;
  value: string;
  url: string;
};

type FeatureProductWorkflowProps = {
  ariaLabel: string;
  description: string;
  heading: string;
  id: string;
  items: readonly FeatureWorkflowItem[];
};

export function FeatureProductWorkflow({
  ariaLabel,
  description,
  heading,
  id,
  items,
}: FeatureProductWorkflowProps) {
  const sectionRef = useRef<HTMLElement>(null);
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const [activeIndex, setActiveIndex] = useState(0);
  const [previousIndex, setPreviousIndex] = useState<number | null>(null);
  const [cycleVersion, setCycleVersion] = useState(0);
  const [isDocumentVisible, setIsDocumentVisible] = useState(true);
  const [isInView, setIsInView] = useState(false);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);
  const canAutoAdvance = isDocumentVisible && isInView && !prefersReducedMotion;

  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    const updateMotionPreference = () => {
      setPrefersReducedMotion(mediaQuery.matches);
    };

    updateMotionPreference();
    mediaQuery.addEventListener("change", updateMotionPreference);

    return () => {
      mediaQuery.removeEventListener("change", updateMotionPreference);
    };
  }, []);

  useEffect(() => {
    const section = sectionRef.current;

    if (!section || typeof IntersectionObserver === "undefined") {
      setIsInView(true);
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsInView(entry?.isIntersecting ?? false);
      },
      { threshold: 0.2 },
    );

    observer.observe(section);

    return () => {
      observer.disconnect();
    };
  }, []);

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
    if (!canAutoAdvance) return;

    const timeout = window.setTimeout(() => {
      setPreviousIndex(activeIndex);
      setActiveIndex((activeIndex + 1) % items.length);
      setCycleVersion((currentVersion) => currentVersion + 1);
    }, AUTO_ADVANCE_DELAY_MS);

    return () => {
      window.clearTimeout(timeout);
    };
  }, [activeIndex, canAutoAdvance, cycleVersion, items.length]);

  const selectWorkflow = (nextIndex: number) => {
    if (nextIndex !== activeIndex) {
      setPreviousIndex(activeIndex);
    }

    setActiveIndex(nextIndex);
    setCycleVersion((currentVersion) => currentVersion + 1);
  };

  const handleTabKeyDown = (
    event: KeyboardEvent<HTMLButtonElement>,
    currentIndex: number,
  ) => {
    let nextIndex: number | undefined;

    if (event.key === "ArrowRight") {
      nextIndex = (currentIndex + 1) % items.length;
    } else if (event.key === "ArrowLeft") {
      nextIndex = (currentIndex - 1 + items.length) % items.length;
    } else if (event.key === "Home") {
      nextIndex = 0;
    } else if (event.key === "End") {
      nextIndex = items.length - 1;
    }

    if (nextIndex === undefined) return;

    event.preventDefault();
    selectWorkflow(nextIndex);
    tabRefs.current[nextIndex]?.focus();
  };

  const headingId = `${id}-title`;

  return (
    <section
      aria-labelledby={headingId}
      className="scroll-mt-24 pt-16 pb-16 md:pt-36 md:pb-28"
      ref={sectionRef}
    >
      <Container>
        <Box className="mb-7 max-w-3xl md:mb-10" data-landing-reveal>
          <Text as="h2" className="text-3xl md:text-5xl" id={headingId}>
            {heading}
          </Text>
          <Text className="text-text-description mt-6 max-w-xl text-base text-pretty">
            {description}
          </Text>
        </Box>

        <Box className="overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <Box
            aria-label={ariaLabel}
            className="flex w-full min-w-[36rem] flex-nowrap gap-5 md:grid md:min-w-0 md:grid-cols-3"
            role="tablist"
          >
            {items.map((item, index) => {
              const isActive = index === activeIndex;
              let progressClassName = "scale-x-0";

              if (canAutoAdvance) {
                progressClassName = styles.progress;
              } else if (prefersReducedMotion) {
                progressClassName = "scale-x-100";
              }

              return (
                <button
                  aria-controls={`${id}-panel-${item.value}`}
                  aria-selected={isActive}
                  className="text-text-muted hover:text-foreground focus-visible:outline-foreground data-[state=active]:text-foreground relative h-14 w-full min-w-0 rounded-none border-0 bg-transparent px-0 py-4 text-left text-base font-normal whitespace-nowrap transition-colors focus-visible:outline-2 focus-visible:outline-offset-2"
                  data-state={isActive ? "active" : "inactive"}
                  id={`${id}-tab-${item.value}`}
                  key={item.value}
                  onClick={() => {
                    selectWorkflow(index);
                  }}
                  onKeyDown={(event) => {
                    handleTabKeyDown(event, index);
                  }}
                  ref={(tab) => {
                    tabRefs.current[index] = tab;
                  }}
                  role="tab"
                  tabIndex={isActive ? 0 : -1}
                  type="button"
                >
                  <span>{item.label}</span>
                  <span
                    aria-hidden="true"
                    className="bg-border absolute inset-x-0 bottom-0 h-px overflow-hidden"
                  >
                    {isActive ? (
                      <span
                        className={cn(
                          "bg-primary block h-0.5 w-full origin-left",
                          progressClassName,
                        )}
                        key={`${item.value}-${cycleVersion}`}
                      />
                    ) : null}
                  </span>
                </button>
              );
            })}
          </Box>
        </Box>
      </Container>

      <Box className={styles.viewport} data-landing-reveal>
        {items.map((item, index) => {
          const isActive = index === activeIndex;
          let slideClassName = styles.queuedSlide;

          if (isActive) {
            slideClassName = styles.activeSlide;
          } else if (index === previousIndex) {
            slideClassName = styles.exitingSlide;
          }

          return (
            <Box
              aria-hidden={!isActive}
              aria-labelledby={`${id}-tab-${item.value}`}
              className={cn(styles.slide, slideClassName)}
              id={`${id}-panel-${item.value}`}
              key={item.value}
              role="tabpanel"
              tabIndex={isActive ? 0 : -1}
            >
              <Container>
                <Box className="mt-8 grid gap-4 md:mt-10 md:grid-cols-[minmax(0,1.15fr)_minmax(18rem,0.85fr)] md:items-end md:gap-12">
                  <Text
                    as="h3"
                    className="max-w-sm text-2xl text-pretty md:text-4xl"
                  >
                    {item.title}
                  </Text>
                  <Text className="text-text-description max-w-md text-base leading-relaxed text-pretty">
                    {item.description}
                  </Text>
                </Box>
              </Container>
              <ProductScreenshot
                alt={item.alt}
                containerClassName="mt-8 md:mt-9"
                cropBrowserOnMobile
                darkImage={item.darkImage}
                lightImage={item.lightImage}
                priority={false}
                reveal={false}
                url={item.url}
              />
            </Box>
          );
        })}
      </Box>
    </section>
  );
}
