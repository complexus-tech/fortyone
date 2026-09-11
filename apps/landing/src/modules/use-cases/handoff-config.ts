import type { useCaseLinks } from "@/lib/use-case-links";
import type { HandoffConfig } from "./handoff-types";

type UseCaseSlug = (typeof useCaseLinks)[number]["slug"];

// Illustrative work examples; source page copy remains in lib/use-cases.ts.
const HANDOFF_CONFIGS = {
  operations: {
    image: "tasks",
    workflows: [
      {
        id: "weekly-ops-packet",
        label: "Weekly review",
        caption: "A clear brief, with the work behind it.",
        preview: {
          kind: "review",
          toolbar: "Weekly operations packet",
          status: "Draft",
          eyebrow: "OPERATIONS / WEEKLY REVIEW",
          title: "The week, at a glance.",
          description: "The work that moved. The decisions still ahead.",
          rows: [
            {
              label: "Prepare onboarding report",
              value: "Due this week",
              icon: "clock",
              tone: "sky",
            },
            {
              label: "Resolve API handoff",
              value: "Blocked",
              icon: "link",
              tone: "amber",
            },
            {
              label: "Review team capacity",
              value: "Ready",
              icon: "team",
              tone: "sage",
            },
          ],
          decision: {
            title: "Who has room for the next request?",
            description: "Compare workload before moving another assignment.",
          },
          sources: ["Slack", "GitHub", "Google Calendar"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
      {
        id: "task-intake",
        label: "Request intake",
        caption: "The request becomes a task. Its context comes with it.",
        preview: {
          kind: "intake",
          source: {
            icon: "Slack",
            label: "#operations",
            badge: "Request",
            author: "Joseph",
            authorIcon: "joseph",
            message: "Can someone prepare the onboarding report for Thursday?",
          },
          task: {
            identifier: "OPS-42",
            title: "Prepare onboarding report",
            description: "Customer context and source request attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Joseph",
                icon: "joseph",
              },
              {
                label: "Estimate",
                value: "3 days",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original conversation",
        },
      },
      {
        id: "capacity",
        label: "Capacity planning",
        caption: "See the tradeoff before you make the assignment.",
        preview: {
          kind: "capacity",
          toolbar: "Team capacity",
          allocations: [
            {
              team: "Operations",
              percent: 68,
            },
            {
              team: "Product",
              percent: 42,
            },
          ],
          recommendation: "More room for the next request.",
          startDay: "Thu",
          proposal: "Product · 3 days · Thursday morning",
        },
      },
    ],
  },
  product: {
    image: "goals",
    workflows: [
      {
        id: "feedback-to-work",
        label: "Product intake",
        caption: "From customer signal to scoped product work.",
        preview: {
          kind: "intake",
          source: {
            icon: "feedback",
            label: "Customer feedback",
            badge: "Request",
            author: "Enterprise onboarding",
            authorIcon: "feedback",
            message:
              "We need a clearer setup checklist before inviting the whole team.",
          },
          task: {
            identifier: "PROD-24",
            title: "Improve onboarding guidance",
            description: "Customer evidence linked to the activation goal.",
            fields: [
              {
                label: "Suggested owner",
                value: "Product team",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "2 days",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original customer request",
        },
      },
      {
        id: "launch-planning",
        label: "Launch review",
        caption: "Every workstream, with its next decision attached.",
        preview: {
          kind: "review",
          toolbar: "Launch readiness",
          status: "Draft",
          eyebrow: "PRODUCT / LAUNCH REVIEW",
          title: "The launch, in view.",
          description: "Cross-functional work and the decisions still ahead.",
          rows: [
            {
              label: "API handoff",
              value: "Due Friday",
              icon: "link",
              tone: "sky",
            },
            {
              label: "Launch page copy",
              value: "Needs review",
              icon: "document",
              tone: "amber",
            },
            {
              label: "Support readiness",
              value: "Ready",
              icon: "check",
              tone: "sage",
            },
          ],
          decision: {
            title: "Unblock the launch page review.",
            description: "Product marketing owns the next handoff.",
          },
          sources: ["GitHub", "Figma", "Slack"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
      {
        id: "prioritization",
        label: "Prioritization",
        caption: "See the delivery impact before adding another priority.",
        preview: {
          kind: "capacity",
          toolbar: "Product capacity",
          allocations: [
            {
              team: "Engineering",
              percent: 82,
            },
            {
              team: "Product",
              percent: 42,
            },
          ],
          recommendation: "Room to scope the next activation improvement.",
          startDay: "Thu",
          proposal: "Product · 2 days · Thursday morning",
        },
      },
    ],
  },
  developers: {
    image: "tasks",
    workflows: [
      {
        id: "technical-intake",
        label: "Technical intake",
        caption: "The issue and implementation context move together.",
        preview: {
          kind: "intake",
          source: {
            icon: "GitHub",
            label: "Webhook delivery",
            badge: "Issue",
            author: "Retry policy",
            authorIcon: "document",
            message:
              "Failed webhook deliveries need a retry policy with backoff.",
          },
          task: {
            identifier: "ENG-128",
            title: "Add webhook retry policy",
            description: "GitHub issue and acceptance criteria attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Platform team",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "3 days",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original GitHub issue",
        },
      },
      {
        id: "capacity-planning",
        label: "Team capacity",
        caption:
          "Balance the next assignment against the work already planned.",
        preview: {
          kind: "capacity",
          toolbar: "Engineering capacity",
          allocations: [
            {
              team: "Platform",
              percent: 82,
            },
            {
              team: "Frontend",
              percent: 54,
            },
          ],
          recommendation: "More room for the next frontend task.",
          startDay: "Thu",
          proposal: "Frontend · 3 days · Thursday morning",
        },
      },
      {
        id: "delivery-tracking",
        label: "Delivery review",
        caption: "Progress and blockers, connected to the implementation.",
        preview: {
          kind: "review",
          toolbar: "Engineering review",
          status: "Draft",
          eyebrow: "ENGINEERING / DELIVERY",
          title: "The work behind the release.",
          description: "Active tasks, dependencies, and the next handoff.",
          rows: [
            {
              label: "Webhook retry policy",
              value: "In progress",
              icon: "document",
              tone: "sky",
            },
            {
              label: "API contract review",
              value: "Blocked",
              icon: "link",
              tone: "amber",
            },
            {
              label: "Frontend handoff",
              value: "Ready",
              icon: "check",
              tone: "sage",
            },
          ],
          decision: {
            title: "Resolve the API contract first.",
            description:
              "Keep the dependent work visible before assigning more.",
          },
          sources: ["GitHub", "Slack", "Google Calendar"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
    ],
  },
  "customer-support": {
    image: "customer-support",
    workflows: [
      {
        id: "escalations",
        label: "Escalation intake",
        caption: "A customer promise becomes accountable internal work.",
        preview: {
          kind: "intake",
          source: {
            icon: "feedback",
            label: "Customer escalation",
            badge: "Urgent",
            author: "Acme renewal",
            authorIcon: "feedback",
            message:
              "Can you confirm the SSO timeline before our renewal call?",
          },
          task: {
            identifier: "CS-38",
            title: "Confirm the SSO timeline",
            description: "Customer commitment and account context attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Platform team",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "1 day",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original customer request",
        },
      },
      {
        id: "renewal-readiness",
        label: "Renewal review",
        caption: "Walk into the account call with the real delivery picture.",
        preview: {
          kind: "review",
          toolbar: "Renewal readiness",
          status: "Draft",
          eyebrow: "CUSTOMER SUCCESS / ACME",
          title: "The promises still open.",
          description:
            "Owners, blockers, and follow-up before the account call.",
          rows: [
            {
              label: "SSO timeline",
              value: "Blocked",
              icon: "link",
              tone: "amber",
            },
            {
              label: "Onboarding follow-up",
              value: "In progress",
              icon: "clock",
              tone: "sky",
            },
            {
              label: "Account call brief",
              value: "Ready",
              icon: "document",
              tone: "sage",
            },
          ],
          decision: {
            title: "Confirm the delivery window.",
            description:
              "Ask Platform for a reviewed timeline before the call.",
          },
          sources: ["Slack", "Google Calendar", "Outlook Calendar"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
      {
        id: "close-the-loop",
        label: "Customer follow-up",
        caption: "A reviewed delivery update becomes the next customer action.",
        preview: {
          kind: "intake",
          source: {
            icon: "Slack",
            label: "#customer-success",
            badge: "Update",
            author: "Delivery handoff",
            authorIcon: "check",
            message:
              "The SSO timeline has been reviewed. Prepare the customer follow-up.",
          },
          task: {
            identifier: "CS-39",
            title: "Prepare the timeline update",
            description:
              "Internal delivery status stays linked to the customer ask.",
            fields: [
              {
                label: "Suggested owner",
                value: "Account team",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "1 hour",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Linked delivery update",
        },
      },
    ],
  },
  "field-crews": {
    image: "construction",
    workflows: [
      {
        id: "site-work-intake",
        label: "Site intake",
        caption: "The field note follows the task from site to office.",
        preview: {
          kind: "intake",
          source: {
            icon: "document",
            label: "Site update",
            badge: "New",
            author: "Level 2 access",
            authorIcon: "document",
            message:
              "The access route is blocked. Please arrange a supervisor review.",
          },
          task: {
            identifier: "SITE-42",
            title: "Resolve Level 2 access issue",
            description: "Site note and inspection context attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Site supervisor",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "1 day",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original field note",
        },
      },
      {
        id: "handoff-coordination",
        label: "Crew handoffs",
        caption: "Plan the next handoff around the people who can move it.",
        preview: {
          kind: "capacity",
          toolbar: "Crew capacity",
          allocations: [
            {
              team: "Site team",
              percent: 76,
            },
            {
              team: "Admin team",
              percent: 48,
            },
          ],
          recommendation: "Room for the permit follow-up with Admin.",
          startDay: "Thu",
          proposal: "Admin · 1 day · Thursday morning",
        },
      },
      {
        id: "schedule-risk",
        label: "Schedule review",
        caption: "Spot the missing decision before it delays the milestone.",
        preview: {
          kind: "review",
          toolbar: "Site handoff review",
          status: "Draft",
          eyebrow: "FIELD CREWS / DELIVERY",
          title: "Before the next milestone.",
          description: "Approvals, suppliers, and crew handoffs in one brief.",
          rows: [
            {
              label: "Electrical rough-in",
              value: "Approval needed",
              icon: "link",
              tone: "amber",
            },
            {
              label: "Doors delivery",
              value: "Date needed",
              icon: "clock",
              tone: "sky",
            },
            {
              label: "Permit follow-up",
              value: "Owner ready",
              icon: "team",
              tone: "sage",
            },
          ],
          decision: {
            title: "Confirm the rough-in approval.",
            description: "Resolve the dependency before framing starts.",
          },
          sources: ["document", "Google Calendar", "Slack"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
    ],
  },
  government: {
    image: "roadmaps",
    workflows: [
      {
        id: "ministry-coordination",
        label: "Program review",
        caption: "The next department and decision stay visible.",
        preview: {
          kind: "review",
          toolbar: "Government program review",
          status: "Draft",
          eyebrow: "GOVERNMENT / PROGRAMS",
          title: "Delivery across departments.",
          description: "Accountable owners and the decisions between offices.",
          rows: [
            {
              label: "Procurement approval",
              value: "Blocked",
              icon: "link",
              tone: "amber",
            },
            {
              label: "Infrastructure review",
              value: "In progress",
              icon: "document",
              tone: "sky",
            },
            {
              label: "Department handoff",
              value: "Owner ready",
              icon: "team",
              tone: "sage",
            },
          ],
          decision: {
            title: "Resolve the procurement decision.",
            description:
              "Keep the accountable department attached to the next action.",
          },
          sources: ["goal", "document", "team"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
      {
        id: "local-deployment",
        label: "Deployment planning",
        caption: "Infrastructure, ownership, and reporting planned together.",
        preview: {
          kind: "review",
          toolbar: "Deployment planning",
          status: "Custom",
          eyebrow: "GOVERNMENT / ENVIRONMENT",
          title: "Built around your requirements.",
          description: "A proposed structure for a controlled environment.",
          rows: [
            {
              label: "Hosting environment",
              value: "Government-led",
              icon: "lock",
              tone: "sky",
            },
            {
              label: "Department structure",
              value: "Mapped",
              icon: "team",
              tone: "sage",
            },
            {
              label: "Approval stages",
              value: "For review",
              icon: "check",
              tone: "amber",
            },
          ],
          decision: {
            title: "Review the deployment requirements.",
            description:
              "Agree the infrastructure and data boundaries with your team.",
          },
          sources: ["lock", "document", "team"],
          note: {
            title: "Requirements for review.",
            detail: "Deployment scope stays with your team.",
          },
        },
      },
      {
        id: "public-service-delivery",
        label: "Service delivery",
        caption:
          "A public-service commitment gets a clear owner and next step.",
        preview: {
          kind: "intake",
          source: {
            icon: "document",
            label: "Service request",
            badge: "New",
            author: "Permit follow-up",
            authorIcon: "document",
            message:
              "Please review the permit backlog and identify the next action.",
          },
          task: {
            identifier: "GOV-24",
            title: "Review permit follow-up",
            description: "Service commitment and department context attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Service team",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "2 days",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original service request",
        },
      },
    ],
  },
  marketing: {
    image: "goals",
    workflows: [
      {
        id: "campaign-planning",
        label: "Campaign intake",
        caption: "The brief stays attached as the campaign becomes work.",
        preview: {
          kind: "intake",
          source: {
            icon: "Google Drive",
            label: "Campaign brief",
            badge: "Draft",
            author: "Activation campaign",
            authorIcon: "document",
            message:
              "Prepare the landing page, email, and launch assets for this campaign.",
          },
          task: {
            identifier: "MKT-32",
            title: "Prepare campaign launch assets",
            description: "Campaign brief and activation goal attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Content team",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "2 days",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original campaign brief",
        },
      },
      {
        id: "approval-flow",
        label: "Approval review",
        caption: "See who owns the decision between ready and launched.",
        preview: {
          kind: "review",
          toolbar: "Campaign approvals",
          status: "Draft",
          eyebrow: "MARKETING / LAUNCH",
          title: "The next approval is clear.",
          description: "Assets, reviewers, and the handoffs before launch.",
          rows: [
            {
              label: "Landing page",
              value: "Product review",
              icon: "document",
              tone: "amber",
            },
            {
              label: "Launch email",
              value: "Design review",
              icon: "link",
              tone: "sky",
            },
            {
              label: "Campaign brief",
              value: "Ready",
              icon: "check",
              tone: "sage",
            },
          ],
          decision: {
            title: "Review the landing page copy.",
            description:
              "Keep the next reviewer and launch dependency visible.",
          },
          sources: ["Figma", "Google Drive", "Slack"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
      {
        id: "post-launch",
        label: "Launch follow-up",
        caption: "Turn the learning into work that fits the next week.",
        preview: {
          kind: "intake",
          source: {
            icon: "Google Drive",
            label: "Campaign review",
            badge: "Notes",
            author: "Post-launch learning",
            authorIcon: "document",
            message:
              "The landing page needs a clearer call to action. Plan the next iteration.",
          },
          task: {
            identifier: "MKT-33",
            title: "Refine the campaign landing page",
            description:
              "Performance note and original campaign brief attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Content team",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "1 day",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original performance note",
        },
      },
    ],
  },
  leadership: {
    image: "ai-planning",
    workflows: [
      {
        id: "priorities-to-work",
        label: "Priority review",
        caption: "See the execution behind each company priority.",
        preview: {
          kind: "review",
          toolbar: "Company priorities",
          status: "Draft",
          eyebrow: "LEADERSHIP / PRIORITIES",
          title: "From strategy to active work.",
          description: "Goals, owners, and the progress behind the update.",
          rows: [
            {
              label: "Improve activation",
              value: "In progress",
              icon: "goal",
              tone: "sky",
            },
            {
              label: "Onboarding handoff",
              value: "Needs decision",
              icon: "link",
              tone: "amber",
            },
            {
              label: "Reporting refresh",
              value: "For review",
              icon: "document",
              tone: "sage",
            },
          ],
          decision: {
            title: "Move onboarding work forward.",
            description:
              "Review the linked work and the team that owns the next step.",
          },
          sources: ["goal", "document", "team"],
          note: {
            title: "Prepared for review.",
            detail: "No assignments changed.",
          },
        },
      },
      {
        id: "operating-rhythm",
        label: "Weekly decisions",
        caption: "The review ends with a clear, traceable next action.",
        preview: {
          kind: "intake",
          source: {
            icon: "document",
            label: "Weekly packet",
            badge: "Review",
            author: "Leadership decision",
            authorIcon: "goal",
            message:
              "Pause the reporting refresh so the team can focus on onboarding.",
          },
          task: {
            identifier: "LEAD-18",
            title: "Review the reporting tradeoff",
            description: "Company priority and leadership decision attached.",
            fields: [
              {
                label: "Suggested owner",
                value: "Operations lead",
                icon: "team",
              },
              {
                label: "Estimate",
                value: "1 day",
                icon: "clock",
              },
              {
                label: "Start",
                value: "Thursday morning",
                icon: "Google Calendar",
              },
            ],
          },
          sourceCaption: "Original leadership decision",
        },
      },
      {
        id: "capacity-decisions",
        label: "Capacity decisions",
        caption: "See what saying yes means for the teams doing the work.",
        preview: {
          kind: "capacity",
          toolbar: "Company capacity",
          allocations: [
            {
              team: "Engineering",
              percent: 82,
            },
            {
              team: "Operations",
              percent: 68,
            },
          ],
          recommendation:
            "Review what can move before adding the next priority.",
          startDay: "Thu",
          proposal: "Operations · 3 days · Thursday morning",
        },
      },
    ],
  },
} satisfies Record<UseCaseSlug, HandoffConfig>;

export function getHandoffConfig(slug: string): HandoffConfig {
  if (!Object.prototype.hasOwnProperty.call(HANDOFF_CONFIGS, slug)) {
    throw new Error(`Missing handoff design for use case: ${slug}`);
  }
  return HANDOFF_CONFIGS[slug as UseCaseSlug];
}
