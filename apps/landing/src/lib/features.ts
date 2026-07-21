import type { MarketingDetail } from "@/components/shared/marketing-detail-page";
import { featureLabels } from "./feature-links";

export type Feature = MarketingDetail;

export const features: Feature[] = [
  {
    slug: "customer-feedback",
    label: featureLabels["customer-feedback"],
    heroTitle: "FortyOne Customer Feedback",
    metaTitle: "Customer Feedback Management and Roadmaps | FortyOne",
    metaDescription:
      "Collect customer requests, prioritize feedback, publish roadmap progress, and turn accepted ideas into project work without losing the original request.",
    intro: [
      "Customer feedback is most useful when teams can connect it to a decision and show what happens next. Otherwise, requests collect in separate tools while the project plan moves on without them.",
      "FortyOne gives customers a place to submit requests, vote, and follow progress. Teams can organize feedback into boards, review what matters, and move accepted requests into the project plan.",
      "The original request stays connected to the task, owner, estimate, goal, and delivery status so teams can act without losing the customer context behind the work.",
    ],
    benefits: [
      [
        "One feedback portal",
        "Customers can submit requests, vote on ideas, and follow progress from one public workspace.",
      ],
      [
        "Clear prioritization",
        "Teams can compare feedback by board, status, votes, customer context, and product priorities.",
      ],
      [
        "Connected delivery",
        "Accepted feedback becomes planned work while the original request remains attached.",
      ],
      [
        "Visible progress",
        "Public roadmaps and status updates show customers what is planned, active, completed, or closed.",
      ],
    ],
    previewCards: [
      {
        heading: "Customer request",
        subheading: "Feedback ready for review",
        badge: "12 votes",
        rows: [
          { label: "Board", value: "Product feedback" },
          { label: "Status", value: "Reviewing" },
        ],
      },
      {
        heading: "Project work",
        subheading: "Accepted request connected to delivery",
        rows: [
          { label: "Owner", value: "Product team" },
          { label: "Roadmap", value: "Planned for Q3" },
        ],
      },
    ],
    sections: [
      {
        id: "collect",
        title: "Collect customer feedback in one place",
        paragraphs: [
          "Customers should not need to send the same request through email, support chat, and meetings. FortyOne provides a public feedback portal where they can submit ideas, add context, vote, and see related requests.",
          "Feedback boards help route each request to the team that owns the decision without exposing the internal project workspace.",
        ],
        rows: [
          ["Submit", "Customers describe the request and why it matters"],
          ["Vote", "Other customers can support an existing idea"],
          ["Organize", "Boards route feedback to the responsible team"],
          ["Discuss", "Comments keep clarification attached to the request"],
        ],
        promptTitle: "Feedback review",
        prompt:
          "Review active customer feedback, group related requests, summarize the underlying need, and show which items need a product decision.",
      },
      {
        id: "prioritize",
        title: "Prioritize requests with customer and project context",
        paragraphs: [
          "Vote counts are useful, but they are not the whole decision. FortyOne helps teams review the customer need alongside company goals, existing work, delivery risk, and available capacity.",
          "Teams can move requests through reviewing, planned, in progress, completed, or closed states while keeping the decision visible.",
        ],
        cards: [
          {
            heading: "Feedback signal",
            subheading: "Requests grouped before planning",
            rows: [
              { label: "Related requests", value: "8 customers" },
              { label: "Primary need", value: "Faster reporting exports" },
            ],
          },
          {
            heading: "Planning decision",
            subheading: "Context ready for review",
            badge: "Planned",
            rows: [
              { label: "Goal", value: "Improve customer retention" },
              { label: "Owner", value: "Platform team" },
            ],
          },
        ],
      },
      {
        id: "deliver",
        title: "Turn accepted feedback into trackable project work",
        paragraphs: [
          "Once a request is accepted, FortyOne can connect it to a task with an owner, estimate, schedule, goal, and delivery status.",
          "The task keeps a link to the original feedback, and customers can follow progress on the public roadmap without seeing internal planning details.",
        ],
        rows: [
          ["Request", "The customer problem and supporting discussion"],
          ["Task", "The work the team accepted into the project plan"],
          ["Owner", "Who is responsible for delivery"],
          ["Roadmap", "The customer-facing status of the work"],
        ],
      },
    ],
    questions: [
      [
        "Can customers submit and vote on feedback?",
        "Yes. Customers can submit requests, vote on ideas, comment, and follow status changes from a public feedback portal.",
      ],
      [
        "Can feedback become a project task?",
        "Yes. Accepted feedback can be connected to a task with an owner, estimate, goal, schedule, and delivery status.",
      ],
      [
        "Can customers see what the team is working on?",
        "Yes. Teams can publish roadmap progress while keeping internal project details private.",
      ],
    ],
  },
  {
    slug: "goals",
    label: featureLabels.goals,
    heroTitle: "FortyOne Goals",
    metaTitle: "Goals | Connect Work to Outcomes | FortyOne",
    metaDescription:
      "See how FortyOne keeps goals, tasks, owners, progress, and AI planning connected in one project management workflow.",
    intro: [
      "Goals are where the work gets its reason. Without them, tasks become a queue of activity that is hard to prioritize, explain, or defend when capacity gets tight.",
      "FortyOne keeps goals close to execution. Teams can connect tasks, projects, owners, decisions, and progress to the outcome they support, then use AI to prepare the next move with that context in view.",
      "That gives managers and teams a clearer answer to a simple question: what work is actually moving the outcome forward?",
    ],
    benefits: [
      [
        "Clearer prioritization",
        "Teams can compare tasks against the goals they support instead of debating work in isolation.",
      ],
      [
        "Better progress visibility",
        "Goal health, blocked work, owners, and next decisions stay connected to the underlying project plan.",
      ],
      [
        "Less status reconstruction",
        "Reviews can start from live goal-linked work instead of copied updates and disconnected notes.",
      ],
      [
        "Stronger AI context",
        "AI can recommend owners, timing, and next steps with the goal behind the work included.",
      ],
    ],
    previewCards: [
      {
        heading: "Activation goal",
        subheading: "Tasks connected to one outcome",
        badge: "On track",
        rows: [
          { label: "Progress", value: "64%", width: "64%" },
          { label: "At risk", value: "2 tasks need decisions" },
        ],
      },
      {
        heading: "Goal review",
        subheading: "AI prepares the next decision",
        rows: [
          { label: "Next step", value: "Move onboarding work up" },
          { label: "Owner", value: "Product lead" },
        ],
      },
    ],
    sections: [
      {
        id: "goal-linked-work",
        title: "Connect every task to the outcome it supports",
        paragraphs: [
          "A task should not sit alone. FortyOne helps teams attach work to goals so the reason for the task travels with the assignment, estimate, source links, and status.",
          "That makes planning easier because the team can see which tasks protect the most important outcomes and which work can wait.",
        ],
        rows: [
          ["Goal", "The outcome the work is meant to improve"],
          ["Task", "The concrete work tied to that outcome"],
          ["Owner", "Who is responsible for moving the task forward"],
          ["Status", "Whether the work is done, blocked, at risk, or ready"],
        ],
        promptTitle: "Goal review",
        prompt:
          "Review this goal and its connected tasks. Show what is moving, what is blocked, who owns each next step, and what should be prioritized this week.",
      },
      {
        id: "progress",
        title: "Review progress without rebuilding status",
        paragraphs: [
          "Goal reviews often become a manual reporting exercise. FortyOne keeps goal progress connected to real work so teams can see the tasks, blockers, owners, and decisions behind the update.",
          "AI can prepare a review note from the connected project data while managers stay in control of what gets shared or changed.",
        ],
        cards: [
          {
            heading: "Goal health",
            subheading: "Progress and risk in one view",
            rows: [
              { label: "Completed", value: "11 tasks" },
              { label: "Blocked", value: "3 tasks" },
            ],
          },
          {
            heading: "Decision queue",
            subheading: "What needs leadership input",
            badge: "Review",
            rows: [
              { label: "Scope", value: "Pause reporting refresh" },
              { label: "Capacity", value: "Support near limit" },
            ],
          },
        ],
      },
      {
        id: "ai-context",
        title: "Give AI the goal before it suggests the plan",
        paragraphs: [
          "Owner recommendations and schedule suggestions are better when AI understands what the work is trying to accomplish.",
          "FortyOne lets AI use the goal, source context, current workload, estimates, and deadlines together before it prepares a plan for review.",
        ],
        rows: [
          ["Priority", "Which outcome the work supports"],
          ["Capacity", "Who can take the next task without creating risk"],
          ["Timing", "When the work should start based on availability"],
          ["Review", "What should be approved before changes are applied"],
        ],
      },
    ],
    questions: [
      [
        "Can tasks exist without goals?",
        "Yes. Teams can create standalone tasks, but goal-linked work gives planning and AI recommendations better context.",
      ],
      [
        "Can AI summarize goal progress?",
        "AI can prepare a review from connected tasks, blockers, owners, estimates, and recent changes.",
      ],
      [
        "Is this only for OKRs?",
        "No. Goals can represent OKRs, launch outcomes, operational priorities, or any measurable team objective.",
      ],
    ],
  },
  {
    slug: "tasks",
    label: featureLabels.tasks,
    heroTitle: "FortyOne Tasks",
    metaTitle: "Tasks | AI-Assisted Task Management | FortyOne",
    metaDescription:
      "See how FortyOne turns requests, context, owners, estimates, and status into task management that stays connected to goals.",
    intro: [
      "Tasks are the execution layer of FortyOne. They hold the work, the owner, the status, the estimate, the source context, and the goal the work supports.",
      "FortyOne is designed for teams that need tasks to be more than a checklist. A task can start from Slack, planning notes, GitHub context, or a manager request, then become assigned work with review controls.",
      "That helps teams move faster without losing the judgment, accountability, and context that good project management depends on.",
    ],
    benefits: [
      [
        "Cleaner intake",
        "Requests can become structured tasks with source links, missing details, and a clear owner path.",
      ],
      [
        "Better ownership",
        "Tasks can carry owner recommendations, estimates, start windows, and review notes before assignment.",
      ],
      [
        "Less context loss",
        "Slack threads, GitHub issues, files, and goals can stay attached to the work they created.",
      ],
      [
        "Clearer status",
        "Teams can see blocked, overdue, active, and review-ready tasks without rebuilding a status report.",
      ],
    ],
    previewCards: [
      {
        heading: "Task intake",
        subheading: "Request converted into work",
        badge: "AI draft",
        rows: [
          { label: "Task", value: "Prepare onboarding report" },
          { label: "Goal", value: "Improve activation" },
        ],
      },
      {
        heading: "Assignment",
        subheading: "Owner and timing ready to review",
        rows: [
          { label: "Owner", value: "Operations" },
          { label: "Estimate", value: "2 days" },
        ],
      },
    ],
    sections: [
      {
        id: "intake",
        title: "Turn requests into tasks the team can act on",
        paragraphs: [
          "Work often starts before it is fully formed. FortyOne helps teams turn rough requests into tasks with enough structure to review, assign, and track.",
          "Instead of copying context from one tool to another, teams can keep the original source attached and let AI prepare the first version of the task.",
        ],
        rows: [
          [
            "Request",
            "A Slack thread, meeting note, support ask, or product idea",
          ],
          ["Task", "A scoped piece of work with owner, goal, and status"],
          [
            "Missing details",
            "Open questions that should be resolved before assignment",
          ],
          ["Review", "The proposed task ready for manager edits or approval"],
        ],
        promptTitle: "Task intake",
        prompt:
          "Turn this request into a task, attach source context, suggest the goal, list missing details, recommend an owner, and stop for review.",
      },
      {
        id: "ownership",
        title: "Assign work with context, capacity, and timing",
        paragraphs: [
          "A good task assignment is not just about who could do the work. It is about who can do it at the right time without creating a hidden bottleneck.",
          "FortyOne can help compare current workload, estimates, availability, and related context before work moves.",
        ],
        cards: [
          {
            heading: "Owner suggestion",
            subheading: "Workload checked before assignment",
            rows: [
              { label: "Operations", value: "68%", width: "68%" },
              { label: "Product", value: "42%", width: "42%" },
            ],
            note: "AI recommends Product for this task.",
          },
          {
            heading: "Review before apply",
            subheading: "Assignment stays editable",
            rows: [
              { label: "Owner", value: "Maya Chen" },
              { label: "Start", value: "Thursday morning" },
            ],
          },
        ],
      },
      {
        id: "status",
        title: "Keep task status tied to the plan",
        paragraphs: [
          "Status should not be a second artifact. FortyOne keeps the task, owner, estimate, blocker, and goal together so updates remain useful after the meeting ends.",
          "Teams can scan what is active, what is blocked, what needs review, and what has changed since the last check-in.",
        ],
        rows: [
          ["Active", "Work in progress with owner and timing visible"],
          ["Blocked", "Tasks waiting on a decision, dependency, or handoff"],
          ["Review", "AI-prepared changes that need approval"],
          ["Done", "Completed work still connected to the goal it supported"],
        ],
      },
    ],
    questions: [
      [
        "Can tasks be created from Slack?",
        "Yes. Slack can be an intake source, with AI helping turn the conversation into a structured task.",
      ],
      [
        "Can AI assign tasks automatically?",
        "AI can suggest owners, estimates, and timing. Teams can keep review controls before changes are applied.",
      ],
      [
        "Can tasks connect to goals?",
        "Yes. Tasks can be connected to goals so teams understand why the work exists.",
      ],
    ],
  },
  {
    slug: "ai-planning",
    label: featureLabels["ai-planning"],
    heroTitle: "FortyOne AI Planning",
    metaTitle: "AI Planning | Owner, Estimate, and Schedule Support | FortyOne",
    metaDescription:
      "See how FortyOne uses AI to prepare owners, estimates, timing, risks, and review-ready project changes.",
    intro: [
      "AI planning in FortyOne is not a black box that silently moves work around. It is a planning layer that prepares recommendations and keeps managers in control.",
      "AI can read the goal, task context, team workload, estimates, calendar availability, and connected tools before suggesting the next owner, start window, and risk notes.",
      "That gives teams the speed of AI without giving up the review step that protects trust, accountability, and project quality.",
    ],
    benefits: [
      [
        "Faster planning",
        "AI can prepare a first version of owner, estimate, start time, and next-step recommendations.",
      ],
      [
        "Better capacity decisions",
        "Recommendations can consider current workload and availability before work is assigned.",
      ],
      [
        "Earlier risk detection",
        "Blocked dependencies, missing estimates, and overloaded owners can surface before deadlines slip.",
      ],
      [
        "Controlled execution",
        "Important changes can stay staged for review instead of being applied without context.",
      ],
    ],
    previewCards: [
      {
        heading: "Planning recommendation",
        subheading: "AI prepares the next move",
        badge: "Review",
        rows: [
          { label: "Owner", value: "Frontend team" },
          { label: "Start", value: "Tomorrow morning" },
        ],
      },
      {
        heading: "Capacity check",
        subheading: "Workload before assignment",
        rows: [
          { label: "Platform", value: "82%", width: "82%" },
          { label: "Frontend", value: "54%", width: "54%" },
        ],
        note: "AI recommends Frontend to reduce delivery risk.",
      },
    ],
    sections: [
      {
        id: "recommendations",
        title: "Prepare owner, estimate, and timing recommendations",
        paragraphs: [
          "Planning often slows down because the team needs to gather context before making a decision. FortyOne helps AI prepare that context and turn it into an editable recommendation.",
          "The result is a planning draft that managers can approve, adjust, or reject before it changes the project.",
        ],
        rows: [
          ["Owner", "Who is best placed to take the work"],
          [
            "Estimate",
            "A first effort estimate based on context and similar tasks",
          ],
          [
            "Start window",
            "When the work can begin without overloading the team",
          ],
          ["Risk note", "What could block the work or affect timing"],
        ],
        promptTitle: "Planning review",
        prompt:
          "Review this task and recommend an owner, estimate, start window, and risk note using current workload, calendar availability, goal context, and linked source material.",
      },
      {
        id: "review-control",
        title: "Keep review before apply",
        paragraphs: [
          "AI can move fast, but project changes still need accountability. FortyOne is designed so important recommendations can be reviewed before they change ownership, timing, estimates, or status.",
          "That gives teams a practical way to use AI for planning while preserving manager judgment.",
        ],
        cards: [
          {
            heading: "Staged change",
            subheading: "Editable before it moves work",
            badge: "Pending",
            rows: [
              { label: "Owner", value: "Assign to Maya Chen" },
              { label: "Estimate", value: "Set to 3 days" },
            ],
          },
          {
            heading: "Approval note",
            subheading: "Why the recommendation was made",
            rows: [
              { label: "Reason", value: "Lowest current workload" },
              { label: "Risk", value: "API review still pending" },
            ],
          },
        ],
      },
      {
        id: "risk",
        title: "Spot planning risk earlier",
        paragraphs: [
          "Planning risk often appears before a deadline is missed: a missing owner, unclear estimate, blocked dependency, or overloaded team.",
          "FortyOne can surface those signals so teams can adjust the plan while there is still time to act.",
        ],
        rows: [
          [
            "Missing owner",
            "Work that cannot move until responsibility is clear",
          ],
          ["Missing estimate", "Tasks that need effort before scheduling"],
          ["Capacity conflict", "Owners or teams nearing workload limits"],
          ["Blocked dependency", "Work waiting on another decision or handoff"],
        ],
      },
    ],
    questions: [
      [
        "Does AI change the plan automatically?",
        "AI can prepare recommendations. Teams can keep important changes staged for review before they are applied.",
      ],
      [
        "What context can AI use?",
        "AI can use goals, tasks, estimates, workload, calendar availability, and connected tool context when available.",
      ],
      [
        "Can managers edit AI recommendations?",
        "Yes. Recommendations are meant to be reviewed, edited, approved, or rejected by the team.",
      ],
    ],
  },
  {
    slug: "roadmaps",
    label: featureLabels.roadmaps,
    heroTitle: "FortyOne Roadmaps",
    metaTitle: "Roadmaps | Connect Priorities to Execution | FortyOne",
    metaDescription:
      "See how FortyOne helps teams turn priorities, launches, capacity, tasks, and status into execution-ready roadmaps.",
    intro: [
      "A roadmap is only useful if it stays connected to the work required to deliver it. Otherwise, it becomes a promise that drifts away from capacity, blockers, and actual progress.",
      "FortyOne helps teams connect roadmap priorities to goals, tasks, owners, estimates, launch work, and AI-prepared planning decisions.",
      "That makes roadmap reviews more grounded because leaders can see not only what is planned, but what is moving, what is blocked, and what tradeoffs need attention.",
    ],
    benefits: [
      [
        "Execution visibility",
        "Roadmap items stay connected to the tasks, owners, estimates, and blockers behind them.",
      ],
      [
        "Better tradeoff decisions",
        "Teams can compare priority, capacity, and timing before adding more work.",
      ],
      [
        "Cleaner launch planning",
        "Cross-functional launch work can be tracked from planning through delivery.",
      ],
      [
        "Faster roadmap reviews",
        "AI can prepare status, risk, and decision notes from live project data.",
      ],
    ],
    previewCards: [
      {
        heading: "Roadmap item",
        subheading: "Priority connected to delivery work",
        badge: "Q3",
        rows: [
          { label: "Goal", value: "Improve activation" },
          { label: "Status", value: "Design review pending" },
        ],
      },
      {
        heading: "Launch readiness",
        subheading: "Cross-functional work in one view",
        rows: [
          { label: "Engineering", value: "API handoff due Friday" },
          { label: "Marketing", value: "Launch copy pending" },
        ],
      },
    ],
    sections: [
      {
        id: "priorities",
        title: "Connect roadmap priorities to the work behind them",
        paragraphs: [
          "Roadmap priorities need a clear path into execution. FortyOne helps teams attach the goal, scope, tasks, owners, estimates, and current status to each priority.",
          "That keeps the roadmap connected to reality as work changes and new constraints appear.",
        ],
        rows: [
          ["Priority", "The roadmap bet or customer promise"],
          ["Goal", "The outcome the priority supports"],
          ["Tasks", "The work required to deliver it"],
          ["Status", "What is done, blocked, late, or ready for review"],
        ],
        promptTitle: "Roadmap review",
        prompt:
          "Review this roadmap area. Show active work, blocked tasks, capacity risks, launch dependencies, and decisions that need approval.",
      },
      {
        id: "capacity",
        title: "Plan roadmaps around capacity",
        paragraphs: [
          "The roadmap can look reasonable until it meets real team capacity. FortyOne helps teams compare current workload, estimates, deadlines, and priorities before adding new work.",
          "That makes it easier to say yes, no, or later with the delivery impact visible.",
        ],
        cards: [
          {
            heading: "Capacity view",
            subheading: "Teams before another commitment",
            rows: [
              { label: "Engineering", value: "79%", width: "79%" },
              { label: "Support", value: "63%", width: "63%" },
            ],
          },
          {
            heading: "Tradeoff",
            subheading: "AI prepares options",
            badge: "Decision",
            rows: [
              { label: "Move now", value: "Activation workstream" },
              { label: "Move later", value: "Reporting refresh" },
            ],
          },
        ],
      },
      {
        id: "launches",
        title: "Keep launch work visible across teams",
        paragraphs: [
          "Most roadmap items depend on multiple teams. FortyOne keeps product, engineering, marketing, support, and operations work visible in the same execution flow.",
          "Teams can see which handoffs are late, which owners are blocked, and what decisions must happen before the launch moves.",
        ],
        rows: [
          [
            "Engineering",
            "Implementation, technical review, and delivery risk",
          ],
          ["Product", "Scope, priority, feedback, and roadmap decisions"],
          ["Marketing", "Launch assets, approvals, and campaign tasks"],
          ["Support", "Customer updates, readiness tasks, and escalations"],
        ],
      },
    ],
    questions: [
      [
        "Does FortyOne replace roadmap tools?",
        "FortyOne focuses on keeping roadmap priorities connected to execution, owners, capacity, and status.",
      ],
      [
        "Can AI help prepare roadmap reviews?",
        "Yes. AI can gather active work, blockers, estimates, owners, and decisions into a review-ready summary.",
      ],
      [
        "Can roadmap items connect to goals?",
        "Yes. Roadmap priorities can be connected to goals so teams understand the outcome behind the work.",
      ],
    ],
  },
  {
    slug: "integrations",
    label: featureLabels.integrations,
    heroTitle: "FortyOne Integrations",
    metaTitle: "Integrations | Turn Tools into AI Context | FortyOne",
    metaDescription:
      "See how FortyOne uses integrations with Slack, GitHub, Google Calendar, Figma, GitLab, and other tools to enrich project planning.",
    intro: [
      "Integrations matter because project work rarely starts inside the project management tool. It starts in chat, code, calendars, design files, support conversations, and documents.",
      "FortyOne brings that context back into the project plan so AI and teams can make better decisions about tasks, owners, estimates, timing, and risk.",
      "The goal is not to replace every tool. It is to make the tools your team already uses useful inside the workflow where work is planned and tracked.",
    ],
    benefits: [
      [
        "Less manual copying",
        "Source context can stay attached to tasks instead of being rewritten into status updates.",
      ],
      [
        "Better AI recommendations",
        "AI can use connected context when preparing assignments, estimates, and schedule suggestions.",
      ],
      [
        "Cleaner intake",
        "Slack and other tools can become paths for turning requests into structured project work.",
      ],
      [
        "Stronger delivery visibility",
        "Code, design, calendar, and conversation context can inform planning and review.",
      ],
    ],
    previewCards: [
      {
        heading: "Slack request",
        subheading: "Conversation converted to work",
        badge: "New",
        rows: [
          { label: "Task", value: "Confirm onboarding timeline" },
          { label: "Source", value: "#customer-launch" },
        ],
      },
      {
        heading: "GitHub context",
        subheading: "Delivery signal attached",
        rows: [
          { label: "Issue", value: "#142 webhook retry policy" },
          { label: "Status", value: "Ready for API review" },
        ],
      },
    ],
    sections: [
      {
        id: "context",
        title: "Bring source context into the project plan",
        paragraphs: [
          "When context lives outside the plan, teams spend time chasing links and rebuilding what happened. FortyOne helps keep the source connected to the task it created.",
          "That gives owners and managers the context they need before work starts, changes, or gets reviewed.",
        ],
        rows: [
          [
            "Slack",
            "Requests, decisions, and follow-up work from conversations",
          ],
          ["GitHub", "Issues, implementation details, and delivery context"],
          ["Calendar", "Availability and timing signals for planning"],
          [
            "Design and docs",
            "Source material connected to the work it supports",
          ],
        ],
        promptTitle: "Context intake",
        prompt:
          "Create a task from this source context, attach relevant links, summarize the request, suggest an owner, estimate the work, and show what needs review.",
      },
      {
        id: "ai",
        title: "Use integrations as AI planning context",
        paragraphs: [
          "AI recommendations improve when the system can see the real signals around the work. FortyOne can use connected context to prepare better owner, estimate, timing, and risk suggestions.",
          "Teams still review the recommendation, but the first draft starts with more of the project reality included.",
        ],
        cards: [
          {
            heading: "Planning inputs",
            subheading: "Context gathered before recommendation",
            rows: [
              { label: "Calendar", value: "Owner available Thursday" },
              { label: "GitHub", value: "API issue linked" },
            ],
          },
          {
            heading: "AI recommendation",
            subheading: "Prepared from connected tools",
            badge: "Review",
            rows: [
              { label: "Owner", value: "Platform team" },
              { label: "Estimate", value: "3 days" },
            ],
          },
        ],
      },
      {
        id: "workflow",
        title: "Keep existing tools in the workflow",
        paragraphs: [
          "Teams should not have to abandon the tools they already use to get better planning. FortyOne connects those tools to the project workflow where decisions happen.",
          "That keeps conversations, source links, delivery context, and availability available when managers and AI prepare the next step.",
        ],
        rows: [
          ["Create", "Turn source context into structured work"],
          ["Attach", "Keep links and summaries with the task"],
          ["Plan", "Use context for owner, estimate, and timing suggestions"],
          ["Review", "Approve changes before they affect the project"],
        ],
      },
    ],
    questions: [
      [
        "Which integrations matter most?",
        "Slack, GitHub, Google Calendar, Figma, GitLab, and documents are important because they hold the context teams use to plan work.",
      ],
      [
        "Do integrations replace existing tools?",
        "No. Integrations bring context from existing tools into FortyOne so planning and task management improve.",
      ],
      [
        "Can AI use integration context?",
        "Yes. Connected context can help AI prepare better owner, estimate, timing, and risk recommendations.",
      ],
    ],
  },
];

export const getFeatureBySlug = (slug: string) => {
  return features.find((feature) => feature.slug === slug);
};
