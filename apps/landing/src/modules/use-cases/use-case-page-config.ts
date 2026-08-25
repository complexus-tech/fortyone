import type { ImageProps } from "next/image";
import calendarWeekDark from "../../../public/images/product/calendar-week-dark.webp";
import calendarWeekLight from "../../../public/images/product/calendar-week-light.webp";
import documentsRelatedWorkDark from "../../../public/images/product/documents-related-work-dark.webp";
import documentsRelatedWorkLight from "../../../public/images/product/documents-related-work-light.webp";
import feedbackPortalDark from "../../../public/images/product/feedback-portal-dark.webp";
import feedbackPortalLight from "../../../public/images/product/feedback-portal-light.webp";
import mayaDeliveryBriefDark from "../../../public/images/product/maya-delivery-brief-dark.webp";
import mayaDeliveryBriefLight from "../../../public/images/product/maya-delivery-brief-light.webp";
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

type UseCaseProductVisual = {
  alt: string;
  darkImage: ImageProps["src"];
  label: string;
  lightImage: ImageProps["src"];
  url: string;
  value: string;
};

export type UseCasePageConfig = {
  ctaDescription: string;
  ctaTitle: string;
  decisionDescription: string;
  decisionHeading: string;
  hero: {
    alt: string;
    darkImage: ImageProps["src"];
    description: string;
    lightImage: ImageProps["src"];
    title: string;
    url: string;
  };
  keywords: readonly string[];
  workflow: {
    description: string;
    heading: string;
    visuals: readonly UseCaseProductVisual[];
  };
};

