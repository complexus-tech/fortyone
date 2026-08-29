import type { ImageProps } from "next/image";
import calendarWeekDark from "../../../public/images/product/calendar-week-dark.webp";
import calendarWeekLight from "../../../public/images/product/calendar-week-light.webp";
import documentsRelatedWorkDark from "../../../public/images/product/documents-related-work-dark.webp";
import documentsRelatedWorkLight from "../../../public/images/product/documents-related-work-light.webp";
import feedbackPortalDark from "../../../public/images/product/feedback-portal-dark.webp";
import feedbackPortalLight from "../../../public/images/product/feedback-portal-light.webp";
import mayaHomeDark from "../../../public/images/product/maya-home-dark.webp";
import mayaHomeLight from "../../../public/images/product/maya-home-light.webp";
import mayaObjectiveRisksDark from "../../../public/images/product/maya-objective-risks-dark.webp";
import mayaObjectiveRisksLight from "../../../public/images/product/maya-objective-risks-light.webp";
import myWorkBoardDark from "../../../public/images/product/my-work-board-dark.webp";
import myWorkBoardLight from "../../../public/images/product/my-work-board-light.webp";
import myWorkListDark from "../../../public/images/product/my-work-list-dark.webp";
import myWorkListLight from "../../../public/images/product/my-work-list-light.webp";
import publicFeedbackRoadmapDark from "../../../public/images/product/public-feedback-roadmap-dark.webp";
import publicFeedbackRoadmapLight from "../../../public/images/product/public-feedback-roadmap-light.webp";
import roadmapTimelineDark from "../../../public/images/product/roadmap-timeline-dark.webp";
import roadmapTimelineLight from "../../../public/images/product/roadmap-timeline-light.webp";
import strategyMapDark from "../../../public/images/product/strategy-map-dark.webp";
import strategyMapLight from "../../../public/images/product/strategy-map-light.webp";
import type { FeatureWorkflowItem } from "./feature-product-workflow";

export type FeatureLandingConfig = {
  ctaDescription: string;
  ctaTitle: string;
  decisionCards: readonly {
    description: string;
    title: string;
  }[];
  decisionDescription: string;
  decisionHeading: string;
  faqHeading: string;
  hero: {
    alt: string;
    darkImage: ImageProps["src"];
    description: string;
    lightImage: ImageProps["src"];
    title: string;
    url: string;
  };
  keywords: string[];
  workflow: {
    ariaLabel: string;
    description: string;
    heading: string;
    id: string;
    items: readonly FeatureWorkflowItem[];
  };
};

