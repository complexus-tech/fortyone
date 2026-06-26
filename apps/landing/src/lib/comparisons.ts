import type { MarketingDetail } from "@/components/shared/marketing-detail-page";

export type ComparisonRow = {
  criteria: string;
  fortyOne: string;
  competitor: string;
};

export type Comparison = {
  bestFor: string[];
  competitor: string;
  competitorStrengths: string[];
  description: string;
  faqs: [string, string][];
  metaDescription: string;
  metaTitle: string;
  rows: ComparisonRow[];
  slug: string;
  summary: string;
  title: string;
  tradeoffs: string[];
};

export const comparisons: Comparison[] = [
  {
    slug: "trello",
    competitor: "Trello",
    title: "FortyOne vs Trello",
    summary:
      "Trello is simple and visual for lightweight boards. FortyOne is built for teams that want tasks, goals, AI planning, workload, and review controls in one execution system.",
    description:
      "Compare FortyOne and Trello for teams deciding between lightweight boards and AI-assisted project management.",
    metaTitle: "FortyOne vs Trello | AI Project Management Comparison",
    metaDescription:
      "Compare FortyOne and Trello for goals, task planning, AI owner suggestions, estimates, roadmaps, and delivery visibility.",
    bestFor: [
      "Goal-connected project planning",
      "AI owner, estimate, and timing suggestions",
      "Review-before-apply AI workflows",
      "Cross-functional roadmap execution",
    ],
    competitorStrengths: [
      "Simple visual Kanban boards",
      "Easy personal and small-team task tracking",
      "Low setup for lightweight workflows",
    ],
    tradeoffs: [
      "Choose Trello when your core need is a simple visual board for lightweight task tracking.",
      "Choose FortyOne when planning needs goals, capacity, AI recommendations, and manager review controls.",
    ],
    rows: [
      {
        criteria: "Primary workflow",
        fortyOne:
          "Goals, tasks, AI planning, workload, and review in one place",
        competitor: "Visual Kanban boards and lightweight task tracking",
      },
      {
        criteria: "AI planning",
        fortyOne:
          "Suggests owners, estimates, timing, and risk notes for review",
        competitor:
          "Automation can help boards, but planning remains mostly manual",
      },
      {
        criteria: "Goal connection",
        fortyOne: "Keeps tasks and roadmap work tied to outcomes",
        competitor:
          "Cards can be organized visually, but goals are not the planning center",
      },
      {
        criteria: "Best fit",
        fortyOne: "Teams that want AI-assisted delivery coordination",
        competitor: "Teams that want a simple board for known work",
      },
    ],
    faqs: [
      [
        "Is FortyOne a Trello replacement?",
        "It can replace Trello for teams that have outgrown simple boards and need project planning, goals, tasks, AI recommendations, and review controls in one workflow.",
      ],
      [
        "When should a team keep Trello?",
        "If the team only needs a simple visual board and already handles planning elsewhere, Trello may remain a better fit.",
      ],
    ],
  },
  {
    slug: "asana",
    competitor: "Asana",
    title: "FortyOne vs Asana",
    summary:
      "Asana is a broad work management platform. FortyOne focuses on AI-assisted planning where goals, task intake, owners, estimates, and review controls stay connected.",
    description:
      "Compare FortyOne and Asana for teams evaluating AI planning, goal-connected tasks, and project execution.",
    metaTitle: "FortyOne vs Asana | AI Project Management Comparison",
    metaDescription:
      "Compare FortyOne and Asana for goal tracking, AI planning, task management, workload, and project execution.",
    bestFor: [
      "AI-assisted task intake",
      "Goal-connected execution",
      "Owner and schedule recommendations",
      "Lightweight project review workflows",
    ],
    competitorStrengths: [
      "Broad work management",
      "Many views and templates",
      "Large ecosystem and enterprise familiarity",
    ],
    tradeoffs: [
      "Choose Asana when you need a mature, broad work-management suite.",
      "Choose FortyOne when AI planning and goal-linked execution are the product center.",
    ],
    rows: [
      {
        criteria: "Primary workflow",
        fortyOne: "AI-assisted planning and goal-connected task execution",
        competitor: "General work management across many team types",
      },
      {
        criteria: "Task intake",
        fortyOne: "Turns rough requests into structured tasks with context",
        competitor: "Strong manual task and project organization",
      },
      {
        criteria: "AI control",
        fortyOne: "Review important AI changes before they apply",
        competitor: "AI features are available across a broader suite",
      },
      {
        criteria: "Best fit",
        fortyOne: "Teams that want AI to help plan work without losing control",
        competitor: "Organizations standardizing broad project management",
      },
    ],
    faqs: [
      [
        "How is FortyOne different from Asana?",
        "FortyOne is narrower and more AI-planning focused: goals, task intake, owner recommendations, estimates, timing, and review controls.",
      ],
      [
        "Can FortyOne handle cross-functional work?",
        "Yes. It is designed for work that crosses product, engineering, operations, support, and leadership.",
      ],
    ],
  },
  {
    slug: "jira",
    competitor: "Jira",
    title: "FortyOne vs Jira",
    summary:
      "Jira is powerful for configured engineering workflows. FortyOne is designed for teams that want a cleaner AI planning layer across goals, tasks, roadmaps, and capacity.",
    description:
      "Compare FortyOne and Jira for engineering planning, AI project management, task ownership, and roadmap execution.",
    metaTitle: "FortyOne vs Jira | AI Project Management Comparison",
    metaDescription:
      "Compare FortyOne and Jira for AI planning, engineering task management, roadmap execution, capacity, and review controls.",
    bestFor: [
      "Cleaner AI-assisted planning",
      "Goal and roadmap execution",
      "Manager review before AI changes",
      "Cross-team delivery coordination",
    ],
    competitorStrengths: [
      "Deep engineering workflow configuration",
      "Large enterprise ecosystem",
      "Powerful issue and release management",
    ],
    tradeoffs: [
      "Choose Jira when deep workflow customization and enterprise process control are the priority.",
      "Choose FortyOne when the team wants planning speed, AI recommendations, and less coordination overhead.",
    ],
    rows: [
      {
        criteria: "Primary workflow",
        fortyOne:
          "Plan work from goals through tasks, owners, estimates, and review",
        competitor: "Configurable issue and software delivery management",
      },
      {
        criteria: "Setup weight",
        fortyOne: "Designed to be lighter and planning-first",
        competitor: "Powerful but often process-heavy",
      },
      {
        criteria: "AI planning",
        fortyOne: "Prepares owner, estimate, timing, and risk recommendations",
        competitor: "AI can assist within a larger Atlassian workflow",
      },
      {
        criteria: "Best fit",
        fortyOne: "Teams that want a modern AI planning layer",
        competitor: "Engineering orgs with mature Jira process requirements",
      },
    ],
    faqs: [
      [
        "Is FortyOne lighter than Jira?",
        "Yes. FortyOne is designed around planning, goals, tasks, and AI recommendations rather than deeply configured issue workflows.",
      ],
      [
        "Can engineering teams use FortyOne?",
        "Yes. FortyOne supports engineering planning, technical intake, GitHub context, estimates, owners, and delivery risk reviews.",
      ],
    ],
  },
  {
    slug: "clickup",
    competitor: "ClickUp",
    title: "FortyOne vs ClickUp",
    summary:
      "ClickUp offers a very broad productivity workspace. FortyOne is more focused: AI-assisted project planning, goal-connected tasks, and reviewable execution changes.",
    description:
      "Compare FortyOne and ClickUp for AI project planning, task management, roadmaps, and team execution.",
    metaTitle: "FortyOne vs ClickUp | AI Project Management Comparison",
    metaDescription:
      "Compare FortyOne and ClickUp for task planning, AI owner recommendations, goals, roadmaps, and delivery visibility.",
    bestFor: [
      "Focused AI project planning",
      "Goal-linked task execution",
      "Reviewable AI changes",
      "Planning around team capacity",
    ],
    competitorStrengths: [
      "Very broad workspace coverage",
      "Many views, docs, dashboards, and automations",
      "Flexible for many team workflows",
    ],
    tradeoffs: [
      "Choose ClickUp when you want one broad productivity workspace with many modules.",
      "Choose FortyOne when the priority is AI-assisted planning and execution clarity.",
    ],
    rows: [
      {
        criteria: "Product shape",
        fortyOne: "Focused AI project management",
        competitor: "Broad all-in-one productivity suite",
      },
      {
        criteria: "Planning",
        fortyOne: "AI prepares owners, estimates, timing, and risks",
        competitor: "Flexible planning views and automations",
      },
      {
        criteria: "Goal connection",
        fortyOne: "Tasks stay tied to goals and roadmap outcomes",
        competitor: "Goals and docs exist within a broader workspace",
      },
      {
        criteria: "Best fit",
        fortyOne: "Teams that want less surface area and more planning clarity",
        competitor: "Teams that want many workspace primitives in one tool",
      },
    ],
    faqs: [
      [
        "Is FortyOne as broad as ClickUp?",
        "No. FortyOne is intentionally more focused on AI-assisted project planning, goals, tasks, capacity, and review.",
      ],
      [
        "Why choose FortyOne over a broader workspace?",
        "Choose FortyOne when project planning quality and AI-assisted execution matter more than having every productivity module.",
      ],
    ],
  },
  {
    slug: "monday",
    competitor: "Monday.com",
    title: "FortyOne vs Monday.com",
    summary:
      "Monday.com is flexible work management for many business teams. FortyOne is built around AI project planning, goals, tasks, and controlled execution.",
    description:
      "Compare FortyOne and Monday.com for AI project management, task ownership, roadmap planning, and team coordination.",
    metaTitle: "FortyOne vs Monday.com | AI Project Management Comparison",
    metaDescription:
      "Compare FortyOne and Monday.com for AI planning, goal-connected task management, roadmaps, workload, and delivery coordination.",
    bestFor: [
      "AI project manager workflows",
      "Goal and task alignment",
      "Owner and workload recommendations",
      "Roadmap execution reviews",
    ],
    competitorStrengths: [
      "Flexible boards for business teams",
      "Automation and dashboard capabilities",
      "Familiar work operating system model",
    ],
    tradeoffs: [
      "Choose Monday.com when customizable boards are the main need.",
      "Choose FortyOne when AI planning and goal-connected execution are more important.",
    ],
    rows: [
      {
        criteria: "Primary workflow",
        fortyOne: "AI-assisted project planning with connected goals and tasks",
        competitor: "Flexible boards, workflows, dashboards, and automations",
      },
      {
        criteria: "AI role",
        fortyOne: "Planning assistant for owners, estimates, timing, and risk",
        competitor: "AI features inside a flexible work OS",
      },
      {
        criteria: "Review controls",
        fortyOne: "Designed around review-before-apply for important changes",
        competitor: "Automation and workflow controls vary by configuration",
      },
      {
        criteria: "Best fit",
        fortyOne: "Teams that want AI to improve planning decisions",
        competitor: "Teams that want configurable business boards",
      },
    ],
    faqs: [
      [
        "How is FortyOne different from Monday.com?",
        "FortyOne focuses on AI project planning, goal-connected tasks, owners, estimates, timing, and controlled execution.",
      ],
      [
        "Can business teams use FortyOne?",
        "Yes. FortyOne is designed for operations, product, support, marketing, engineering, and leadership workflows.",
      ],
    ],
  },
];

