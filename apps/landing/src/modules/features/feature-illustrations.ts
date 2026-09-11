import type { MarketingVisual } from "@/components/shared/marketing-detail-page";
import type { IntegrationBrandName } from "@/modules/integrations/integration-brand";
import type { PreviewKind } from "./marketing-visual-card";

type FeatureIllustration = {
  kind: PreviewKind;
  brand?: IntegrationBrandName;
  visual: MarketingVisual;
};

const FEATURE_ILLUSTRATIONS: Record<string, readonly FeatureIllustration[]> = {
  "customer-feedback": [
    {
      kind: "feedback",
      visual: {
        heading: "Customer feedback",
        subheading: "One place for customer demand",
        badge: "12 votes",
        rows: [
          { label: "Request", value: "Make onboarding easier" },
          { label: "Board", value: "Product ideas" },
        ],
        note: "Requests, votes, and context together",
      },
    },
    {
      kind: "work",
      visual: {
        heading: "From request to work",
        subheading: "The original evidence stays attached",
        badge: "Planned",
        rows: [
          { label: "PRD-142", value: "Redesign onboarding flow" },
          { label: "Source", value: "Make onboarding easier · 12 votes" },
        ],
        note: "Goal · Improve activation",
      },
    },
    {
      kind: "goals",
      visual: {
        heading: "Priority review",
        subheading: "Choose the work with a clear case",
        rows: [
          {
            label: "Customer demand",
            value: "12 votes across related requests",
          },
          { label: "Outcome", value: "Improve activation" },
        ],
        note: "Evidence before commitment",
      },
    },
  ],
  goals: [
    {
      kind: "goals",
      visual: {
        heading: "Company direction",
        subheading: "A shared outcome for the team",
        badge: "On track",
        rows: [
          { label: "Objective", value: "Improve customer activation" },
          { label: "Progress", value: "64%", width: "64%" },
        ],
        note: "Strategy connected to execution",
      },
    },
    {
      kind: "work",
      visual: {
        heading: "Work with a purpose",
        subheading: "Every task supports the outcome",
        rows: [
          { label: "PRD-142", value: "Redesign onboarding flow" },
          { label: "Goal", value: "Improve customer activation" },
        ],
        note: "One task. A clear reason.",
      },
    },
    {
      kind: "planning",
      visual: {
        heading: "Goal health",
        subheading: "The next decision is in view",
        badge: "Review",
        rows: [
          { label: "At risk", value: "2 tasks need decisions" },
          { label: "Next step", value: "Resolve onboarding dependencies" },
        ],
        note: "Review progress with the work attached",
      },
    },
  ],
  tasks: [
    {
      kind: "work",
      visual: {
        heading: "Ready for delivery",
        subheading: "A task with the essentials attached",
        badge: "Planned",
        rows: [
          { label: "Task", value: "Prepare onboarding handoff" },
          { label: "Owner · Estimate", value: "Product team · 4 hours" },
        ],
        note: "Source context stays with the task",
      },
    },
    {
      kind: "planning",
      visual: {
        heading: "A realistic work window",
        subheading: "Workload before another assignment",
        rows: [
          { label: "Current load", value: "68%", width: "68%" },
          { label: "First block", value: "Tuesday · 10:30" },
        ],
        note: "Review the proposal with your team",
      },
    },
    {
      kind: "roadmap",
      visual: {
        heading: "Follow the work",
        subheading: "A clear path through delivery",
        rows: [
          { label: "In progress", value: "Build onboarding flow" },
          { label: "Review", value: "Confirm the customer handoff" },
        ],
        note: "Ownership and context at every step",
      },
    },
  ],
  roadmaps: [
    {
      kind: "roadmap",
      visual: {
        heading: "Roadmap priorities",
        subheading: "From direction to delivery",
        badge: "Q3",
        rows: [
          { label: "Now", value: "Improve onboarding" },
          { label: "Next", value: "Simplify account setup" },
          { label: "Later", value: "Refresh reporting" },
        ],
        note: "Priorities with a path forward",
      },
    },
    {
      kind: "work",
      visual: {
        heading: "Launch readiness",
        subheading: "The handoffs behind the milestone",
        rows: [
          { label: "Engineering", value: "API review · Friday" },
          { label: "Marketing", value: "Launch copy · In review" },
        ],
        note: "One launch, connected teams",
      },
    },
    {
      kind: "planning",
      visual: {
        heading: "Before you commit",
        subheading: "Make capacity part of the decision",
        rows: [
          { label: "Engineering", value: "79%", width: "79%" },
          { label: "Support", value: "63%", width: "63%" },
        ],
        note: "Make tradeoffs with the work in view",
      },
    },
  ],
  integrations: [
    {
      kind: "feedback",
      brand: "Slack",
      visual: {
        heading: "Conversation to work",
        subheading: "Keep the original request attached",
        badge: "Created",
        rows: [
          { label: "Source", value: "#customer-launch" },
          { label: "Task", value: "Confirm onboarding timeline" },
        ],
        note: "The conversation stays connected",
      },
    },
    {
      kind: "work",
      brand: "GitHub",
      visual: {
        heading: "Code meets the plan",
        subheading: "Delivery context on the task",
        rows: [
          { label: "ENG-214", value: "Improve search" },
          { label: "Pull request", value: "Ready for review · 1 approval" },
        ],
        note: "Issues, branches, and pull requests",
      },
    },
    {
      kind: "planning",
      brand: "Google Calendar",
      visual: {
        heading: "Time to focus",
        subheading: "Meetings beside planned work",
        rows: [
          { label: "09:00", value: "Product review" },
          { label: "11:00", value: "Focus block · OPS-42" },
        ],
        note: "Availability before another commitment",
      },
    },
  ],
};

export function getFeatureIllustrations(slug: string) {
  return FEATURE_ILLUSTRATIONS[slug];
}