const USE_CASE_PAGE_CONFIGS = {
  operations: {
    ctaDescription:
      "Start free and give every operational request a clear owner, place, and next decision.",
    ctaTitle: "Keep operations moving without rebuilding the plan.",
    decisionDescription:
      "Give operations teams the live context they need to prioritize requests, balance ownership, and act before work stalls.",
    decisionHeading:
      "Make the next operational decision with the work in view.",
    hero: {
      alt: "Maya preparing an operations delivery brief from live FortyOne project work",
      darkImage: mayaDeliveryBriefDark,
      description:
        "Turn recurring requests, handoffs, meetings, and delivery risks into owned work that teams can follow.",
      lightImage: mayaDeliveryBriefLight,
      title: "Turn operational requests into work teams can actually deliver.",
      url: "https://fortyone.app/maya",
    },
    keywords: [
      "operations project management software",
      "operations workflow management",
      "AI operations planning",
      "cross-functional operations",
    ],
    workflow: {
      description:
        "Carry the request from intake into ownership, timing, and the review that keeps execution moving.",
      heading:
        "Follow the work from operational request to accountable delivery.",
      visuals: [
        {
          alt: "Maya summarizing operational priorities and blocked work",
          darkImage: mayaDeliveryBriefDark,
          label: "See the operating picture",
          lightImage: mayaDeliveryBriefLight,
          url: "https://fortyone.app/maya",
          value: "see-the-picture",
        },
        {
          alt: "FortyOne My Work list showing operational tasks and owners",
          darkImage: myWorkListDark,
          label: "Assign the work",
          lightImage: myWorkListLight,
          url: "https://fortyone.app/my-work",
          value: "assign-the-work",
        },
        {
          alt: "FortyOne calendar planning operational work around existing commitments",
          darkImage: calendarWeekDark,
          label: "Plan the timing",
          lightImage: calendarWeekLight,
          url: "https://fortyone.app/calendar",
          value: "plan-the-timing",
        },
      ],
    },
  },
  product: {
    ctaDescription:
      "Start free and connect the next customer signal to the roadmap and work behind it.",
    ctaTitle: "Give product decisions a traceable path into delivery.",
    decisionDescription:
      "Bring customer evidence, goals, roadmap tradeoffs, and delivery reality together before making another promise.",
    decisionHeading: "Decide what the product should change next.",
    hero: {
      alt: "FortyOne public feedback roadmap showing customer priorities moving through delivery",
      darkImage: publicFeedbackRoadmapDark,
      description:
        "Connect customer feedback, product priorities, roadmap decisions, and delivery work in one plan.",
      lightImage: publicFeedbackRoadmapLight,
      title: "Turn customer signals into a roadmap teams can deliver.",
      url: "https://fortyone.app/feedback/roadmap",
    },
    keywords: [
      "product management software",
      "product planning platform",
      "customer feedback to roadmap",
      "AI product management",
    ],
    workflow: {
      description:
        "Keep the customer need attached as the team moves from evidence to priority, delivery, and visible progress.",
      heading: "Follow the path from customer signal to product progress.",
      visuals: [
        {
          alt: "FortyOne feedback portal collecting customer requests and votes",
          darkImage: feedbackPortalDark,
          label: "Collect the signal",
          lightImage: feedbackPortalLight,
          url: "https://fortyone.app/feedback",
          value: "collect-the-signal",
        },
        {
          alt: "FortyOne roadmap timeline sequencing product priorities",
          darkImage: roadmapTimelineDark,
          label: "Sequence priorities",
          lightImage: roadmapTimelineLight,
          url: "https://fortyone.app/roadmap",
          value: "sequence-priorities",
        },
        {
          alt: "FortyOne My Work board connecting product priorities to delivery",
          darkImage: myWorkBoardDark,
          label: "Connect delivery",
          lightImage: myWorkBoardLight,
          url: "https://fortyone.app/my-work",
          value: "connect-delivery",
        },
      ],
    },
  },
  developers: {
    ctaDescription:
      "Start free and keep the next engineering request connected to its source, owner, and delivery status.",
    ctaTitle: "Keep engineering context close to the work.",
    decisionDescription:
      "Give engineering teams enough source context, priority, ownership, and capacity information to make a realistic commitment.",
    decisionHeading:
      "Move engineering work without losing the reason behind it.",
    hero: {
      alt: "FortyOne My Work board showing engineering priorities moving through delivery",
      darkImage: myWorkBoardDark,
      description:
        "Keep GitHub context, technical decisions, estimates, owners, and project priorities connected from request to release.",
      lightImage: myWorkBoardLight,
      title: "Keep engineering context connected from request to release.",
      url: "https://fortyone.app/my-work",
    },
    keywords: [
      "engineering project management",
      "developer task management",
      "GitHub project management integration",
      "AI engineering planning",
    ],
    workflow: {
      description:
        "Carry the technical source into a scoped task, a realistic plan, and the review that protects delivery.",
      heading: "Follow the work from technical context to shipped change.",
      visuals: [
        {
          alt: "FortyOne document keeping technical context linked to related work",
          darkImage: documentsRelatedWorkDark,
          label: "Keep the source",
          lightImage: documentsRelatedWorkLight,
          url: "https://fortyone.app/documents",
          value: "keep-the-source",
        },
        {
          alt: "FortyOne My Work list showing scoped engineering tasks",
          darkImage: myWorkListDark,
          label: "Scope the work",
          lightImage: myWorkListLight,
          url: "https://fortyone.app/my-work",
          value: "scope-the-work",
        },
        {
          alt: "Maya highlighting delivery risks that need an engineering decision",
          darkImage: mayaObjectiveRisksDark,
          label: "Review delivery risk",
          lightImage: mayaObjectiveRisksLight,
          url: "https://fortyone.app/maya",
          value: "review-delivery-risk",
        },
      ],
    },
  },
  "customer-support": {
    ctaDescription:
      "Start free and give the next recurring customer issue a visible path into product work.",
    ctaTitle: "Close the loop between support and delivery.",
    decisionDescription:
      "Help support teams connect recurring issues to product work, owners, and customer-facing progress.",
    decisionHeading:
      "Turn support patterns into decisions the product team can act on.",
    hero: {
      alt: "FortyOne feedback portal organizing customer needs for review",
      darkImage: feedbackPortalDark,
      description:
        "Capture recurring customer needs, connect accepted requests to delivery, and show customers what changed.",
      lightImage: feedbackPortalLight,
      title: "Turn recurring customer issues into visible product work.",
      url: "https://fortyone.app/feedback",
    },
    keywords: [
      "customer support project management",
      "support feedback management",
      "support to product workflow",
      "customer request tracking",
    ],
    workflow: {
      description:
        "Keep the customer need intact as support identifies a pattern, product accepts the work, and progress becomes visible.",
      heading: "Follow the path from support signal to customer update.",
      visuals: [
        {
          alt: "FortyOne feedback portal grouping recurring customer issues",
          darkImage: feedbackPortalDark,
          label: "Group recurring needs",
          lightImage: feedbackPortalLight,
          url: "https://fortyone.app/feedback",
          value: "group-recurring-needs",
        },
        {
          alt: "FortyOne My Work board showing accepted support requests in delivery",
          darkImage: myWorkBoardDark,
          label: "Connect product work",
          lightImage: myWorkBoardLight,
          url: "https://fortyone.app/my-work",
          value: "connect-product-work",
        },
        {
          alt: "FortyOne public roadmap sharing product progress with customers",
          darkImage: publicFeedbackRoadmapDark,
          label: "Share the outcome",
          lightImage: publicFeedbackRoadmapLight,
          url: "https://fortyone.app/feedback/roadmap",
          value: "share-the-outcome",
        },
      ],
    },
  },
  "field-crews": {
    ctaDescription:
      "Start free and connect the next site request to the office plan, owner, and timing behind it.",
    ctaTitle: "Keep office plans and field work moving together.",
    decisionDescription:
      "Give field and office teams one view of assignments, dependencies, timing, and the decisions blocking site progress.",
    decisionHeading: "Make field decisions with the latest plan in view.",
    hero: {
      alt: "FortyOne calendar planning field assignments around current commitments",
      darkImage: calendarWeekDark,
      description:
        "Connect site requests, office coordination, assignments, dependencies, and schedules in one delivery plan.",
      lightImage: calendarWeekLight,
      title: "Keep field work moving from office plan to site handoff.",
      url: "https://fortyone.app/calendar",
    },
    keywords: [
      "field service project management",
      "field crew scheduling",
      "construction task management",
      "site operations planning",
    ],
    workflow: {
      description:
        "Move the request into an owned task, a realistic work window, and a clear delivery brief for the next handoff.",
      heading: "Follow the work from site request to completed handoff.",
      visuals: [
        {
          alt: "FortyOne My Work list showing field tasks and owners",
          darkImage: myWorkListDark,
          label: "Structure the request",
          lightImage: myWorkListLight,
          url: "https://fortyone.app/my-work",
          value: "structure-the-request",
        },
        {
          alt: "FortyOne calendar scheduling field work around commitments",
          darkImage: calendarWeekDark,
          label: "Plan the site window",
          lightImage: calendarWeekLight,
          url: "https://fortyone.app/calendar",
          value: "plan-the-site-window",
        },
        {
          alt: "Maya preparing a delivery brief for field coordination",
          darkImage: mayaDeliveryBriefDark,
          label: "Prepare the handoff",
          lightImage: mayaDeliveryBriefLight,
          url: "https://fortyone.app/maya",
          value: "prepare-the-handoff",
        },
      ],
    },
  },
  government: {
    ctaDescription:
      "Start with a conversation about the deployment, governance, and workflow requirements behind your program.",
    ctaTitle: "Connect public priorities to accountable delivery.",
    decisionDescription:
      "Give leaders and delivery teams a traceable view of program outcomes, ownership, risk, and the work behind each update.",
    decisionHeading:
      "Keep public programs connected to the work that delivers them.",
    hero: {
      alt: "FortyOne Strategy Map connecting government priorities to programs and measurable outcomes",
      darkImage: strategyMapDark,
      description:
        "Connect policy priorities, programs, department handoffs, delivery work, and reporting in a governed workspace.",
      lightImage: strategyMapLight,
      title: "Connect public priorities to accountable delivery.",
      url: "https://fortyone.app/strategy-map",
    },
    keywords: [
      "government project management software",
      "public sector program management",
      "government strategy execution",
      "sovereign project management",
    ],
    workflow: {
      description:
        "Carry the public outcome into program sequencing, department ownership, and the evidence behind the next report.",
      heading: "Follow the path from public priority to program delivery.",
      visuals: [
        {
          alt: "FortyOne Strategy Map connecting public priorities to outcomes",
          darkImage: strategyMapDark,
          label: "Set program direction",
          lightImage: strategyMapLight,
          url: "https://fortyone.app/strategy-map",
          value: "set-program-direction",
        },
        {
          alt: "FortyOne roadmap timeline sequencing cross-department programs",
          darkImage: roadmapTimelineDark,
          label: "Coordinate delivery",
          lightImage: roadmapTimelineLight,
          url: "https://fortyone.app/roadmap",
          value: "coordinate-delivery",
        },
        {
          alt: "FortyOne documents keeping program evidence linked to related work",
          darkImage: documentsRelatedWorkDark,
          label: "Keep the evidence",
          lightImage: documentsRelatedWorkLight,
          url: "https://fortyone.app/documents",
          value: "keep-the-evidence",
        },
      ],
    },
  },
  marketing: {
    ctaDescription:
      "Start free and connect the next campaign brief to its owners, dependencies, review steps, and launch date.",
    ctaTitle: "Give every campaign a delivery plan the team can follow.",
    decisionDescription:
      "Bring briefs, approvals, ownership, workload, and launch dependencies together before committing to another date.",
    decisionHeading: "Plan campaigns around the work required to launch them.",
    hero: {
      alt: "FortyOne My Work board showing campaign work moving across teams",
      darkImage: myWorkBoardDark,
      description:
        "Connect briefs, creative work, approvals, owners, dependencies, and launch timing in one campaign plan.",
      lightImage: myWorkBoardLight,
      title: "Plan campaigns around real owners, dependencies, and capacity.",
      url: "https://fortyone.app/my-work",
    },
    keywords: [
      "marketing project management software",
      "campaign planning software",
      "marketing workflow management",
      "creative project planning",
    ],
    workflow: {
      description:
        "Keep the brief connected as work moves into ownership, review, timing, and launch readiness.",
      heading: "Follow the campaign from brief to coordinated launch.",
      visuals: [
        {
          alt: "FortyOne document keeping a campaign brief linked to related work",
          darkImage: documentsRelatedWorkDark,
          label: "Keep the brief",
          lightImage: documentsRelatedWorkLight,
          url: "https://fortyone.app/documents",
          value: "keep-the-brief",
        },
        {
          alt: "FortyOne My Work board coordinating campaign assignments",
          darkImage: myWorkBoardDark,
          label: "Coordinate production",
          lightImage: myWorkBoardLight,
          url: "https://fortyone.app/my-work",
          value: "coordinate-production",
        },
        {
          alt: "FortyOne calendar placing campaign work around existing commitments",
          darkImage: calendarWeekDark,
          label: "Protect the launch",
          lightImage: calendarWeekLight,
          url: "https://fortyone.app/calendar",
          value: "protect-the-launch",
        },
      ],
    },
  },
  leadership: {
    ctaDescription:
      "Start free and connect the next leadership priority to the work, capacity, and decisions behind it.",
    ctaTitle: "See where strategy needs leadership attention.",
    decisionDescription:
      "Give leaders a current view of priority, progress, ownership, capacity, and risk without another status reconstruction exercise.",
    decisionHeading:
      "Make leadership decisions from live strategy and execution.",
    hero: {
      alt: "Maya highlighting at-risk objectives and leadership decisions in FortyOne",
      darkImage: mayaObjectiveRisksDark,
      description:
        "Connect company priorities to live work, owners, progress, capacity, and the decisions that keep outcomes moving.",
      lightImage: mayaObjectiveRisksLight,
      title: "See where strategy is moving—and where it needs a decision.",
      url: "https://fortyone.app/maya",
    },
    keywords: [
      "executive project dashboard",
      "strategy execution software",
      "leadership planning software",
      "AI executive project management",
    ],
    workflow: {
      description:
        "Carry company direction into sequenced work, then review progress and risk from the same live plan.",
      heading: "Follow the path from company priority to leadership decision.",
      visuals: [
        {
          alt: "FortyOne Strategy Map connecting leadership priorities to objectives",
          darkImage: strategyMapDark,
          label: "Set direction",
          lightImage: strategyMapLight,
          url: "https://fortyone.app/strategy-map",
          value: "set-direction",
        },
        {
          alt: "FortyOne roadmap timeline sequencing company priorities",
          darkImage: roadmapTimelineDark,
          label: "Connect execution",
          lightImage: roadmapTimelineLight,
          url: "https://fortyone.app/roadmap",
          value: "connect-execution",
        },
        {
          alt: "Maya highlighting objective risks that need a leadership decision",
          darkImage: mayaObjectiveRisksDark,
          label: "Review what changed",
          lightImage: mayaObjectiveRisksLight,
          url: "https://fortyone.app/maya",
          value: "review-what-changed",
        },
      ],
    },
  },
} as const satisfies Record<string, UseCasePageConfig>;

export function getUseCasePageConfig(slug: string) {
  return USE_CASE_PAGE_CONFIGS[slug as keyof typeof USE_CASE_PAGE_CONFIGS];
}
