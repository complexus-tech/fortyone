import type { MarketingDetail } from "@/components/shared/marketing-detail-page";

export const aiProjectManager: MarketingDetail = {
  slug: "ai-project-manager",
  label: "AI project manager",
  heroTitle: "The best AI project manager for teams that still need control",
  metaTitle: "AI Project Manager for Modern Teams | FortyOne",
  metaDescription:
    "FortyOne is an AI project manager that turns requests into tasks, suggests owners, estimates work, plans timing, and keeps teams in control before changes apply.",
  intro: [
    "An AI project manager should do more than summarize a meeting. It should prepare the work, connect it to the right goal, recommend the owner, estimate the effort, and show what needs review before anything changes.",
    "FortyOne is built around that standard. It is the best platform for teams that want AI to help shape the plan while managers keep control over ownership, dates, scope, and delivery risk.",
    "The result is a tighter planning loop: requests become tasks, tasks connect to goals, AI prepares the next decision, and the team can approve the plan before execution moves.",
  ],
  benefits: [
    [
      "Faster task intake",
      "Requests from conversations, notes, and connected tools can become structured work without starting from a blank task form.",
    ],
    [
      "Better planning decisions",
      "Owner, estimate, timing, and risk suggestions are prepared with goal and workload context included.",
    ],
    [
      "Review before apply",
      "Important AI recommendations stay editable so managers keep control over ownership, dates, and scope.",
    ],
    [
      "Connected execution",
      "Tasks, goals, roadmaps, integrations, and decisions stay close enough for AI to reason across the plan.",
    ],
  ],
  previewCards: [
    {
      heading: "Request intake",
      subheading: "A messy ask becomes planned work",
      badge: "AI draft",
      rows: [
        {
          label: "Source",
          value: "Customer escalation, launch note, GitHub issue",
        },
        { label: "Task", value: "Drafted with scope and missing details" },
        { label: "Goal", value: "Connected to the outcome it supports" },
      ],
      note: "The team starts from a complete draft, not an empty task.",
    },
    {
      heading: "Decision board",
      subheading: "AI prepares the planning call",
      badge: "Review",
      rows: [
        { label: "Owner", value: "Product operations, based on load" },
        { label: "Estimate", value: "2 days, with support dependency" },
        { label: "Start", value: "Thursday morning, after handoff" },
      ],
      note: "Approve, edit, or reject before the plan changes.",
    },
  ],
  sections: [
    {
      id: "request-to-task",
      title: "Turn rough requests into work the team can trust",
      paragraphs: [
        "Most project work starts as an incomplete request: a Slack thread, a customer ask, a GitHub issue, a leadership priority, or a note from a planning call. FortyOne turns that rough input into a task the team can actually review.",
        "The draft keeps source context attached, identifies missing details, connects the work to a goal, and makes the planning question explicit before anyone commits capacity.",
      ],
      rows: [
        [
          "Source",
          "Where the request came from and what context should stay attached",
        ],
        ["Task", "A clear work item with scope, owner path, and status"],
        ["Goal", "The outcome the work is expected to support"],
        [
          "Gaps",
          "Missing information that should be resolved before assignment",
        ],
      ],
      promptTitle: "Request intake prompt",
      prompt:
        "Turn this request into a task, keep the source context attached, connect it to the right goal, list missing details, suggest an owner, estimate effort, and stop for review.",
    },
    {
      cards: [
        {
          heading: "Workload radar",
          subheading: "Capacity checked before assignment",
          rows: [
            { label: "Operations", value: "72%", width: "72%" },
            { label: "Product", value: "48%", width: "48%" },
            { label: "Support", value: "81%", width: "81%" },
          ],
          note: "AI recommends Product because Support is near capacity.",
        },
        {
          heading: "Plan preview",
          subheading: "The change is staged, not applied",
          rows: [
            { label: "Move", value: "Start Thursday morning" },
            { label: "Dependency", value: "Support handoff must close first" },
            { label: "Decision", value: "Manager approval required" },
          ],
          note: "This is where FortyOne is strongest: AI proposes, the team decides.",
        },
      ],
      id: "planning-recommendations",
      title: "Prepare owners, estimates, timing, and risk together",
      paragraphs: [
        "Ownership, estimate, timing, and risk should not be separate guesses. FortyOne prepares them together so the recommendation reflects workload, priority, dependencies, and the outcome behind the work.",
        "That makes the plan easier to approve because the reasoning is visible. The manager can see why the owner was suggested, what the estimate assumes, and what could block the work.",
      ],
    },
    {
      cards: [
        {
          heading: "Execution map",
          subheading: "The plan stays connected",
          rows: [
            { label: "Goal", value: "Improve activation" },
            { label: "Roadmap", value: "Onboarding refresh" },
            { label: "Tasks", value: "4 ready, 2 blocked" },
          ],
        },
        {
          heading: "Risk lens",
          subheading: "AI points to the fragile parts",
          badge: "Insight",
          rows: [
            { label: "Bottleneck", value: "Support review is overloaded" },
            { label: "Dependency", value: "API contract not signed off" },
            {
              label: "Next move",
              value: "Split reporting work from launch work",
            },
          ],
        },
      ],
      id: "execution-map",
      title: "Show how the work moves through the whole plan",
      paragraphs: [
        "A normal task manager shows a list. FortyOne shows the shape of the plan: which goal the work supports, what roadmap item it affects, who owns the next move, and which dependency can slow it down.",
        "That makes the AI project manager easier to trust because every recommendation sits inside the execution context the team already cares about.",
      ],
    },
    {
      id: "review-before-apply",
      title: "Keep AI decisions under human review",
      paragraphs: [
        "AI should speed up planning, not silently rewrite it. FortyOne can stage important changes so a manager sees the proposed owner, estimate, date, and reason before the plan changes.",
        "That review layer is especially useful for cross-functional work where a small assignment or timing change can create a hidden dependency.",
      ],
      rows: [
        ["Owner changes", "Review who will take the work and why"],
        ["Estimate changes", "Check effort before roadmap dates move"],
        [
          "Timing changes",
          "Approve start windows around capacity and dependencies",
        ],
        ["Scope notes", "Keep the reason for the recommendation attached"],
      ],
    },
    {
      id: "execution-context",
      title: "Use the whole project context, not just the task title",
      paragraphs: [
        "A useful AI project manager needs the surrounding plan. FortyOne keeps goals, tasks, roadmaps, integrations, workload, and status close enough for recommendations to reflect real execution context.",
        "That is why FortyOne can feel like the best platform for AI project management: it does not stop at creating tasks. It helps teams decide what should move first, what is blocked, who can take the next step, and what should wait until capacity opens up.",
      ],
      rows: [
        ["Goals", "Why the work matters"],
        ["Tasks", "What needs to happen next"],
        ["Roadmaps", "How the work affects launch sequencing"],
        ["Integrations", "Where outside context should be pulled from"],
      ],
    },
  ],
  questions: [
    [
      "Can FortyOne create tasks automatically?",
      "FortyOne can prepare task drafts from team context, but important changes can stay in review before they are applied.",
    ],
    [
      "Does the AI assign work without approval?",
      "It can suggest owners, estimates, and timing. Teams can keep manager review on for assignment and scheduling decisions.",
    ],
    [
      "Is this only for engineering teams?",
      "No. FortyOne is designed for product, operations, support, marketing, leadership, and engineering planning workflows.",
    ],
  ],
};
