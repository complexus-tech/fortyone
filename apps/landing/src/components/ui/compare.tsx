"use client";

import type {
  ComponentPropsWithoutRef,
  ComponentType,
  CSSProperties,
} from "react";
import { useEffect, useRef, useState } from "react";
import { CheckIcon } from "icons";
import { cn } from "lib";
import { Box, Button, Text } from "ui";
import { SIGNUP_URL } from "@/lib/app-url";
import type { Plan, PlanFeatureKey } from "@/lib/plan-data";
import { plans } from "@/lib/plan-data";
import { Container } from "./container";
import {
  PricingAccessIcon,
  PricingAdminIcon,
  PricingAiIcon,
  PricingFeedbackIcon,
  PricingIntegrationIcon,
  PricingPlanningIcon,
  PricingSupportIcon,
  PricingTeamIcon,
  PricingWorkIcon,
} from "./pricing-icons";

type ComparisonIcon = ComponentType<ComponentPropsWithoutRef<"svg">>;
type ComparisonValue = boolean | string | undefined;

type ComparisonRow = {
  label: string;
  description: string;
  getValue: (plan: Plan) => ComparisonValue;
};

type ComparisonSection = {
  id: string;
  title: string;
  icon: ComparisonIcon;
  iconClassName: string;
  iconSurfaceClassName: string;
  rows: ComparisonRow[];
};

const featureRow = (
  feature: PlanFeatureKey,
  label: string,
  description: string,
): ComparisonRow => ({
  label,
  description,
  getValue: (plan) => plan.features[feature],
});

