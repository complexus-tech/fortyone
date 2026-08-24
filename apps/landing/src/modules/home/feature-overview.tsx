import type { ComponentPropsWithoutRef, ComponentType } from "react";
import Link from "next/link";
import {
  AiIcon,
  ArrowRightIcon,
  CheckListIcon,
  GoalIcon,
  RequestsIcon,
  RoadmapIcon,
} from "icons";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";

type FeatureSummary = {
  description: string;
  href: string;
  icon: ComponentType<ComponentPropsWithoutRef<"svg">>;
  title: string;
};

const featureSummaries: readonly FeatureSummary[] = [
  {
    description:
      "Capture requests, votes, and the context behind what customers need.",
    href: "/features/customer-feedback",
    icon: RequestsIcon,
    title: "Customer feedback",
  },
  {
    description:
      "Connect priorities to the work and progress that moves them forward.",
    href: "/features/goals",
    icon: GoalIcon,
    title: "Goals",
  },
  {
    description: "Turn rough requests into clear, owned, trackable work.",
    href: "/features/tasks",
    icon: CheckListIcon,
    title: "Tasks",
  },
  {
    description:
      "Let Maya prepare owners, effort, timing, and delivery risks for review.",
    href: "/features/ai-planning",
    icon: AiIcon,
    title: "AI planning",
  },
  {
    description:
      "Keep priorities linked to goals, tasks, owners, and delivery status.",
    href: "/features/roadmaps",
    icon: RoadmapIcon,
    title: "Roadmaps",
  },
];

export const FeatureOverview = () => {
  return (
    <Container
      aria-labelledby="feature-overview-title"
      as="section"
      className="scroll-mt-24 py-16 md:py-28"
      id="feature-overview"
    >
      <Box className="max-w-3xl" data-landing-reveal>
        <Text
          as="h2"
          className="text-3xl md:text-5xl"
          id="feature-overview-title"
        >
          Explore each workflow.
        </Text>
        <Text className="text-text-muted mt-6 max-w-2xl text-pretty">
          Start with the overview, then open a feature for the complete workflow
          and product details.
        </Text>
      </Box>

      <Box
        as="ul"
        className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 md:mt-14 xl:grid-cols-6"
      >
        {featureSummaries.map((feature, index) => {
          const Icon = feature.icon;

          return (
            <Box
              as="li"
              className={cn(
                "min-w-0",
                index < 3 ? "xl:col-span-2" : "xl:col-span-3",
                index === featureSummaries.length - 1 &&
                  "sm:col-span-2 xl:col-span-3",
              )}
              data-landing-reveal
              key={feature.href}
              style={{ transitionDelay: `${index * 45}ms` }}
            >
              <Link
                className="bg-surface-muted/65 hover:bg-surface-muted focus-visible:outline-primary dark:bg-surface-elevated/70 dark:hover:bg-surface-elevated group flex h-full min-h-52 flex-col rounded-xl p-6 transition-colors focus-visible:outline-2 focus-visible:outline-offset-4 md:p-7"
                href={feature.href}
              >
                <Box className="bg-background/75 text-primary dark:bg-surface-prominent/70 flex size-10 items-center justify-center rounded-lg shadow-sm shadow-black/5">
                  <Icon aria-hidden="true" className="size-5" />
                </Box>

                <Box className="mt-auto pt-10">
                  <Text as="h3" className="text-lg font-semibold">
                    {feature.title}
                  </Text>
                  <Text className="text-text-muted mt-2 max-w-sm leading-relaxed">
                    {feature.description}
                  </Text>
                  <Text
                    as="span"
                    className="text-foreground mt-5 inline-flex items-center gap-2 text-sm font-medium"
                  >
                    Explore feature
                    <ArrowRightIcon
                      aria-hidden="true"
                      className="size-4 transition-transform group-hover:translate-x-0.5"
                    />
                  </Text>
                </Box>
              </Link>
            </Box>
          );
        })}
      </Box>
    </Container>
  );
};