const FEATURE_PAGE_CONFIGS = {
  "customer-feedback": {
    ctaDescription:
      "Start free and give customers one clear place to share requests and follow progress.",
    ctaTitle: "Turn the next customer signal into progress.",
    decisionCards: [
      {
        description:
          "Customers can submit requests, vote on ideas, and follow progress from one public workspace.",
        title: "One feedback portal",
      },
      {
        description:
          "Accepted feedback becomes planned work while the original request remains attached.",
        title: "Connected delivery",
      },
      {
        description:
          "Compare related requests, customer evidence, and product priorities before committing the team.",
        title: "Clear prioritization",
      },
    ],
    decisionDescription:
      "Bring demand, product priorities, and delivery context together before the team commits to a request.",
    decisionHeading: "Decide what customer feedback should change.",
    faqHeading: "Frequently asked questions about customer feedback",
    hero: {
      alt: "FortyOne public feedback roadmap showing planned, active, and completed customer priorities",
      darkImage: publicFeedbackRoadmapDark,
      description:
        "Collect requests, understand demand, turn accepted ideas into owned work, and show customers what changed.",
      lightImage: publicFeedbackRoadmapLight,
      title: "Turn customer feedback into progress people can follow.",
      url: "https://fortyone.app/feedback/roadmap",
    },
    keywords: [
      "customer feedback management",
      "feedback portal",
      "feature request software",
      "customer feedback roadmap",
      "product feedback management",
      "public product roadmap",
    ],
    workflow: {
      ariaLabel: "Explore the customer feedback workflow",
      description:
        "Keep the original request connected as the team moves from customer demand to a delivery decision.",
      heading: "Follow the path from customer request to visible progress.",
      id: "customer-feedback-workflow",
      items: [
        {
          alt: "FortyOne customer feedback portal with requests, votes, and customer context",
          darkImage: feedbackPortalDark,
          description:
            "Give customers one place to submit requests, add context, vote, and find related ideas.",
          label: "Collect requests",
          lightImage: feedbackPortalLight,
          title: "Hear what customers need without rebuilding the signal.",
          value: "collect-requests",
          url: "https://fortyone.app/feedback",
        },
        {
          alt: "FortyOne My Work board showing customer-backed priorities moving through delivery",
          darkImage: myWorkBoardDark,
          description:
            "Move accepted requests into owned work while keeping the customer evidence attached to the decision.",
          label: "Plan the work",
          lightImage: myWorkBoardLight,
          title: "Turn the strongest feedback into accountable work.",
          value: "plan-the-work",
          url: "https://fortyone.app/my-work",
        },
        {
          alt: "FortyOne public feedback roadmap showing planned, active, and completed customer priorities",
          darkImage: publicFeedbackRoadmapDark,
          description:
            "Publish planned, active, and completed work without exposing the internal planning workspace.",
          label: "Share progress",
          lightImage: publicFeedbackRoadmapLight,
          title:
            "Show customers what is moving without exposing internal work.",
          value: "share-progress",
          url: "https://fortyone.app/feedback/roadmap",
        },
      ],
    },
  },
  goals: {
    ctaDescription:
      "Start free and connect the first objective to the work that moves it forward.",
    ctaTitle: "Give every important task a reason.",
    decisionCards: [
      {
        description:
          "Compare tasks against the goals they support instead of debating work in isolation.",
        title: "Clearer prioritization",
      },
      {
        description:
          "Prepare owners, timing, and next steps with the outcome behind the work included.",
        title: "Stronger AI context",
      },
      {
        description:
          "Start reviews from live goal-linked work instead of copied updates and disconnected notes.",
        title: "Less status reconstruction",
      },
    ],
    decisionDescription:
      "Use live work, ownership, and risk to understand whether an outcome is actually moving.",
    decisionHeading: "Keep every planning decision tied to an outcome.",
    faqHeading: "Frequently asked questions about goals and OKRs",
    hero: {
      alt: "FortyOne Strategy Map connecting company goals to strategic pillars and measurable objectives",
      darkImage: strategyMapDark,
      description:
        "Connect company direction to measurable objectives, owned work, and the decisions that keep progress moving.",
      lightImage: strategyMapLight,
      title: "Turn company goals into work teams can follow.",
      url: "https://fortyone.app/strategy-map",
    },
    keywords: [
      "goal management software",
      "OKR software",
      "strategy execution software",
      "goal tracking",
      "connect goals to tasks",
      "team objectives software",
    ],
    workflow: {
      ariaLabel: "Explore the goals workflow",
      description:
        "Carry the outcome from company direction into the roadmap, daily work, and the next leadership decision.",
      heading: "Follow the path from goal to measurable progress.",
      id: "goals-workflow",
      items: [
        {
          alt: "FortyOne roadmap timeline sequencing objectives that support company goals",
          darkImage: roadmapTimelineDark,
          description:
            "Translate strategic direction into objectives with owners, dates, and a place in the delivery sequence.",
          label: "Sequence objectives",
          lightImage: roadmapTimelineLight,
          title: "Give every outcome a realistic place in the plan.",
          value: "sequence-objectives",
          url: "https://fortyone.app/roadmap",
        },
        {
          alt: "FortyOne My Work board showing tasks connected to strategic outcomes",
          darkImage: myWorkBoardDark,
          description:
            "Keep the outcome close to assignments, estimates, status, and the context behind the work.",
          label: "Connect the work",
          lightImage: myWorkBoardLight,
          title: "Make daily work traceable to the goal it supports.",
          value: "connect-the-work",
          url: "https://fortyone.app/my-work",
        },
        {
          alt: "Maya highlighting at-risk objectives and the decisions that need leadership attention",
          darkImage: mayaObjectiveRisksDark,
          description:
            "Bring progress, blockers, owners, and risk into the review without reconstructing status by hand.",
          label: "Review progress",
          lightImage: mayaObjectiveRisksLight,
          title: "See which goals need a decision while there is time to act.",
          value: "review-progress",
          url: "https://fortyone.app/summary",
        },
      ],
    },
  },
  tasks: {
    ctaDescription:
      "Start free and bring the next request into a task with its context still attached.",
    ctaTitle: "Move the next task with clarity.",
    decisionCards: [
      {
        description:
          "Turn rough requests into reviewable tasks with the source and missing details still visible.",
        title: "Cleaner intake",
      },
      {
        description:
          "Recommend an owner and work window from real workload, availability, and related context.",
        title: "Better ownership",
      },
      {
        description:
          "Compare team load before assignment so the next task does not create a hidden bottleneck.",
        title: "Plan around capacity",
      },
    ],
    decisionDescription:
      "Keep intake, ownership, source context, and status together so work can move without another reconstruction step.",
    decisionHeading: "Give every task the context it needs to move.",
    faqHeading: "Frequently asked questions about task management",
    hero: {
      alt: "FortyOne My Work board showing prioritized tasks across a delivery workflow",
      darkImage: myWorkBoardDark,
      description:
        "Turn requests into structured work with an owner, goal, estimate, source context, and clear delivery status.",
      lightImage: myWorkBoardLight,
      title: "Keep every task connected to the plan.",
      url: "https://fortyone.app/my-work",
    },
    keywords: [
      "task management software",
      "AI task management",
      "team task tracking",
      "task intake workflow",
      "project task management",
      "connect tasks to goals",
    ],
    workflow: {
      ariaLabel: "Explore the task management workflow",
      description:
        "Move from a clear intake queue to scheduled work without losing the request, decision, or related context.",
      heading: "Follow the path from request to completed work.",
      id: "tasks-workflow",
      items: [
        {
          alt: "FortyOne My Work list showing tasks, owners, priority, and status",
          darkImage: myWorkListDark,
          description:
            "Turn rough requests into reviewable tasks with enough structure for the team to act.",
          label: "Structure intake",
          lightImage: myWorkListLight,
          title: "Start with a task the team can understand.",
          value: "structure-intake",
          url: "https://fortyone.app/my-work",
        },
        {
          alt: "FortyOne weekly calendar placing assigned tasks around existing commitments",
          darkImage: calendarWeekDark,
          description:
            "Use ownership, effort, workload, and availability to place the work in a realistic window.",
          label: "Plan the timing",
          lightImage: calendarWeekLight,
          title: "Put assigned work where it can actually happen.",
          value: "plan-the-timing",
          url: "https://fortyone.app/calendar",
        },
        {
          alt: "FortyOne document with related tasks and objectives attached to the source context",
          darkImage: documentsRelatedWorkDark,
          description:
            "Keep documents, goals, and related work attached so the next owner does not have to rebuild the brief.",
          label: "Keep the context",
          lightImage: documentsRelatedWorkLight,
          title: "Let the reason for the task travel with the work.",
          value: "keep-the-context",
          url: "https://fortyone.app/documents",
        },
      ],
    },
  },
  roadmaps: {
    ctaDescription:
      "Start free and connect the next roadmap priority to the work required to deliver it.",
    ctaTitle: "Build a roadmap grounded in execution.",
    decisionCards: [
      {
        description:
          "Keep roadmap items connected to the tasks, owners, estimates, and blockers behind them.",
        title: "Execution visibility",
      },
      {
        description:
          "Track cross-functional launch work from planning through delivery in one view.",
        title: "Cleaner launch planning",
      },
      {
        description:
          "Compare priority, capacity, and timing before adding another commitment to the roadmap.",
        title: "Better tradeoff decisions",
      },
    ],
    decisionDescription:
      "Bring priority, capacity, cross-functional work, and customer visibility together before making another promise.",
    decisionHeading: "Build a roadmap that stays connected to delivery.",
    faqHeading: "Frequently asked questions about product roadmaps",
    hero: {
      alt: "FortyOne roadmap timeline connecting priorities, owners, dates, and delivery status",
      darkImage: roadmapTimelineDark,
      description:
        "Connect priorities to goals, tasks, owners, capacity, and customer-facing progress in one delivery system.",
      lightImage: roadmapTimelineLight,
      title: "Build roadmaps your team can actually deliver.",
      url: "https://fortyone.app/roadmap",
    },
    keywords: [
      "product roadmap software",
      "roadmap planning tool",
      "project roadmap software",
      "public product roadmap",
      "roadmap capacity planning",
      "connect roadmap to execution",
    ],
    workflow: {
      ariaLabel: "Explore the roadmap workflow",
      description:
        "Carry strategic direction into sequenced work, then share the right level of progress with customers.",
      heading: "Follow the path from priority to delivered roadmap.",
      id: "roadmaps-workflow",
      items: [
        {
          alt: "FortyOne Strategy Map connecting roadmap priorities to company direction",
          darkImage: strategyMapDark,
          description:
            "Start with the outcome and strategic pillar the roadmap priority is meant to support.",
          label: "Set direction",
          lightImage: strategyMapLight,
          title: "Give every roadmap priority a reason to exist.",
          value: "set-direction",
          url: "https://fortyone.app/strategy-map",
        },
        {
          alt: "FortyOne My Work board showing roadmap priorities moving through execution",
          darkImage: myWorkBoardDark,
          description:
            "Connect each priority to the tasks, owners, estimates, and blockers required to deliver it.",
          label: "Connect delivery",
          lightImage: myWorkBoardLight,
          title: "Keep the roadmap tied to the work behind it.",
          value: "connect-delivery",
          url: "https://fortyone.app/my-work",
        },
        {
          alt: "FortyOne public roadmap sharing planned, active, and completed priorities with customers",
          darkImage: publicFeedbackRoadmapDark,
          description:
            "Publish planned, active, and completed work without exposing internal planning details.",
          label: "Share progress",
          lightImage: publicFeedbackRoadmapLight,
          title: "Show customers what is moving and what changed.",
          value: "share-progress",
          url: "https://fortyone.app/feedback/roadmap",
        },
      ],
    },
  },
  integrations: {
    ctaDescription:
      "Start free and connect the tools where your team already discusses, builds, and schedules work.",
    ctaTitle: "Keep your tools connected to the plan.",
    decisionCards: [
      {
        description:
          "Keep source context attached to tasks instead of rewriting it into another status update.",
        title: "Less manual copying",
      },
      {
        description:
          "Keep GitHub issues, pull requests, and commits linked to the work they move forward.",
        title: "Connected delivery context",
      },
      {
        description:
          "Use calendar availability and current workload to place work around existing commitments.",
        title: "More realistic scheduling",
      },
    ],
    decisionDescription:
      "Turn conversations, delivery signals, calendar availability, and AI access into usable project context.",
    decisionHeading: "Keep every tool useful inside the plan.",
    faqHeading: "Frequently asked questions about FortyOne integrations",
    hero: {
      alt: "FortyOne weekly calendar combining scheduled project work with connected calendar commitments",
      darkImage: calendarWeekDark,
      description:
        "Connect Slack, GitHub, calendars, and MCP-compatible AI clients to the work your team plans and delivers.",
      lightImage: calendarWeekLight,
      title: "Keep project context connected across every tool.",
      url: "https://fortyone.app/calendar",
    },
    keywords: [
      "project management integrations",
      "Slack project management integration",
      "GitHub project management integration",
      "Google Calendar task planning",
      "Outlook Calendar project planning",
      "MCP project management server",
    ],
    workflow: {
      ariaLabel: "Explore the integrations workflow",
      description:
        "Keep the source attached as conversations, documents, AI clients, and delivery work move through FortyOne.",
      heading: "Follow the path from external context to planned work.",
      id: "integrations-workflow",
      items: [
        {
          alt: "FortyOne document showing related work attached to its original project context",
          darkImage: documentsRelatedWorkDark,
          description:
            "Bring source material into the plan so owners can understand the request without chasing another tool.",
          label: "Bring in context",
          lightImage: documentsRelatedWorkLight,
          title: "Keep the source connected to the work it creates.",
          value: "bring-in-context",
          url: "https://fortyone.app/documents",
        },
        {
          alt: "Maya ready to use connected FortyOne project context from an MCP-compatible AI client",
          darkImage: mayaHomeDark,
          description:
            "Use FortyOne from ChatGPT, Claude, Cursor, Codex, and other MCP-compatible clients with permission-aware access.",
          label: "Work with AI",
          lightImage: mayaHomeLight,
          title: "Bring governed project context into the AI tools you use.",
          value: "work-with-ai",
          url: "https://fortyone.app/maya",
        },
        {
          alt: "FortyOne My Work list showing structured tasks created from connected source context",
          darkImage: myWorkListDark,
          description:
            "Turn the external request into structured, owned work with the original source still attached.",
          label: "Move the work",
          lightImage: myWorkListLight,
          title: "Let connected context become work the team can track.",
          value: "move-the-work",
          url: "https://fortyone.app/my-work",
        },
      ],
    },
  },
} as const satisfies Record<string, FeatureLandingConfig>;

export function getFeaturePageConfig(slug: string) {
  return FEATURE_PAGE_CONFIGS[slug as keyof typeof FEATURE_PAGE_CONFIGS];
}