const comparisonSections: ComparisonSection[] = [
  {
    id: "usage",
    title: "Usage",
    icon: PricingWorkIcon,
    iconClassName: "text-info",
    iconSurfaceClassName: "bg-info/15",
    rows: [
      {
        label: "Members",
        description:
          "The number of people who can actively work in a workspace.",
        getValue: (plan) => plan.limits.members,
      },
      {
        label: "File uploads",
        description: "The maximum size supported for each uploaded file.",
        getValue: (plan) => plan.limits.fileUploads,
      },
      {
        label: "Tasks",
        description:
          "The number of tasks that can be created across the workspace.",
        getValue: (plan) => plan.limits.issues,
      },
    ],
  },
  {
    id: "planning",
    title: "Strategy & roadmaps",
    icon: PricingPlanningIcon,
    iconClassName: "text-primary",
    iconSurfaceClassName: "bg-primary/12",
    rows: [
      featureRow(
        "strategyMap",
        "Strategy map",
        "Connect the ultimate goal, strategic pillars, objectives, and key results.",
      ),
      featureRow(
        "objectives",
        "Objectives",
        "Set measurable objectives and connect them to active work.",
      ),
      featureRow(
        "trackOKRs",
        "OKR tracking",
        "Track outcomes, progress, and confidence from one place.",
      ),
      featureRow(
        "roadmaps",
        "Roadmaps",
        "Sequence objectives and communicate ownership, health, progress, and dates.",
      ),
      featureRow(
        "unlimitedEverything",
        "Unlimited planning",
        "Remove limits across the core planning experience.",
      ),
    ],
  },
  {
    id: "feedback",
    title: "Customer feedback",
    icon: PricingFeedbackIcon,
    iconClassName: "text-info",
    iconSurfaceClassName: "bg-info/15",
    rows: [
      featureRow(
        "customerFeedback",
        "Feedback boards",
        "Capture and organize customer requests in focused feedback boards.",
      ),
      featureRow(
        "feedbackVoting",
        "Voting and comments",
        "Let customers vote, comment, and follow the requests that matter to them.",
      ),
      featureRow(
        "publicRoadmap",
        "Public feedback portal and roadmap",
        "Keep customers informed as feedback moves from request to planned work and delivery.",
      ),
    ],
  },
  {
    id: "ai-planning",
    title: "Maya AI agent",
    icon: PricingAiIcon,
    iconClassName: "text-success",
    iconSurfaceClassName: "bg-success/15",
    rows: [
      featureRow(
        "mayaMessages",
        "AI agent messages",
        "Ask the Maya AI agent about workspace context, delivery risk, priorities, and next steps.",
      ),
      featureRow(
        "aiWorkPlanning",
        "AI work planning",
        "Let Maya suggest an owner and realistic delivery window from workload and availability.",
      ),
      featureRow(
        "calendarPlanning",
        "Calendar and focus planning",
        "See meetings and planned focus time together while scheduling project work.",
      ),
    ],
  },
  {
    id: "work-management",
    title: "Work management",
    icon: PricingWorkIcon,
    iconClassName: "text-secondary",
    iconSurfaceClassName: "bg-secondary/15",
    rows: [
      featureRow(
        "myWork",
        "My Work",
        "Bring assigned, created, and followed work into one personal view.",
      ),
      featureRow(
        "kanbanAndList",
        "Kanban and list views",
        "Move between visual workflow boards and structured task lists.",
      ),
      featureRow(
        "sprints",
        "Sprints",
        "Plan time-boxed delivery cycles and keep active sprint work visible.",
      ),
      featureRow(
        "summaryAnalytics",
        "Summary and analytics",
        "Monitor workspace activity, workload, progress, and delivery signals.",
      ),
      featureRow(
        "documents",
        "Shared project documents",
        "Keep project context beside the stories and objectives it supports.",
      ),
    ],
  },
  {
    id: "integrations",
    title: "Integrations",
    icon: PricingIntegrationIcon,
    iconClassName: "text-primary",
    iconSurfaceClassName: "bg-primary/12",
    rows: [
      featureRow(
        "slackIntegration",
        "Slack",
        "Create structured work from conversations and ask Maya from Slack.",
      ),
      featureRow(
        "githubIntegration",
        "GitHub",
        "Sync issues and keep engineering delivery connected to project work.",
      ),
      featureRow(
        "figmaIntegration",
        "Figma",
        "Bring design context into stories without recreating it by hand.",
      ),
      featureRow(
        "googleCalendarIntegration",
        "Google Calendar",
        "Plan project work around meetings and real personal availability.",
      ),
    ],
  },
  {
    id: "collaboration",
    title: "Collaboration",
    icon: PricingTeamIcon,
    iconClassName: "text-info",
    iconSurfaceClassName: "bg-info/15",
    rows: [
      featureRow(
        "teams",
        "Teams",
        "Create focused team spaces inside the same workspace.",
      ),
      featureRow(
        "unlimitedGuests",
        "Guests",
        "Bring external collaborators into the work they need to see.",
      ),
      featureRow(
        "customWorkflows",
        "Custom workflows",
        "Adapt statuses and workflows to the way each team works.",
      ),
    ],
  },
  {
    id: "administration",
    title: "Administration",
    icon: PricingAdminIcon,
    iconClassName: "text-secondary",
    iconSurfaceClassName: "bg-secondary/15",
    rows: [
      featureRow(
        "rbac",
        "Role-based access control",
        "Set workspace permissions by role and responsibility.",
      ),
      featureRow(
        "privateTeams",
        "Private teams",
        "Restrict sensitive team spaces to selected members.",
      ),
      featureRow(
        "customTerminology",
        "Custom terminology",
        "Match FortyOne labels to the language your organization uses.",
      ),
    ],
  },
  {
    id: "security",
    title: "Security",
    icon: PricingAccessIcon,
    iconClassName: "text-secondary",
    iconSurfaceClassName: "bg-secondary/15",
    rows: [
      featureRow(
        "sso",
        "Single Sign-On (SSO)",
        "Let members sign in through your identity provider.",
      ),
      featureRow(
        "onPremise",
        "Private deployment",
        "Deploy on-premise or in a private cloud environment.",
      ),
    ],
  },
  {
    id: "support",
    title: "Support",
    icon: PricingSupportIcon,
    iconClassName: "text-success",
    iconSurfaceClassName: "bg-success/15",
    rows: [
      featureRow(
        "emailSupport",
        "Email support",
        "Get help from the FortyOne support team by email.",
      ),
      featureRow(
        "prioritySupport",
        "Priority support",
        "Move critical support requests to the front of the queue.",
      ),
      featureRow(
        "customOnboarding",
        "Custom onboarding",
        "Launch with a guided onboarding plan for your organization.",
      ),
      featureRow(
        "dedicatedManager",
        "Dedicated account manager",
        "Work with a consistent point of contact as you scale.",
      ),
      featureRow(
        "volumeDiscounts",
        "Volume discounts",
        "Access custom commercial terms for larger deployments.",
      ),
    ],
  },
];

