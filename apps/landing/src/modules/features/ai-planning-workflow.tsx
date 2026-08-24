"use client";

import type { KeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { ProductScreenshot } from "@/modules/home/product-screenshot";
import calendarWeekDark from "../../../public/images/product/calendar-week-dark.webp";
import calendarWeekLight from "../../../public/images/product/calendar-week-light.webp";
import mayaHomeDark from "../../../public/images/product/maya-home-dark.webp";
import mayaHomeLight from "../../../public/images/product/maya-home-light.webp";
import mayaObjectiveRisksDark from "../../../public/images/product/maya-objective-risks-dark.webp";
import mayaObjectiveRisksLight from "../../../public/images/product/maya-objective-risks-light.webp";
import styles from "./ai-planning-workflow.module.css";

const AUTO_ADVANCE_DELAY_MS = 7000;

const PLANNING_WORKFLOW = [
  {
    alt: "Maya ready to answer a planning question using the work already in FortyOne",
    darkImage: mayaHomeDark,
    description:
      "Ask in plain language. Maya starts from the goals, tasks, workload, and connected project context already in FortyOne.",
    label: "Ask in context",
    lightImage: mayaHomeLight,
    title: "Start with the decision the team needs to make.",
    value: "ask-in-context",
    url: "https://fortyone.app/maya",
  },
  {
    alt: "Maya identifying at-risk objectives and the delivery decisions that need attention",
    darkImage: mayaObjectiveRisksDark,
    description:
      "Bring capacity, timing, dependencies, and goal health into view before new work is added to the plan.",
    label: "See the tradeoffs",
    lightImage: mayaObjectiveRisksLight,
    title: "Understand what the next move will affect.",
    value: "see-the-tradeoffs",
    url: "https://fortyone.app/summary",
  },
  {
    alt: "FortyOne weekly calendar showing project work planned around the team's existing commitments",
    darkImage: calendarWeekDark,
    description:
      "Use workload and calendar availability to find a realistic start window before the team commits to more work.",
    label: "Find a work window",
    lightImage: calendarWeekLight,
    title: "Place the work where the team can deliver it.",
    value: "find-a-work-window",
    url: "https://fortyone.app/calendar",
  },
] as const;

export function AiPlanningWorkflow() {
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
      setActiveIndex((activeIndex + 1) % PLANNING_WORKFLOW.length);
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
      nextIndex = (currentIndex + 1) % PLANNING_WORKFLOW.length;
    } else if (event.key === "ArrowLeft") {
      nextIndex =
        (currentIndex - 1 + PLANNING_WORKFLOW.length) %
        PLANNING_WORKFLOW.length;
    } else if (event.key === "Home") {
      nextIndex = 0;
    } else if (event.key === "End") {
      nextIndex = PLANNING_WORKFLOW.length - 1;
    }

    if (nextIndex === undefined) return;

    event.preventDefault();
    selectWorkflow(nextIndex);
    tabRefs.current[nextIndex]?.focus();
  };

  return (
    <section
      aria-labelledby="ai-planning-workflow-title"
      className="scroll-mt-24 pt-16 pb-16 md:pt-36 md:pb-28"
      ref={sectionRef}
    >
      <Container>
        <Box className="mb-7 max-w-3xl md:mb-10" data-landing-reveal>
          <Text
            as="h2"
            className="text-3xl md:text-5xl"
            id="ai-planning-workflow-title"
          >
            Follow the path from question to approved plan.
          </Text>
          <Text className="text-text-description mt-6 max-w-xl text-base text-pretty">
            Maya turns the context already around your work into a planning
            recommendation your team can understand and control.
          </Text>
        </Box>

        <Box className="overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <Box
            aria-label="Explore the AI planning workflow"
            className="flex w-full min-w-[36rem] flex-nowrap gap-5 md:grid md:min-w-0 md:grid-cols-3"
            role="tablist"
          >
            {PLANNING_WORKFLOW.map((workflow, index) => {
              const isActive = index === activeIndex;
              let progressClassName = "scale-x-0";

              if (canAutoAdvance) {
                progressClassName = styles.progress;
              } else if (prefersReducedMotion) {
                progressClassName = "scale-x-100";
              }

              return (
                <button
                  aria-controls={`ai-planning-panel-${workflow.value}`}
                  aria-selected={isActive}
                  className="text-text-muted hover:text-foreground focus-visible:outline-foreground data-[state=active]:text-foreground relative h-14 w-full min-w-0 rounded-none border-0 bg-transparent px-0 py-4 text-left text-base font-normal whitespace-nowrap transition-colors focus-visible:outline-2 focus-visible:outline-offset-2"
                  data-state={isActive ? "active" : "inactive"}
                  id={`ai-planning-tab-${workflow.value}`}
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
        {PLANNING_WORKFLOW.map((workflow, index) => {
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
              aria-labelledby={`ai-planning-tab-${workflow.value}`}
              className={cn(styles.slide, slideClassName)}
              id={`ai-planning-panel-${workflow.value}`}
              key={workflow.value}
              role="tabpanel"
              tabIndex={isActive ? 0 : -1}
            >
              <Container>
                <Box className="mt-8 grid gap-4 md:mt-10 md:grid-cols-[minmax(0,1.15fr)_minmax(18rem,0.85fr)] md:items-end md:gap-12">
                  <Text
                    as="h3"
                    className="max-w-sm text-2xl text-pretty md:text-4xl"
                  >
                    {workflow.title}
                  </Text>
                  <Text className="text-text-description max-w-md text-base leading-relaxed text-pretty">
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
}