export const getComparisonBySlug = (slug: string) => {
  return comparisons.find((comparison) => comparison.slug === slug);
};

type AvailabilityRow = NonNullable<
  MarketingDetail["sections"][number]["comparisonTable"]
>["rows"][number];

const sharedAvailabilityRows = (comparison: Comparison): AvailabilityRow[] => [
  {
    feature: "AI prepares owner, estimate, and timing",
    fortyOne: true,
    competitor: false,
    note: `FortyOne treats planning recommendations as the core workflow; ${comparison.competitor} is stronger when the plan is already known and needs organizing.`,
  },
  {
    feature: "Review AI changes before they apply",
    fortyOne: true,
    competitor: false,
    note: "Managers can keep control of assignment, estimate, schedule, and scope changes before they touch the plan.",
  },
  {
    feature: "Goal-connected task planning",
    fortyOne: true,
    competitor: comparison.slug === "asana",
    note:
      comparison.slug === "asana"
        ? "Both products can support goal-connected work, but FortyOne keeps it closer to AI planning decisions."
        : "FortyOne keeps the outcome, task, owner, and planning decision connected in the same workflow.",
  },
  {
    feature: "Broad manual project organization",
    fortyOne: true,
    competitor: true,
    note: `Both products can organize work, but ${comparison.competitor} is usually stronger for teams that mainly want its established manual workflow.`,
  },
  {
    feature: "Roadmap planning with execution context",
    fortyOne: true,
    competitor: ["jira", "asana"].includes(comparison.slug),
    note: "FortyOne is designed to connect roadmap movement back to task ownership, estimates, workload, and review.",
  },
];