const planCtas: Record<string, { href: string; label: string }> = {
  Hobby: { href: SIGNUP_URL, label: "Start free" },
  Professional: { href: SIGNUP_URL, label: "Try Professional" },
  Business: { href: SIGNUP_URL, label: "Try Business" },
  Enterprise: { href: "mailto:info@complexus.app", label: "Talk to sales" },
};

const COMPARISON_DOCK_TOP = 80;

const PLAN_COLUMN_SURFACE: CSSProperties = {
  background:
    "linear-gradient(to right, transparent 0 0.55rem, color-mix(in oklab, var(--color-accent) 28%, transparent) 0.55rem calc(100% - 0.55rem), transparent calc(100% - 0.55rem))",
};

const FeatureValue = ({ value }: { value: ComparisonValue }) => {
  if (typeof value === "string") {
    return <Text className="text-sm leading-5">{value}</Text>;
  }

  if (value) {
    return (
      <span className="bg-background-inverse text-foreground-inverse inline-flex size-5 items-center justify-center rounded-[0.45rem]">
        <CheckIcon
          aria-label="Included"
          className="h-3.5 w-auto"
          strokeWidth={2.7}
        />
      </span>
    );
  }

  return <span aria-label="Not included">—</span>;
};

const ComparisonCategoryIcon = ({
  section,
}: {
  section: ComparisonSection;
}) => {
  const Icon = section.icon;

  return (
    <span
      className={cn(
        "inline-flex size-7 shrink-0 items-center justify-center rounded-lg",
        section.iconSurfaceClassName,
      )}
    >
      <Icon
        aria-hidden="true"
        className={cn("size-[1.125rem]", section.iconClassName)}
        strokeWidth={1.7}
      />
    </span>
  );
};

const PlanHeaderContent = ({ plan }: { plan: Plan }) => {
  const cta = planCtas[plan.name];

  return (
    <>
      <Text className="mb-3 text-base font-semibold">{plan.name}</Text>
      <Button
        align="center"
        className="px-4 text-sm whitespace-nowrap"
        color={plan.highlighted ? "invert" : "tertiary"}
        fullWidth
        href={cta.href}
        rounded="md"
        size="lg"
        variant={plan.highlighted ? "solid" : "outline"}
      >
        {cta.label}
      </Button>
    </>
  );
};

