import { useCaseLabels } from "./use-case-links";

export type UseCaseVisualRow = {
  label: string;
  value: string;
  width?: string;
};

export type UseCaseVisual = {
  badge?: string;
  heading: string;
  note?: string;
  rows: UseCaseVisualRow[];
  subheading: string;
};

export type UseCaseSection = {
  cards?: UseCaseVisual[];
  id: string;
  paragraphs: string[];
  prompt?: string;
  promptTitle?: string;
  rows?: [string, string][];
  tableHead?: [string, string];
  title: string;
};

export type UseCase = {
  benefits: [string, string][];
  heroTitle: string;
  intro: string[];
  label: string;
  metaDescription: string;
  metaTitle: string;
  previewCards: UseCaseVisual[];
  questions: [string, string][];
  sections: UseCaseSection[];
  slug: string;
};

export const useCases: UseCase[] = [
  {
    slug: "operations",
    label: useCaseLabels.operations,
    heroTitle: "FortyOne for Operations",
    metaTitle: "FortyOne for Operations | AI Project Management Use Case",
    metaDescription:
      "See how operations teams use FortyOne to turn requests, capacity, assignments, and project reviews into one AI-assisted project management workflow.",
    intro: [
      "Operations work usually starts with scattered context: a Slack thread, a dashboard screenshot, a calendar constraint, and unanswered questions about who should own the next step.",
      "FortyOne turns that context into a working project plan. It can create the task, suggest the right owner, fill in an estimate, plan around real capacity, and stop for review before important AI actions are applied.",
      "That means operations leaders spend less time rebuilding status from tools and more time deciding what should move, what should wait, and where the team needs help.",
    ],
    benefits: [
      [
        "Less manual status work",
        "Weekly reviews can start from live project data instead of copied notes, spreadsheets, and repeated follow-ups.",
      ],
      [
        "Better assignment decisions",
        "AI can compare workload, context, estimates, and timing before recommending an owner.",
      ],
      [
        "Cleaner meeting prep",
        "Blocked work, overdue tasks, review notes, and decisions are collected before the meeting starts.",
      ],
      [
        "Controlled AI execution",
        "Important owner, estimate, and schedule changes can stay staged until a manager approves them.",
      ],
    ],
    previewCards: [
      {
        heading: "Weekly operations review",
        subheading: "Prepared from live projects",
        badge: "Ready",
        rows: [
          { label: "Blocked work", value: "4 tasks need owner decisions" },
          { label: "Capacity", value: "Product can take one more request" },
        ],
      },
      {
        heading: "Slack request",
        subheading: "Turn a loose request into work",
        badge: "AI draft",
        rows: [
          { label: "Task", value: "Prepare onboarding report" },
          { label: "Owner", value: "Operations" },
        ],
      },
    ],
    sections: [
      {
        id: "weekly-ops-packet",
        title: "Weekly ops packet without the manual chase",
        paragraphs: [
          "Every operations meeting needs the same basic answer: what moved, what is blocked, what needs a decision, and who has enough room to take the next piece of work.",
          "FortyOne keeps that packet connected to the actual project plan, so the review is not rebuilt from scratch in slides, spreadsheets, and status messages.",
        ],
        rows: [
          [
            "Project status",
            "At-risk tasks, blocked owners, and work due this week",
          ],
          [
            "Capacity",
            "Which teams can take new work and which teams are near limit",
          ],
          ["Assignments", "Suggested owners, estimates, and schedule windows"],
          [
            "Review note",
            "What changed, what needs approval, and what can wait",
          ],
        ],
        promptTitle: "Weekly operations packet",
        prompt:
          "Build this week's operations packet from active projects, overdue tasks, blocked work, and team capacity. Include source links and stop before changing any assignments.",
      },
      {
        id: "task-intake",
        title: "Turn requests into assigned work",
        paragraphs: [
          "Requests often arrive before they are ready to assign. FortyOne can turn rough notes into a structured task, connect it to a goal, and recommend who should own it based on workload and context.",
          "When Slack integration is enabled, teams can create work from the place where requests already happen, then let AI prepare the owner, estimate, and timing for review.",
        ],
        rows: [
          [
            "Slack request",
            "Create a task, attach context, and suggest an owner",
          ],
          ["Calendar signal", "Find a realistic work window before assignment"],
          [
            "Missing estimate",
            "Fill an initial estimate from similar prior work",
          ],
          ["Manager review", "Stop before applying important AI changes"],
        ],
        promptTitle: "Task intake",
        prompt:
          "Create a task from this Slack request, attach the relevant goal, suggest the best owner, estimate the work, and choose a schedule window that does not overload the team.",
      },
      {
        id: "capacity",
        title: "Plan around real team capacity",
        paragraphs: [
          "The best assignment is not just the person who can do the work. It is the person who can do it at the right time without creating a hidden bottleneck somewhere else.",
          "FortyOne gives operations leads a faster way to compare team load, estimates, schedule windows, and ownership before work is moved.",
        ],
        cards: [
          {
            heading: "Capacity check",
            subheading: "Owners, estimates, and timing reviewed together",
            rows: [
              { label: "Operations", value: "68%", width: "68%" },
              { label: "Product", value: "42%", width: "42%" },
            ],
            note: "AI recommends Product for the next operations request.",
          },
          {
            heading: "Review before apply",
            subheading: "Every AI change stays editable before it moves work",
            rows: [
              { label: "Owner", value: "Assign to Maya Chen" },
              { label: "Estimate", value: "Set initial estimate to 3 days" },
              { label: "Start", value: "Schedule for Thursday morning" },
            ],
          },
        ],
        promptTitle: "Review before apply",
        prompt:
          "Review the AI plan for this workstream. Show the proposed owner, estimate, start time, and task changes. Let me edit or approve the changes before anything moves.",
      },
    ],
    questions: [
      [
        "Can FortyOne create tasks from Slack?",
        "Yes. Slack can become an intake path for project work, with AI helping turn conversations into structured tasks.",
      ],
      [
        "Can AI assign work automatically?",
        "AI can suggest owners, estimates, and timing. Teams can keep review controls in place before important changes are applied.",
      ],
      [
        "Does this replace project managers?",
        "No. It removes coordination drag so managers can spend more time deciding priorities and less time rebuilding status.",
      ],
    ],
  },
  {
    slug: "product",
    label: useCaseLabels.product,
    heroTitle: "FortyOne for Product Management",
    metaTitle:
      "FortyOne for Product Management | AI Project Management Use Case",
    metaDescription:
      "See how product teams use FortyOne to turn goals, feedback, roadmap decisions, and launch work into AI-assisted execution.",
    intro: [
      "Product work moves across ideas, customer feedback, roadmap bets, launch deadlines, and engineering capacity. The hard part is keeping that work connected after decisions leave the planning meeting.",
      "FortyOne helps product teams turn goals into clear work, attach the context behind each decision, and use AI to suggest ownership, estimates, and schedule windows before tasks move.",
      "The result is a roadmap that is easier to execute because priorities, tasks, handoffs, and status all live in the same system.",
    ],
    benefits: [
      [
        "Roadmap work stays connected",
        "Goals, customer requests, tasks, and launch milestones remain linked as the plan changes.",
      ],
      [
        "Faster intake decisions",
        "AI can summarize requests, identify missing details, and prepare a scoped task for review.",
      ],
      [
        "Cleaner product to engineering handoff",
        "Tasks carry the goal, context, estimate, and source links that engineering needs before starting.",
      ],
      [
        "Better launch visibility",
        "Product leads can see which work is blocked, which handoffs are late, and what needs a decision.",
      ],
    ],
    previewCards: [
      {
        heading: "Roadmap intake",
        subheading: "Customer signal linked to a goal",
        badge: "Draft",
        rows: [
          { label: "Signal", value: "Enterprise onboarding request" },
          { label: "Goal", value: "Improve activation" },
        ],
      },
      {
        heading: "Launch readiness",
        subheading: "Cross-functional work in one view",
        badge: "At risk",
        rows: [
          { label: "Blocked", value: "Billing copy review" },
          { label: "Owner", value: "Product marketing" },
        ],
      },
    ],
    sections: [
      {
        id: "feedback-to-work",
        title: "Turn feedback into work the team can execute",
        paragraphs: [
          "Feedback is only useful when it becomes a clear decision and a clear next step. FortyOne helps product teams move from a customer request or internal idea to a task with context, owner, estimate, and goal.",
          "Instead of losing source material in calls and chat threads, product teams can keep the reason for the work attached to the work itself.",
        ],
        rows: [
          [
            "Customer request",
            "Summarize the ask and connect it to a product goal",
          ],
          [
            "Internal idea",
            "Create a scoped task with missing details highlighted",
          ],
          [
            "Roadmap decision",
            "Link follow-up work to the priority it supports",
          ],
          ["Review", "Stage owner, estimate, and timing before assignment"],
        ],
        promptTitle: "Product intake",
        prompt:
          "Turn this customer request into a product task, connect it to the activation goal, list missing details, suggest an owner, and prepare it for review.",
      },
      {
        id: "launch-planning",
        title: "Keep launch work visible across every team",
        paragraphs: [
          "A product launch usually depends on engineering, design, marketing, support, and operations. FortyOne helps product managers see which tasks are ready, which are blocked, and what decision is needed next.",
          "AI can prepare the launch review by collecting overdue work, source links, owners, estimates, and unresolved decisions.",
        ],
        cards: [
          {
            heading: "Launch plan",
            subheading: "Tasks grouped by workstream",
            badge: "Review",
            rows: [
              { label: "Engineering", value: "API handoff due Friday" },
              { label: "Marketing", value: "Launch page copy pending" },
            ],
          },
          {
            heading: "AI recommendation",
            subheading: "Next owner and schedule window",
            rows: [
              { label: "Owner", value: "Product marketing" },
              { label: "Start", value: "Tomorrow morning" },
              { label: "Estimate", value: "2 days" },
            ],
          },
        ],
        promptTitle: "Launch review",
        prompt:
          "Prepare the launch review for this product area. Show blocked work, missing owners, overdue tasks, and the next decisions that need product approval.",
      },
      {
        id: "prioritization",
        title: "Prioritize with goals, capacity, and timing in view",
        paragraphs: [
          "Prioritization breaks down when the roadmap ignores team capacity. FortyOne helps product leaders compare priority, workload, estimates, and delivery timing before adding more work to a team.",
          "That makes it easier to say yes, no, or later with the execution impact visible.",
        ],
        rows: [
          ["Goal impact", "Show which objective the work supports"],
          ["Capacity", "Compare current owner and team load"],
          ["Estimate", "Suggest an initial estimate when none exists"],
          ["Decision", "Explain tradeoffs before work is added"],
        ],
        promptTitle: "Prioritization support",
        prompt:
          "Compare these roadmap candidates against goals, team capacity, and estimated effort. Recommend what should move now, what should wait, and why.",
      },
    ],
    questions: [
      [
        "Can FortyOne help with product intake?",
        "Yes. Product teams can turn feedback and ideas into structured tasks with context, owner suggestions, estimates, and review controls.",
      ],
      [
        "Can it connect work to goals?",
        "Yes. The platform is designed to keep tasks connected to goals so teams understand why work exists.",
      ],
      [
        "Can it help with launches?",
        "Yes. Launch tasks, blockers, owners, and next decisions can be reviewed in one workflow.",
      ],
    ],
  },
  {
    slug: "developers",
    label: useCaseLabels.developers,
    heroTitle: "FortyOne for Developers",
    metaTitle: "FortyOne for Developers | AI Project Management Use Case",
    metaDescription:
      "See how developers use FortyOne to turn technical requests, estimates, assignments, and delivery tracking into an AI-assisted workflow.",
    intro: [
      "Engineering teams do not just need a list of tasks. They need context, priorities, estimates, ownership, dependencies, and a realistic plan for when work should start.",
      "FortyOne helps turn technical requests, Slack conversations, GitHub context, and product goals into work that is ready to assign and track.",
      "AI supports the planning work around engineering: preparing estimates, suggesting owners, identifying blockers, and keeping managers in control before changes are applied.",
    ],
    benefits: [
      [
        "Cleaner technical intake",
        "Requests can become scoped tasks with source links, missing details, and owner recommendations.",
      ],
      [
        "Better estimates",
        "AI can suggest an initial estimate from similar work when the team has not provided one yet.",
      ],
      [
        "Less context switching",
        "Slack, GitHub, files, goals, and project tasks can stay connected in the same workflow.",
      ],
      [
        "More realistic planning",
        "Assignments can account for workload and timing instead of only who has the right skill.",
      ],
    ],
    previewCards: [
      {
        heading: "Technical intake",
        subheading: "Request converted into scoped work",
        badge: "AI draft",
        rows: [
          { label: "Task", value: "Add webhook retry policy" },
          { label: "Context", value: "Linked to GitHub issue" },
        ],
      },
      {
        heading: "Assignment plan",
        subheading: "Estimate and owner suggested",
        rows: [
          { label: "Owner", value: "Platform team" },
          { label: "Estimate", value: "3 days" },
        ],
      },
    ],
    sections: [
      {
        id: "technical-intake",
        title: "Turn technical requests into scoped tasks",
        paragraphs: [
          "Engineering work often starts as a rough request: a bug report, a Slack message, a support escalation, or a product handoff. FortyOne helps convert that request into a task with context attached.",
          "The team can see the source, the goal, the likely owner, the initial estimate, and the missing details before work is accepted.",
        ],
        rows: [
          ["Slack thread", "Create a task and attach the conversation"],
          ["GitHub issue", "Link implementation context and source details"],
          ["Missing estimate", "Suggest a first estimate from similar work"],
          ["Manager review", "Approve owner and timing before assignment"],
        ],
        promptTitle: "Engineering intake",
        prompt:
          "Create an engineering task from this request, attach the GitHub context, suggest the owner, estimate the effort, and list the missing details before assignment.",
      },
      {
        id: "capacity-planning",
        title: "Assign work around capacity, not just skill",
        paragraphs: [
          "The best engineer for a task may not be the best person to assign today. FortyOne helps managers compare current workload, timing, and estimated effort before assigning new work.",
          "This is useful for teams that want AI support without giving up review control.",
        ],
        cards: [
          {
            heading: "Team load",
            subheading: "Capacity before assignment",
            rows: [
              { label: "Platform", value: "82%", width: "82%" },
              { label: "Frontend", value: "54%", width: "54%" },
            ],
            note: "AI recommends Frontend for the next task.",
          },
          {
            heading: "Delivery risk",
            subheading: "Blockers and handoffs",
            badge: "Watch",
            rows: [
              { label: "Blocked", value: "API contract review" },
              { label: "Due", value: "Friday" },
            ],
          },
        ],
        promptTitle: "Assignment review",
        prompt:
          "Review this engineering task. Recommend an owner, estimate, start time, and risk notes based on current workload and linked context.",
      },
      {
        id: "delivery-tracking",
        title: "Track delivery without rebuilding status",
        paragraphs: [
          "Engineering status should not require a second system. FortyOne keeps work connected to goals, owners, estimates, and project progress so managers can see what is moving and what is stuck.",
          "AI can prepare the review by highlighting at-risk tasks, blocked owners, and handoffs that need attention.",
        ],
        rows: [
          ["Active work", "Show tasks in progress and due soon"],
          ["Blocked work", "List blockers with source context"],
          ["Handoffs", "Identify work waiting on another owner"],
          ["Review note", "Summarize what changed since the last check-in"],
        ],
        promptTitle: "Engineering status",
        prompt:
          "Prepare an engineering status review from active tasks, blocked work, estimates, and GitHub-linked context. Show what needs a decision.",
      },
    ],
    questions: [
      [
        "Can FortyOne connect to engineering tools?",
        "It can use integrations such as GitHub and Slack to keep technical context connected to the project plan.",
      ],
      [
        "Can AI estimate engineering tasks?",
        "AI can suggest an initial estimate from context and similar work, then leave it for manager or team review.",
      ],
      [
        "Can managers approve AI changes first?",
        "Yes. Important changes can be staged for review before they are applied.",
      ],
    ],
  },
  {
    slug: "customer-support",
    label: useCaseLabels["customer-support"],
    heroTitle: "FortyOne for Customer Support",
    metaTitle: "FortyOne for Customer Support | AI Project Management Use Case",
    metaDescription:
      "See how customer success teams use FortyOne to turn escalations, renewals, customer asks, and internal follow-up into accountable work.",
    intro: [
      "Customer success teams live between customer promises and internal execution. The risk is not usually the request itself. The risk is losing ownership once the request crosses into product, engineering, or operations.",
      "FortyOne helps customer teams turn escalations, renewal risks, onboarding tasks, and customer asks into project work that has an owner, context, estimate, and review path.",
      "That makes it easier to close the loop with customers because every internal follow-up is tracked against the work that needs to happen.",
    ],
    benefits: [
      [
        "Clearer escalation ownership",
        "Customer issues can become tasks with internal owners, due windows, and source context.",
      ],
      [
        "Better renewal preparation",
        "Risk items, blocked asks, and promised follow-ups can be reviewed before account meetings.",
      ],
      [
        "Faster internal follow-up",
        "AI can summarize the ask, suggest the next owner, and prepare the work for approval.",
      ],
      [
        "Less customer promise drift",
        "Teams can track whether promised actions have a real owner and are moving.",
      ],
    ],
    previewCards: [
      {
        heading: "Account escalation",
        subheading: "Customer ask converted to work",
        badge: "Urgent",
        rows: [
          { label: "Customer", value: "Acme renewal risk" },
          { label: "Owner", value: "Product support" },
        ],
      },
      {
        heading: "Follow-up plan",
        subheading: "Actions ready for review",
        rows: [
          { label: "Task", value: "Confirm SSO timeline" },
          { label: "Due", value: "Before renewal call" },
        ],
      },
    ],
    sections: [
      {
        id: "escalations",
        title: "Turn escalations into accountable work",
        paragraphs: [
          "Customer escalations often arrive with urgency but not structure. FortyOne helps convert the customer context into work that has a clear owner, timing, and next action.",
          "The customer team can keep the original request attached, while internal teams see the project context they need to respond.",
        ],
        rows: [
          ["Customer note", "Summarize the issue and attach source context"],
          ["Internal task", "Create work for the right team"],
          ["Owner", "Suggest who should take the next step"],
          ["Review", "Stage the assignment before applying it"],
        ],
        promptTitle: "Escalation intake",
        prompt:
          "Turn this customer escalation into internal work. Summarize the ask, suggest the right owner, estimate the response effort, and prepare a customer follow-up note.",
      },
      {
        id: "renewal-readiness",
        title: "Prepare renewals with real project status",
        paragraphs: [
          "Renewal risk often comes from open promises: a missing feature, a blocked onboarding step, or an unresolved support handoff. FortyOne helps customer teams see which promises are moving and which need attention.",
          "AI can prepare a renewal packet with open tasks, owners, risk notes, and decisions needed before the account call.",
        ],
        cards: [
          {
            heading: "Renewal packet",
            subheading: "Open customer commitments",
            badge: "Review",
            rows: [
              { label: "Blocked", value: "SSO timeline confirmation" },
              { label: "Owner", value: "Platform team" },
            ],
          },
          {
            heading: "Risk actions",
            subheading: "Next steps before the call",
            rows: [
              { label: "Follow-up", value: "Send timeline update" },
              { label: "Start", value: "Today" },
            ],
          },
        ],
        promptTitle: "Renewal packet",
        prompt:
          "Prepare the renewal readiness packet for this account. Include open promises, blocked work, owners, due dates, and recommended follow-up actions.",
      },
      {
        id: "close-the-loop",
        title: "Close the loop with customers",
        paragraphs: [
          "The customer-facing update should match what is actually happening internally. FortyOne keeps the internal task connected to the customer ask so the account team can communicate with confidence.",
          "When work changes, the team can see what changed, who owns it, and what the customer should be told next.",
        ],
        rows: [
          ["Internal change", "Summarize what moved and what is still blocked"],
          ["Customer update", "Draft a clear follow-up note"],
          ["Owner change", "Show who owns the next step"],
          ["Next review", "Schedule a realistic check-in window"],
        ],
        promptTitle: "Customer update",
        prompt:
          "Draft a customer update from the current project status. Include what changed, what is still blocked, and when the customer should expect the next update.",
      },
    ],
    questions: [
      [
        "Can customer requests become internal tasks?",
        "Yes. Customer asks can be converted into project work with source context, owners, estimates, and review controls.",
      ],
      [
        "Can this help with renewal risk?",
        "Yes. Open promises, blockers, and follow-up tasks can be gathered into a renewal readiness view.",
      ],
      [
        "Can AI draft customer updates?",
        "AI can prepare a draft from project status, while the team remains responsible for reviewing what is sent.",
      ],
    ],
  },
  {
    slug: "field-crews",
    label: useCaseLabels["field-crews"],
    heroTitle: "FortyOne for Field Crews",
    metaTitle: "FortyOne for Field Crews | AI Project Management Use Case",
    metaDescription:
      "See how construction and field crews use FortyOne to coordinate site work, subcontractor handoffs, approvals, schedules, and AI-assisted project tracking.",
    intro: [
      "Construction and field work depends on coordination between site managers, subcontractors, suppliers, inspectors, office teams, and clients. A single missed handoff can delay the whole schedule.",
      "FortyOne helps teams turn site notes, Slack messages, inspection findings, and meeting decisions into assigned work with owners, estimates, due windows, and review controls.",
      "The platform is useful beyond software teams because it focuses on the core project-management problem: clear work, accountable owners, realistic timing, and visibility into what is blocked.",
    ],
    benefits: [
      [
        "Site work becomes trackable",
        "Field notes, daily updates, and inspection items can become assigned tasks instead of loose messages.",
      ],
      [
        "Handoffs stay visible",
        "Subcontractor, supplier, approval, and office follow-ups can be tracked against the same project plan.",
      ],
      [
        "Schedule risk is easier to see",
        "Blocked tasks, late decisions, and overloaded teams can be surfaced before they affect milestones.",
      ],
      [
        "AI supports coordination",
        "AI can summarize updates, suggest owners, estimate follow-up effort, and stage changes for manager review.",
      ],
    ],
    previewCards: [
      {
        heading: "Site update",
        subheading: "Field note converted to work",
        badge: "New",
        rows: [
          { label: "Task", value: "Resolve Level 2 access issue" },
          { label: "Owner", value: "Site supervisor" },
        ],
      },
      {
        heading: "Schedule risk",
        subheading: "Blocked handoffs in view",
        badge: "Watch",
        rows: [
          { label: "Blocked", value: "Electrical rough-in approval" },
          { label: "Due", value: "Before framing starts" },
        ],
      },
    ],
    sections: [
      {
        id: "site-work-intake",
        title: "Turn site updates into assigned work",
        paragraphs: [
          "Site updates often arrive as short notes, photos, messages, or meeting comments. FortyOne can help turn those updates into structured tasks with the right owner, timing, and context.",
          "That gives project managers a clearer way to track what needs to happen next without manually rewriting every update into a task list.",
        ],
        rows: [
          ["Field note", "Create a task and attach the original context"],
          ["Inspection item", "Assign the follow-up and due window"],
          ["Subcontractor request", "Route the work to the right owner"],
          [
            "Manager review",
            "Approve owner, estimate, and timing before changes apply",
          ],
        ],
        promptTitle: "Site work intake",
        prompt:
          "Turn this site update into project work. Create the task, suggest the owner, estimate the follow-up effort, and show anything that needs manager review.",
      },
      {
        id: "handoff-coordination",
        title: "Coordinate subcontractor and supplier handoffs",
        paragraphs: [
          "Many construction delays come from work waiting between teams. FortyOne helps track handoffs between field crews, subcontractors, suppliers, approvals, and office coordination.",
          "AI can prepare a handoff review by listing blocked work, missing owners, upcoming due dates, and schedule risks.",
        ],
        cards: [
          {
            heading: "Handoff review",
            subheading: "Work waiting between teams",
            badge: "Review",
            rows: [
              { label: "Electrical", value: "Rough-in approval pending" },
              { label: "Supplier", value: "Doors delivery date missing" },
            ],
          },
          {
            heading: "Crew capacity",
            subheading: "Workload before assignment",
            rows: [
              { label: "Site team", value: "76%", width: "76%" },
              { label: "Admin team", value: "48%", width: "48%" },
            ],
            note: "AI recommends Admin for the permit follow-up.",
          },
        ],
        promptTitle: "Handoff review",
        prompt:
          "Prepare the handoff review for this project. Show blocked subcontractor work, missing approvals, supplier dependencies, owners, and next actions.",
      },
      {
        id: "schedule-risk",
        title: "Spot schedule risk before it becomes a delay",
        paragraphs: [
          "A schedule risk is often visible before the delay is official: a missing approval, an overloaded crew, a late material decision, or an unclear owner.",
          "FortyOne helps project managers review those risks in one workflow and turn decisions into assigned follow-up tasks.",
        ],
        rows: [
          ["Blocked work", "List tasks waiting on approvals or handoffs"],
          ["Owner gap", "Show work without a clear accountable owner"],
          ["Due window", "Identify tasks that affect upcoming milestones"],
          ["Decision", "Create follow-up work from the project review"],
        ],
        promptTitle: "Schedule risk review",
        prompt:
          "Review this project for schedule risk. Highlight blocked work, missing owners, late approvals, and tasks that could affect the next milestone.",
      },
    ],
    questions: [
      [
        "Can FortyOne work for construction teams?",
        "Yes. The workflow is built around projects, tasks, owners, estimates, and review controls, which applies well to field and office coordination.",
      ],
      [
        "Can field updates become tasks?",
        "Yes. A field note or message can become structured project work with context, owner suggestions, and timing.",
      ],
      [
        "Can it help with subcontractor handoffs?",
        "Yes. Handoffs, approvals, blockers, due windows, and follow-up tasks can be tracked in one project workflow.",
      ],
    ],
  },
  {
    slug: "government",
    label: useCaseLabels.government,
    heroTitle: "FortyOne for Government Teams",
    metaTitle: "FortyOne for Government Teams | AI Project Management Use Case",
    metaDescription:
      "See how government teams use FortyOne to coordinate work across ministries, departments, agencies, programs, approvals, and locally hosted project operations.",
    intro: [
      "Government work depends on coordination across ministries, departments, agencies, contractors, field teams, and public-service programs. The risk is not only whether work exists. The risk is losing ownership, status, and decisions as work moves between offices.",
      "FortyOne helps government teams turn priorities, programs, citizen-service requests, infrastructure work, and meeting decisions into assigned project work with owners, estimates, timelines, blockers, and review controls.",
      "For public-sector environments that need more control, FortyOne can support customized deployments where the platform runs on the government's infrastructure and project data stays inside the environment they manage.",
    ],
    benefits: [
      [
        "Cross-department visibility",
        "Ministries, departments, agencies, and project teams can see ownership, status, blockers, and next actions in one workflow.",
      ],
      [
        "Clearer public-service delivery",
        "Citizen-facing commitments, internal follow-ups, and program work can become trackable tasks with accountable owners.",
      ],
      [
        "Private deployment options",
        "Government teams can discuss local or dedicated deployments that fit their infrastructure, data structure, and operating requirements.",
      ],
      [
        "Human-reviewed AI support",
        "AI can prepare plans, summaries, owners, estimates, and status packets while important changes stay staged for approval.",
      ],
    ],
    previewCards: [
      {
        heading: "Ministerial program",
        subheading: "Priorities connected to delivery work",
        badge: "Review",
        rows: [
          { label: "At risk", value: "3 department handoffs" },
          { label: "Decision", value: "Procurement approval needed" },
        ],
      },
      {
        heading: "Local deployment",
        subheading: "Hosted on controlled infrastructure",
        badge: "Private",
        rows: [
          { label: "Data", value: "Stays in government environment" },
          { label: "Structure", value: "Configured around ministry workflows" },
        ],
      },
    ],
    sections: [
      {
        id: "ministry-coordination",
        title: "Track projects across ministries and departments",
        paragraphs: [
          "Public-sector programs often cross many offices before delivery is complete. A policy priority may depend on finance, procurement, technology, infrastructure, local authorities, and field teams.",
          "FortyOne gives government leaders and program managers a clearer way to see which work is moving, which department owns the next step, and which decisions are blocking delivery.",
        ],
        rows: [
          ["Program priority", "Connect the goal to active projects and tasks"],
          [
            "Department owner",
            "Show which ministry, agency, or team owns the next action",
          ],
          [
            "Blocked work",
            "Surface approvals, handoffs, and missing decisions",
          ],
          [
            "Leadership review",
            "Prepare status, risks, and actions for review",
          ],
        ],
        promptTitle: "Government program review",
        prompt:
          "Prepare a government program review from active projects across ministries and departments. Show owners, blocked work, overdue decisions, delivery risks, and the actions that need approval.",
      },
      {
        id: "local-deployment",
        title: "Deploy around government infrastructure and data needs",
        paragraphs: [
          "Some government teams need more than a standard cloud workspace. They may need the platform deployed in a controlled environment, connected to internal systems, and configured around existing data structures.",
          "FortyOne can be positioned for customized deployments where project data remains on infrastructure controlled by the government team, with workflows shaped around ministries, departments, programs, approvals, and reporting lines.",
        ],
        cards: [
          {
            heading: "Controlled environment",
            subheading: "Deployment shaped around public-sector needs",
            badge: "Custom",
            rows: [
              {
                label: "Hosting",
                value: "Government-controlled infrastructure",
              },
              {
                label: "Data",
                value: "Project data stays inside the environment",
              },
            ],
          },
          {
            heading: "Configured workflow",
            subheading: "Departments, programs, and approvals mapped together",
            rows: [
              { label: "Ministries", value: "Custom ownership structure" },
              {
                label: "Reports",
                value: "Status views for leadership reviews",
              },
            ],
          },
        ],
        promptTitle: "Deployment planning",
        prompt:
          "Map this government's project structure into FortyOne. Show ministries, departments, programs, approval stages, reporting views, and the deployment requirements needed to keep data in the controlled environment.",
      },
      {
        id: "public-service-delivery",
        title: "Turn public-service commitments into accountable work",
        paragraphs: [
          "Government delivery depends on promises becoming action: a road repair, a digital-service update, a permit process, a grant program, a clinic improvement, or a local authority follow-up.",
          "FortyOne helps teams turn those commitments into assigned work with clear owners, realistic timing, status visibility, and review controls before AI-assisted changes are applied.",
        ],
        rows: [
          ["Citizen request", "Convert follow-up into assigned internal work"],
          [
            "Infrastructure item",
            "Track field work, contractors, and approvals",
          ],
          [
            "Digital service",
            "Coordinate form, system, website, and support updates",
          ],
          [
            "Meeting decision",
            "Turn leadership actions into tasks with owners",
          ],
        ],
        promptTitle: "Public-service follow-up",
        prompt:
          "Turn these public-service commitments into project work. Create tasks, suggest owners, identify approvals, estimate timing, and prepare the plan for review before anything is assigned.",
      },
    ],
    questions: [
      [
        "Can FortyOne be used by government teams?",
        "Yes. Government teams can use FortyOne to coordinate projects, programs, approvals, owners, blockers, and reporting across departments and agencies.",
      ],
      [
        "Can the platform be deployed on government infrastructure?",
        "FortyOne can support conversations about customized deployments where the platform is configured around government infrastructure, data structures, and operating requirements.",
      ],
      [
        "Does AI make decisions automatically?",
        "No. AI can prepare summaries, plans, owners, estimates, and status packets, while important changes can stay staged for human review before they are applied.",
      ],
    ],
  },
  {
    slug: "marketing",
    label: useCaseLabels.marketing,
    heroTitle: "FortyOne for Marketing Campaigns",
    metaTitle:
      "FortyOne for Marketing Campaigns | AI Project Management Use Case",
    metaDescription:
      "See how marketing teams use FortyOne to plan campaigns, assign launch work, coordinate approvals, and track execution with AI assistance.",
    intro: [
      "Marketing work has a lot of moving parts: campaign briefs, content, design, landing pages, launches, approvals, and follow-up tasks after performance data arrives.",
      "FortyOne helps marketing teams turn campaign plans into assigned work, keep dependencies visible, and use AI to prepare the next action without losing review control.",
      "Campaign managers get a clearer view of what is ready, what is blocked, and what needs approval before launch dates slip.",
    ],
    benefits: [
      [
        "Campaign plans become work",
        "Briefs, assets, approvals, and launch tasks can be tracked in one project workflow.",
      ],
      [
        "Fewer missed handoffs",
        "Dependencies between copy, design, web, and product teams stay visible.",
      ],
      [
        "Smarter task assignment",
        "AI can suggest owners and timing based on team load and launch deadlines.",
      ],
      [
        "Better post-launch follow-up",
        "Performance reviews can turn into concrete next actions instead of loose notes.",
      ],
    ],
    previewCards: [
      {
        heading: "Campaign launch",
        subheading: "Workstreams ready to review",
        badge: "Launch",
        rows: [
          { label: "Landing page", value: "Copy review needed" },
          { label: "Email", value: "Design in progress" },
        ],
      },
      {
        heading: "Asset handoff",
        subheading: "Owner and timing suggested",
        rows: [
          { label: "Owner", value: "Content team" },
          { label: "Start", value: "Tomorrow" },
        ],
      },
    ],
    sections: [
      {
        id: "campaign-planning",
        title: "Turn campaign briefs into assigned work",
        paragraphs: [
          "Campaign briefs are only useful when they become a clear execution plan. FortyOne helps teams break the brief into tasks, connect those tasks to goals, and assign work around actual capacity.",
          "AI can prepare the first version of the plan, then let the campaign lead edit owners, estimates, and timing before anything moves.",
        ],
        rows: [
          ["Campaign brief", "Create tasks for copy, design, web, and launch"],
          ["Goal", "Connect the work to pipeline or activation targets"],
          ["Owner", "Suggest the right team or person for each task"],
          ["Review", "Stage the plan for approval before assignment"],
        ],
        promptTitle: "Campaign plan",
        prompt:
          "Turn this campaign brief into a project plan. Create tasks for copy, design, web, and launch, suggest owners, estimate effort, and stop for review.",
      },
      {
        id: "approval-flow",
        title: "Keep approvals and handoffs visible",
        paragraphs: [
          "Marketing timelines slip when approvals are unclear. FortyOne helps teams see which assets are ready, which are waiting on review, and who owns the next step.",
          "This gives campaign managers a single place to track the work behind a launch instead of chasing updates across messages and documents.",
        ],
        cards: [
          {
            heading: "Approval queue",
            subheading: "Assets waiting on review",
            badge: "2 due",
            rows: [
              { label: "Landing page", value: "Product review" },
              { label: "Email", value: "Design approval" },
            ],
          },
          {
            heading: "Capacity check",
            subheading: "Campaign workload",
            rows: [
              { label: "Content", value: "74%", width: "74%" },
              { label: "Design", value: "61%", width: "61%" },
            ],
            note: "AI recommends Content for the next asset update.",
          },
        ],
        promptTitle: "Approval review",
        prompt:
          "Prepare the campaign approval review. Show which assets are waiting, who owns each next step, what is blocked, and what should be approved today.",
      },
      {
        id: "post-launch",
        title: "Turn launch learning into follow-up work",
        paragraphs: [
          "After launch, teams often collect performance notes without turning them into action. FortyOne helps translate results, feedback, and open decisions into follow-up tasks.",
          "That keeps campaign learning connected to future work instead of disappearing into a retrospective document.",
        ],
        rows: [
          ["Performance note", "Summarize what changed after launch"],
          ["Follow-up task", "Create work for the next iteration"],
          ["Owner", "Suggest who should handle the action"],
          ["Timing", "Schedule follow-up without overloading the team"],
        ],
        promptTitle: "Post-launch follow-up",
        prompt:
          "Review this campaign performance note and create follow-up tasks. Suggest owners, estimates, and timing for the next iteration.",
      },
    ],
    questions: [
      [
        "Can FortyOne manage campaign launches?",
        "Yes. Campaign tasks, owners, approvals, blockers, and launch follow-up can be managed as project work.",
      ],
      [
        "Can AI help create the campaign plan?",
        "AI can draft tasks, owners, estimates, and timing from a campaign brief for team review.",
      ],
      [
        "Can it track approvals?",
        "Yes. Approval tasks can stay visible alongside the rest of the campaign plan.",
      ],
    ],
  },
  {
    slug: "leadership",
    label: useCaseLabels.leadership,
    heroTitle: "FortyOne for Executive Leadership",
    metaTitle:
      "FortyOne for Executive Leadership | AI Project Management Use Case",
    metaDescription:
      "See how leadership teams use FortyOne to connect company priorities to execution, status, capacity, and AI-assisted decision support.",
    intro: [
      "Leadership teams need to know whether the company is making progress against the work that matters most. That requires more than status updates. It requires priorities, execution, ownership, and risk in one place.",
      "FortyOne helps leaders connect goals to tasks, see which work is moving, and use AI to prepare weekly review packets without rebuilding reports manually.",
      "The value is sharper operating rhythm: fewer vague updates, clearer blockers, and better decisions about what should move next.",
    ],
    benefits: [
      [
        "Priorities stay connected to execution",
        "Company goals can stay linked to the work that teams are actually doing.",
      ],
      [
        "Weekly reviews take less preparation",
        "AI can prepare status, blockers, owners, and next decisions from active projects.",
      ],
      [
        "Capacity is visible",
        "Leaders can see where teams are overloaded before adding more work.",
      ],
      [
        "Decisions become action",
        "Meeting outcomes can be turned into assigned tasks with owners, estimates, and review controls.",
      ],
    ],
    previewCards: [
      {
        heading: "Leadership review",
        subheading: "Company priorities and risks",
        badge: "Ready",
        rows: [
          { label: "At risk", value: "2 company priorities" },
          { label: "Blocked", value: "5 owner decisions" },
        ],
      },
      {
        heading: "Decision to task",
        subheading: "Meeting action prepared",
        rows: [
          { label: "Decision", value: "Move onboarding work up" },
          { label: "Owner", value: "Product lead" },
        ],
      },
    ],
    sections: [
      {
        id: "priorities-to-work",
        title: "Connect company priorities to actual work",
        paragraphs: [
          "Leadership alignment breaks when priorities live in one place and execution lives somewhere else. FortyOne keeps goals, projects, and tasks connected so leaders can see whether priorities are becoming work.",
          "That gives leadership a clearer way to inspect progress without asking every team to rebuild the same update.",
        ],
        rows: [
          ["Company priority", "Show linked projects and active tasks"],
          ["Owner", "Identify who owns each piece of work"],
          ["Progress", "Highlight what moved and what is stuck"],
          ["Risk", "Surface blockers that need leadership decisions"],
        ],
        promptTitle: "Priority review",
        prompt:
          "Prepare a leadership review for these company priorities. Show linked work, owners, progress, blockers, and decisions needed this week.",
      },
      {
        id: "operating-rhythm",
        title: "Create a sharper weekly operating rhythm",
        paragraphs: [
          "Weekly leadership meetings should focus on decisions, not status collection. FortyOne can prepare the packet from live project data and highlight the work that needs attention.",
          "Leaders can use the review to decide what should move, what should pause, and which teams need help.",
        ],
        cards: [
          {
            heading: "Weekly packet",
            subheading: "Prepared from active work",
            badge: "Ready",
            rows: [
              { label: "Blocked", value: "5 tasks need decisions" },
              { label: "Capacity", value: "Engineering near limit" },
            ],
          },
          {
            heading: "Decision queue",
            subheading: "Actions ready to assign",
            rows: [
              { label: "Action", value: "Pause reporting refresh" },
              { label: "Owner", value: "Operations lead" },
            ],
          },
        ],
        promptTitle: "Weekly leadership packet",
        prompt:
          "Build the weekly leadership packet. Include progress against priorities, blocked work, capacity risks, and the decisions leaders need to make.",
      },
      {
        id: "capacity-decisions",
        title: "Make capacity-aware decisions",
        paragraphs: [
          "Leadership decisions often add work before the impact is visible. FortyOne helps show team load, estimated effort, and timing before new priorities are assigned.",
          "This makes tradeoffs easier to discuss because the cost of saying yes is visible.",
        ],
        rows: [
          ["New priority", "Estimate work and likely teams involved"],
          ["Team load", "Compare current capacity before assignment"],
          ["Timing", "Suggest a realistic start window"],
          ["Tradeoff", "Show what may need to pause"],
        ],
        promptTitle: "Capacity decision",
        prompt:
          "Evaluate this new priority against current team capacity. Recommend when it should start, who should own it, and what work may need to move.",
      },
    ],
    questions: [
      [
        "Can leaders see work connected to goals?",
        "Yes. FortyOne is designed to keep goals and tasks connected so leadership can inspect execution against priorities.",
      ],
      [
        "Can AI prepare leadership updates?",
        "AI can prepare review packets from active work, blockers, capacity, and next decisions.",
      ],
      [
        "Can leadership actions become tasks?",
        "Yes. Decisions can become assigned work with owners, estimates, timing, and review controls.",
      ],
    ],
  },
];

export const getUseCaseBySlug = (slug: string) => {
  return useCases.find((useCase) => useCase.slug === slug);
};
