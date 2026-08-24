"use client";

import type { KeyboardEvent } from "react";
import type { StaticImageData } from "next/image";
import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import { cn } from "lib";
import { Box, Button, Text } from "ui";
import { Container } from "@/components/ui";
import { PricingArrowIcon } from "@/components/ui/pricing-icons";
import aiPlanningImage from "../../../public/images/workflows/ai-planning.webp";
import customerFeedbackImage from "../../../public/images/workflows/customer-feedback.webp";
import goalsImage from "../../../public/images/workflows/goals.webp";
import roadmapsImage from "../../../public/images/workflows/roadmaps.webp";
import tasksImage from "../../../public/images/workflows/tasks.webp";
import styles from "./feature-overview.module.css";

type FeatureOverviewItem = {
  alt: string;
  ctaLabel: string;
  description: string;
  href: string;
  image: StaticImageData;
  label: string;
  title: string;
  value: string;
};

const AUTO_ADVANCE_DELAY_MS = 7000;

const FEATURES: readonly FeatureOverviewItem[] = [
  {
    alt: "A product researcher listening closely during a customer feedback interview",
    ctaLabel: "Explore feedback",
    description:
      "Collect requests, votes, and context in one place, then connect the strongest signals to planned work.",
    href: "/features/customer-feedback",
    image: customerFeedbackImage,
    label: "Customer feedback",
    title: "Hear what customers need.",
    value: "customer-feedback",
  },
  {
    alt: "Three leaders aligning company priorities around a planning table",
    ctaLabel: "Explore goals",
    description:
      "Turn company goals into clear priorities, then connect them to the objectives and work that move progress forward.",
    href: "/features/goals",
    image: goalsImage,
    label: "Goals",
    title: "Make strategy visible.",
    value: "goals",
  },
  {
    alt: "Two teammates arranging the next pieces of work at a studio table",
    ctaLabel: "Explore tasks",
    description:
      "Turn ideas and requests into owned, trackable work with the context your team needs to keep moving.",
    href: "/features/tasks",
    image: tasksImage,
    label: "Tasks",
    title: "Give every task a clear path.",
    value: "tasks",
  },
  {
    alt: "A product lead reviewing an AI-assisted delivery plan at a desk",
    ctaLabel: "Explore AI",
    description:
      "Maya checks priorities, ownership, workload, and timing, then brings you a delivery brief to review.",
    href: "/features/ai-planning",
    image: aiPlanningImage,
    label: "AI planning",
    title: "Let Maya prepare the plan.",
    value: "ai-planning",
  },
  {
    alt: "A cross-functional team tracing a delivery roadmap on a planning wall",
    ctaLabel: "Explore roadmaps",
    description:
      "Bring priorities, owners, timing, and delivery risk into one roadmap that stays connected to the work.",
    href: "/features/roadmaps",
    image: roadmapsImage,
    label: "Roadmaps",
    title: "Keep the plan clear as work changes.",
    value: "roadmaps",
  },
] as const;

