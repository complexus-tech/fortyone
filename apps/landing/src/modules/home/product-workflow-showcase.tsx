"use client";

import type { FocusEvent, KeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { cn } from "lib";
import { Box } from "ui";
import { Container } from "@/components/ui";
import feedbackPortalDark from "../../../public/images/product/feedback-portal-dark.webp";
import feedbackPortalLight from "../../../public/images/product/feedback-portal-light.webp";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
import myWorkBoardDark from "../../../public/images/product/my-work-board-dark.webp";
import myWorkBoardLight from "../../../public/images/product/my-work-board-light.webp";
import roadmapTimelineDark from "../../../public/images/product/roadmap-timeline-dark.webp";
import roadmapTimelineLight from "../../../public/images/product/roadmap-timeline-light.webp";
import strategyMapDark from "../../../public/images/product/strategy-map-dark.webp";
import strategyMapLight from "../../../public/images/product/strategy-map-light.webp";
import { ProductScreenshot } from "./product-screenshot";
import styles from "./product-workflow-showcase.module.css";

const AUTO_ADVANCE_DELAY_MS = 7000;

const PRODUCT_WORKFLOWS = [
  {
    alt: "FortyOne feedback portal showing customer requests, votes, and product feedback",
    darkImage: feedbackPortalDark,
    label: "Customer feedback",
    lightImage: feedbackPortalLight,
    value: "customer-feedback",
    url: "https://fortyone.app/feedback",
  },
  {
    alt: "FortyOne Strategy Map connecting company goals to strategic pillars and objectives",
    darkImage: strategyMapDark,
    label: "Goals",
    lightImage: strategyMapLight,
    value: "goals",
    url: "https://fortyone.app/strategy-map",
  },
  {
    alt: "FortyOne My Work board showing owned tasks moving through delivery",
    darkImage: myWorkBoardDark,
    label: "Tasks",
    lightImage: myWorkBoardLight,
    value: "tasks",
    url: "https://fortyone.app/my-work",
  },
  {
    alt: "FortyOne Maya AI delivery brief showing objective counts, completion metrics, and delivery trends",
    darkImage: mayaDeliveryBriefDark,
    label: "AI planning",
    lightImage: mayaDeliveryBriefLight,
    value: "ai-planning",
    url: "https://fortyone.app/maya",
  },
  {
    alt: "FortyOne Roadmap timeline showing coordinated initiatives, milestones, owners, and forecasts",
    darkImage: roadmapTimelineDark,
    label: "Roadmaps",
    lightImage: roadmapTimelineLight,
    value: "roadmaps",
    url: "https://fortyone.app/roadmap",
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
  const [isInteractionPaused, setIsInteractionPaused] = useState(false);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  const canAutoAdvance =
    isDocumentVisible &&
    isInView &&
    !isInteractionPaused &&
    !prefersReducedMotion;

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
  }, [activeIndex, canAutoAdvance]);

  const selectWorkflow = (nextIndex: number) => {
    if (nextIndex === activeIndex) return;

    setPreviousIndex(activeIndex);
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

  const handleBlur = (event: FocusEvent<HTMLElement>) => {
    if (!event.currentTarget.contains(event.relatedTarget)) {
      setIsInteractionPaused(false);
    }
  };

  return (
    <section
      aria-label="Explore FortyOne product workflows"
      className="scroll-mt-24 pt-4 pb-16 md:pt-8 md:pb-28"
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
      ref={sectionRef}
    >
      <Container>
        <Box className="overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <Box
            aria-label="Explore FortyOne product views"
            className="mx-0 flex w-full min-w-[50rem] flex-nowrap gap-5 rounded-none bg-transparent p-0 md:grid md:min-w-0 md:grid-cols-5 dark:bg-transparent"
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
                  className="text-text-muted hover:text-foreground focus-visible:outline-foreground data-[state=active]:text-foreground relative h-14 w-full min-w-0 justify-start rounded-none border-0 bg-transparent px-0 py-4 text-left text-base font-normal whitespace-nowrap transition-colors hover:bg-transparent focus-visible:bg-transparent focus-visible:outline-2 focus-visible:outline-offset-2 data-[state=active]:border-0 data-[state=active]:bg-transparent"
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
                    className="bg-border absolute inset-x-0 bottom-0 h-px overflow-hidden"
                  >
                    {isActive ? (
                      <span
                        className={cn(
                          "bg-primary block h-0.5 w-full origin-left",
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