export const ComparePlans = () => {
  const dockAnchorRef = useRef<HTMLDivElement>(null);
  const comparisonEndRef = useRef<HTMLDivElement>(null);
  const [dockStyle, setDockStyle] = useState<CSSProperties>();

  useEffect(() => {
    let animationFrame: number | undefined;

    const updateDockPosition = () => {
      if (animationFrame) return;

      animationFrame = window.requestAnimationFrame(() => {
        animationFrame = undefined;

        const anchor = dockAnchorRef.current;
        const comparisonEnd = comparisonEndRef.current;

        if (!anchor || !comparisonEnd || window.innerWidth < 768) {
          setDockStyle((currentStyle) =>
            currentStyle ? undefined : currentStyle,
          );
          return;
        }

        const anchorBounds = anchor.getBoundingClientRect();
        const comparisonEndBounds = comparisonEnd.getBoundingClientRect();
        const shouldDock =
          anchorBounds.top <= COMPARISON_DOCK_TOP &&
          comparisonEndBounds.bottom >
            COMPARISON_DOCK_TOP + anchorBounds.height;

        setDockStyle((currentStyle) => {
          if (!shouldDock) return currentStyle ? undefined : currentStyle;
          if (
            currentStyle?.left === anchorBounds.left &&
            currentStyle.width === anchorBounds.width
          ) {
            return currentStyle;
          }

          return {
            left: anchorBounds.left,
            position: "fixed",
            top: COMPARISON_DOCK_TOP,
            width: anchorBounds.width,
          };
        });
      });
    };

    updateDockPosition();
    window.addEventListener("resize", updateDockPosition);
    window.addEventListener("scroll", updateDockPosition, { passive: true });

    return () => {
      if (animationFrame) window.cancelAnimationFrame(animationFrame);
      window.removeEventListener("resize", updateDockPosition);
      window.removeEventListener("scroll", updateDockPosition);
    };
  }, []);

  return (
    <Container className="max-w-300 pt-24 pb-8 md:pt-32 md:pb-12">
      <Box className="mx-auto mb-14 max-w-2xl text-center md:mb-16">
        <Text
          as="h2"
          className="font-serif text-4xl leading-tight font-medium tracking-[-0.035em] md:text-5xl"
        >
          Compare plans
        </Text>
      </Box>

      <Box className="hidden min-h-[7.25rem] md:block" ref={dockAnchorRef}>
        <Box
          className={cn(
            "border-border/60 bg-background grid grid-cols-[30%_repeat(4,17.5%)] border-b",
            dockStyle && "z-10",
          )}
          style={dockStyle}
        >
          <Box className="p-4" />
          {plans.map((plan) => (
            <Box className="px-4 pt-4 pb-5 text-center" key={plan.name}>
              <PlanHeaderContent plan={plan} />
            </Box>
          ))}
        </Box>
      </Box>

      <Box
        className="overflow-x-auto md:overflow-visible"
        ref={comparisonEndRef}
      >
        <table className="w-full min-w-[64rem] table-fixed border-separate border-spacing-0 text-left">
          <colgroup>
            <col className="w-[30%]" />
            {plans.map((plan) => (
              <col className="w-[17.5%]" key={plan.name} />
            ))}
          </colgroup>
          <thead className="md:sr-only">
            <tr>
              <th
                className="border-border/60 bg-background border-b p-4"
                scope="col"
              />
              {plans.map((plan) => (
                <th
                  className="border-border/60 bg-background border-b px-4 pt-4 pb-5 text-center"
                  key={plan.name}
                  scope="col"
                >
                  <PlanHeaderContent plan={plan} />
                </th>
              ))}
            </tr>
          </thead>
          {comparisonSections.map((section) => (
            <tbody id={`comparison-${section.id}`} key={section.id}>
              <tr>
                <th
                  className="border-border/60 bg-background scroll-mt-36 border-b px-4 pt-9 pb-5"
                  scope="colgroup"
                >
                  <span className="flex items-center gap-3">
                    <ComparisonCategoryIcon section={section} />
                    <span className="text-xl font-semibold">
                      {section.title}
                    </span>
                  </span>
                </th>
                {plans.map((plan) => (
                  <td
                    className="border-border/60 border-b"
                    key={`${section.id}-${plan.name}-heading`}
                    style={PLAN_COLUMN_SURFACE}
                  />
                ))}
              </tr>
              {section.rows.map((row) => (
                <tr className="[&>td]:border-b [&>th]:border-b" key={row.label}>
                  <th
                    className="border-border/60 px-4 py-[1.125rem] font-normal"
                    scope="row"
                  >
                    <span
                      className="decoration-border inline border-b border-dotted text-sm leading-5"
                      title={row.description}
                    >
                      {row.label}
                    </span>
                  </th>
                  {plans.map((plan) => (
                    <td
                      className="border-border/60 px-4 py-[1.125rem] text-center align-middle"
                      key={`${section.id}-${row.label}-${plan.name}`}
                      style={PLAN_COLUMN_SURFACE}
                    >
                      <FeatureValue value={row.getValue(plan)} />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          ))}
        </table>
      </Box>
    </Container>
  );
};