export const FeatureOverview = () => {
  const sectionRef = useRef<HTMLElement>(null);
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const [activeIndex, setActiveIndex] = useState(0);
  const [cycleVersion, setCycleVersion] = useState(0);
  const [isDocumentVisible, setIsDocumentVisible] = useState(true);
  const [isInView, setIsInView] = useState(true);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  const activeFeature = FEATURES[activeIndex];
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
      { threshold: 0.25 },
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
    if (!canAutoAdvance) {
      return;
    }

    const timeout = window.setTimeout(() => {
      setActiveIndex((currentIndex) => (currentIndex + 1) % FEATURES.length);
      setCycleVersion((currentVersion) => currentVersion + 1);
    }, AUTO_ADVANCE_DELAY_MS);

    return () => {
      window.clearTimeout(timeout);
    };
  }, [activeIndex, canAutoAdvance, cycleVersion]);

  const selectFeature = (nextIndex: number) => {
    setActiveIndex(nextIndex);
    setCycleVersion((currentVersion) => currentVersion + 1);
  };

  const handleTabKeyDown = (
    event: KeyboardEvent<HTMLButtonElement>,
    currentIndex: number,
  ) => {
    let nextIndex: number | undefined;

    if (event.key === "ArrowRight") {
      nextIndex = (currentIndex + 1) % FEATURES.length;
    } else if (event.key === "ArrowLeft") {
      nextIndex = (currentIndex - 1 + FEATURES.length) % FEATURES.length;
    } else if (event.key === "Home") {
      nextIndex = 0;
    } else if (event.key === "End") {
      nextIndex = FEATURES.length - 1;
    }

    if (nextIndex === undefined) {
      return;
    }

    event.preventDefault();
    selectFeature(nextIndex);
    tabRefs.current[nextIndex]?.focus();
  };

  return (
    <section
      aria-labelledby="feature-overview-title"
      className="scroll-mt-24 py-16 md:py-28"
      id="feature-overview"
      ref={sectionRef}
    >
      <Container>
        <Box className="max-w-3xl" data-landing-reveal>
          <Text
            as="h2"
            className="text-3xl md:text-5xl"
            id="feature-overview-title"
          >
            See how work moves from idea to outcome.
          </Text>
          <Text className="text-text-description mt-6 max-w-2xl text-base text-pretty">
            Explore the connected workflows that help your team listen, decide,
            plan, and deliver.
          </Text>
        </Box>

        <Box className="mt-20 md:mt-24">
          <Box>
            <Box className="overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
              <Box
                aria-label="Explore FortyOne workflows"
                className="mx-0 flex w-full min-w-[50rem] flex-nowrap gap-5 rounded-none bg-transparent p-0 md:grid md:min-w-0 md:grid-cols-5 dark:bg-transparent"
                role="tablist"
              >
                {FEATURES.map((feature, index) => {
                  const isActive = index === activeIndex;
                  let progressClassName = "scale-x-0";

                  if (canAutoAdvance) {
                    progressClassName = styles.progress;
                  } else if (prefersReducedMotion) {
                    progressClassName = "scale-x-100";
                  }

                  return (
                    <button
                      aria-controls={`feature-overview-panel-${feature.value}`}
                      aria-selected={isActive}
                      className="text-text-muted hover:text-foreground focus-visible:outline-foreground data-[state=active]:text-foreground relative h-14 w-full min-w-0 justify-start rounded-none border-0 bg-transparent px-0 py-4 text-left text-base font-normal whitespace-nowrap transition-colors hover:bg-transparent focus-visible:bg-transparent focus-visible:outline-2 focus-visible:outline-offset-2 data-[state=active]:border-0 data-[state=active]:bg-transparent"
                      data-state={isActive ? "active" : "inactive"}
                      id={`feature-overview-tab-${feature.value}`}
                      key={feature.value}
                      onClick={() => {
                        selectFeature(index);
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
                      <span>{feature.label}</span>
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
                            key={`${feature.value}-${cycleVersion}`}
                          />
                        ) : null}
                      </span>
                    </button>
                  );
                })}
              </Box>
            </Box>

            <Box
              aria-labelledby={`feature-overview-tab-${activeFeature.value}`}
              className="mt-8 outline-none md:mt-9"
              id={`feature-overview-panel-${activeFeature.value}`}
              role="tabpanel"
              tabIndex={0}
            >
              <Box className="bg-surface-muted/45 dark:bg-surface-elevated/45 grid overflow-hidden rounded p-3 md:min-h-[30rem] md:grid-cols-2 md:gap-5">
                <Box className="flex items-center px-6 py-10 sm:px-8 md:px-6 md:py-12 md:pr-12">
                  <Box className="max-w-md">
                    <Text as="h3" className="text-2xl md:text-3xl">
                      {activeFeature.title}
                    </Text>
                    <Text className="text-text-description mt-4 text-base leading-relaxed text-pretty">
                      {activeFeature.description}
                    </Text>
                    <Button
                      className="mt-8"
                      color="gradient"
                      href={activeFeature.href}
                      prefetch
                      rightIcon={
                        <PricingArrowIcon
                          aria-hidden="true"
                          className="h-4.5 w-auto"
                          strokeWidth={2}
                        />
                      }
                      rounded="lg"
                      size="lg"
                      variant="outline"
                    >
                      {activeFeature.ctaLabel}
                    </Button>
                  </Box>
                </Box>

                <Box className="relative min-h-[21rem] overflow-hidden rounded-sm md:min-h-0">
                  {FEATURES.map((feature, index) => {
                    const isActiveImage = index === activeIndex;

                    return (
                      <Image
                        alt={isActiveImage ? feature.alt : ""}
                        aria-hidden={!isActiveImage}
                        className={cn(
                          "h-full w-full object-cover object-center",
                          styles.image,
                          isActiveImage && styles.imageActive,
                        )}
                        fill
                        key={feature.value}
                        priority={index === 0}
                        sizes="(max-width: 767px) calc(100vw - 3rem), 50vw"
                        src={feature.image}
                      />
                    );
                  })}
                </Box>
              </Box>
            </Box>
          </Box>
        </Box>
      </Container>
    </section>
  );
};
