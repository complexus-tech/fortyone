"use client";

import type { KeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import feedbackPortalDark from "../../../public/images/product/feedback-portal-dark.webp";
import feedbackPortalLight from "../../../public/images/product/feedback-portal-light.webp";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import myWorkBoardDark from "../../../public/images/product/my-work-board-dark.webp";
import myWorkBoardLight from "../../../public/images/product/my-work-board-light.webp";
import publicFeedbackRoadmapDark from "../../../public/images/product/public-feedback-roadmap-dark.webp";
import publicFeedbackRoadmapLight from "../../../public/images/product/public-feedback-roadmap-light.webp";
import roadmapTimelineDark from "../../../public/images/product/roadmap-timeline-dark.webp";
import roadmapTimelineLight from "../../../public/images/product/roadmap-timeline-light.webp";
import strategyMapDark from "../../../public/images/product/strategy-map-dark.webp";
import strategyMapLight from "../../../public/images/product/strategy-map-light.webp";
import { ProductScreenshot } from "./product-screenshot";
import styles from "./product-workflow-showcase.module.css";

const AUTO_ADVANCE_DELAY_MS = 7000;

const PRODUCT_WORKFLOWS = [
  {
    alt: "FortyOne Strategy Map connecting company goals to strategic pillars and objectives",
    darkImage: strategyMapDark,
    description:
      "Connect company goals to the outcomes every team is responsible for.",
    label: "Set direction",
    lightImage: strategyMapLight,
    title: "Turn strategy into a shared direction.",
    value: "set-direction",
    url: "https://fortyone.app/strategy-map",
  },
  {
    alt: "FortyOne internal roadmap showing objectives sequenced across a delivery timeline",
    darkImage: roadmapTimelineDark,
    description:
      "Sequence objectives on an internal roadmap with clear owners, dates, and dependencies.",
    label: "Build the roadmap",
    lightImage: roadmapTimelineLight,
    title: "Turn direction into a delivery path.",
    value: "build-the-roadmap",
    url: "https://fortyone.app/roadmap",
  },
  {
    alt: "FortyOne feedback portal showing customer requests, votes, and product feedback",
    darkImage: feedbackPortalDark,
    description:
      "Bring requests, votes, and context together before deciding what moves next.",
    label: "Listen to customers",
    lightImage: feedbackPortalLight,
    title: "Hear what customers need.",
    value: "listen-to-customers",
    url: "https://fortyone.app/feedback",
  },
  {
    alt: "FortyOne My Work Kanban board showing prioritized customer needs moving through delivery",
    darkImage: myWorkBoardDark,
    description:
      "Turn the strongest signals into owned work with a clear place in the plan.",
    label: "Prioritize",
    lightImage: myWorkBoardLight,
    title: "Move the right ideas into delivery.",
    value: "prioritize",
    url: "https://fortyone.app/my-work",
  },
  {
    alt: "FortyOne Maya AI delivery brief showing objective counts, completion metrics, and delivery trends",
    darkImage: mayaDeliveryBriefDark,
    description:
      "Let Maya summarize progress, surface delivery risks, and prepare the next decision.",
    label: "Plan with Maya",
    lightImage: mayaDeliveryBriefLight,
    title: "See risks before they slow the plan.",
    value: "plan-with-maya",
    url: "https://fortyone.app/maya",
  },
  {
    alt: "FortyOne public roadmap showing planned, in progress, and completed customer priorities",
    darkImage: publicFeedbackRoadmapDark,
    description:
      "Publish a clear roadmap so customers can follow planned, active, and completed work.",
    label: "Share progress",
    lightImage: publicFeedbackRoadmapLight,
    title: "Show customers what is moving.",
    value: "share-progress",
    url: "https://fortyone.app/feedback/roadmap",
  },
] as const;

export const ProductWorkflowShowcase = () => {
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
      setActiveIndex((activeIndex + 1) % PRODUCT_WORKFLOWS.length);
      setCycleVersion((currentVersion) => currentVersion + 1);
    }, AUTO_ADVANCE_DELAY_MS);

    return () => {
      window.clearTimeout(timeout);
    };
  }, [activeIndex, canAutoAdvance, cycleVersion]);

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
      nextIndex = (currentIndex + 1) % PRODUCT_WORKFLOWS.length;
    } else if (event.key === "ArrowLeft") {
      nextIndex =
        (currentIndex - 1 + PRODUCT_WORKFLOWS.length) %
        PRODUCT_WORKFLOWS.length;
    } else if (event.key === "Home") {
      nextIndex = 0;
    } else if (event.key === "End") {
      nextIndex = PRODUCT_WORKFLOWS.length - 1;
    }

    if (nextIndex === undefined) return;

    event.preventDefault();
    selectWorkflow(nextIndex);
    tabRefs.current[nextIndex]?.focus();
  };

  return (
    <section
      aria-labelledby="product-workflow-title"
      className="scroll-mt-24 pt-4 pb-16 md:pt-8 md:pb-28"
      ref={sectionRef}
    >
      <Container>
        <Box className="mb-7 max-w-3xl md:mb-10" data-landing-reveal>
          <Text
            as="h2"
            className="text-3xl md:text-5xl"
            id="product-workflow-title"
          >
            Follow the path from signal to progress.
          </Text>
          <Text className="text-text-description mt-6 max-w-lg text-base text-pretty">
            One connected story for deciding what matters, acting on it, and
            showing customers what changed.
          </Text>
        </Box>

        <Box className="snap-x snap-mandatory overflow-x-auto overscroll-x-contain [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <Box
            aria-label="Explore FortyOne product views"
            className="mx-0 flex w-full min-w-[60rem] flex-nowrap gap-5 rounded-none bg-transparent p-0 md:grid md:min-w-0 md:grid-cols-6 dark:bg-transparent"
            role="tablist"
          >
            {PRODUCT_WORKFLOWS.map((workflow, index) => {
              const isActive = index === activeIndex;
              let progressClassName = "scale-x-0";

              if (canAutoAdvance) {
                progressClassName = styles.progress;
              } else if (prefersReducedMotion) {
                progressClassName = "scale-x-100";
              }

              return (
                <button
                  aria-controls={`product-workflow-panel-${workflow.value}`}
                  aria-selected={isActive}
                  className="text-text-muted hover:text-foreground focus-visible:outline-foreground data-[state=active]:text-foreground relative h-14 w-full min-w-0 snap-start justify-start rounded-none border-0 bg-transparent px-0 py-4 text-left text-base font-normal whitespace-nowrap transition-colors hover:bg-transparent focus-visible:bg-transparent focus-visible:outline-2 focus-visible:outline-offset-2 data-[state=active]:border-0 data-[state=active]:bg-transparent"
                  data-state={isActive ? "active" : "inactive"}
                  id={`product-workflow-tab-${workflow.value}`}
                  key={workflow.value}
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
                  <span>{workflow.label}</span>
                  <span
                    aria-hidden="true"
                    className="bg-primary/20 dark:bg-border absolute inset-x-0 bottom-0 h-[1.5px] overflow-hidden dark:h-px"
                  >
                    {isActive ? (
                      <span
                        className={cn(
                          "bg-primary block h-full w-full origin-left",
                          progressClassName,
                        )}
                        key={`${workflow.value}-${cycleVersion}`}
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
        {PRODUCT_WORKFLOWS.map((workflow, index) => {
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
              aria-labelledby={`product-workflow-tab-${workflow.value}`}
              className={cn(styles.slide, slideClassName)}
              id={`product-workflow-panel-${workflow.value}`}
              key={workflow.value}
              role="tabpanel"
              tabIndex={isActive ? 0 : -1}
            >
              <Container>
                <Box className="mt-8 grid gap-4 md:mt-10 md:grid-cols-[minmax(0,1.15fr)_minmax(18rem,0.85fr)] md:items-end md:gap-12">
                  <Text
                    as="h3"
                    className="max-w-[18rem] text-left text-2xl text-pretty md:text-4xl"
                  >
                    {workflow.title}
                  </Text>
                  <Text className="text-text-description max-w-md text-left text-base leading-relaxed text-pretty md:justify-self-start">
                    {workflow.description}
                  </Text>
                </Box>
              </Container>
              <ProductScreenshot
                alt={workflow.alt}
                containerClassName="mt-8 md:mt-9"
                cropBrowserOnMobile
                darkImage={workflow.darkImage}
                lightImage={workflow.lightImage}
                priority={false}
                reveal={false}
                url={workflow.url}
              />
            </Box>
          );
        })}
      </Box>
    </section>
  );
};