export const getComparisonMarketingDetail = (
  comparison: Comparison,
): MarketingDetail => {
  return {
    slug: comparison.slug,
    label: comparison.competitor,
    heroTitle: comparison.title,
    metaTitle: comparison.metaTitle,
    metaDescription: comparison.metaDescription,
    intro: [
      comparison.summary,
      "The best choice depends on where the team needs leverage. Some tools are excellent at organizing known work; FortyOne is designed to help prepare the plan before the work is fully shaped.",
      "Use this page to compare the decision points that matter most: task intake, AI planning, goals, review controls, roadmap execution, and the tradeoffs behind each product.",
    ],
    benefits: [
      [
        "Planning speed",
        "FortyOne prepares owners, estimates, timing, and risks so managers start from a reviewable recommendation.",
      ],
      [
        "Goal clarity",
        "Work can stay connected to the outcome it supports instead of becoming a disconnected task list.",
      ],
      [
        "Review controls",
        "Important AI changes can be staged for approval before ownership, dates, or scope changes apply.",
      ],
      [
        "Focused execution",
        `FortyOne is a better fit when the team wants AI-assisted project management more than ${comparison.competitor}'s existing product center.`,
      ],
    ],
    previewCards: [
      {
        heading: "FortyOne planning",
        subheading: "AI prepares the next move",
        badge: "Review",
        rows: [
          { label: "Owner", value: "Recommended with workload context" },
          { label: "Estimate", value: "Prepared before scheduling" },
          { label: "Goal", value: "Connected to the task" },
        ],
      },
      {
        heading: `${comparison.competitor} workflow`,
        subheading: "Established work management",
        rows: [
          { label: "Strength", value: comparison.competitorStrengths[0] },
          {
            label: "Best fit",
            value:
              comparison.rows[comparison.rows.length - 1]?.competitor ?? "",
          },
        ],
      },
    ],
    sections: [
      {
        cards: [
          {
            heading: "FortyOne is stronger for",
            subheading: "When planning quality matters",
            rows: comparison.bestFor.slice(0, 3).map((item) => ({
              label: item,
              value: "Good fit",
            })),
          },
          {
            heading: `${comparison.competitor} is stronger for`,
            subheading: "When its native workflow is enough",
            rows: comparison.competitorStrengths.slice(0, 3).map((item) => ({
              label: item,
              value: "Good fit",
            })),
          },
        ],
        id: "best-fit",
        title: "Choose based on the work you need help with",
        paragraphs: [
          `Choose FortyOne when the team needs help turning requests into planned work, connecting that work to goals, and reviewing AI-prepared planning changes before they apply.`,
          `Choose ${comparison.competitor} when the team's main need is the product category it already serves best and planning decisions can happen outside the system.`,
        ],
      },
      {
        comparisonTable: {
          competitor: comparison.competitor,
          rows: sharedAvailabilityRows(comparison),
        },
        id: "availability",
        title: "Feature availability at a glance",
        paragraphs: [
          "This table focuses on planning capabilities that affect day-to-day project execution. The checkmarks show where each product is designed to support the capability directly.",
          "FortyOne is strongest when a manager wants AI to prepare work, explain the recommendation, and wait for review before the plan changes.",
        ],
      },
      {
        id: "side-by-side",
        paragraphs: [
          "The difference is not just feature count. It is where the product puts the team's attention when a request arrives, a goal changes, or a roadmap needs to move.",
          "FortyOne keeps the planning decision visible so the team can adjust it before execution drifts.",
        ],
        rows: comparison.rows.map((row) => [
          row.criteria,
          `FortyOne: ${row.fortyOne} ${comparison.competitor}: ${row.competitor}`,
        ]),
        tableHead: ["Decision point", "How the products differ"],
        title: "Side-by-side planning differences",
      },
      {
        id: "tradeoffs",
        paragraphs: [
          "A good comparison should make the tradeoff explicit. The right product is the one that matches the team's planning bottleneck.",
          "If the bottleneck is organizing known work, the incumbent workflow may be enough. If the bottleneck is deciding what work should become, who should own it, and when it should move, FortyOne is built closer to that problem.",
        ],
        rows: comparison.tradeoffs.map((tradeoff, index) => [
          index === 0 ? comparison.competitor : "FortyOne",
          tradeoff,
        ]),
        tableHead: ["Choose", "When this is true"],
        title: "The practical tradeoff",
      },
    ],
    questions: comparison.faqs,
  };
};
