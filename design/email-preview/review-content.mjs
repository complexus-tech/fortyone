const templates = "apps/server/templates/";
const rules = "apps/server/internal/modules/notifications/service/";
const jobs = "apps/server/pkg/jobs/";
const replies = "apps/server/internal/modules/emailreply/service/";
const action = (label, path = "workspace") => ({ label, path });
const sam = { name: "Sam Taylor", initials: "ST" };
const alex = { name: "Alex Morgan", initials: "AM" };
const base = (id, name, group, title, paragraphs, extra = {}) => ({
  id,
  name,
  group,
  title,
  paragraphs,
  subject: title,
  preheader: paragraphs[0] || title,
  source: `${templates}notifications/notification.html`,
  fields: { workspace: "Acme", recipient: "Alex Morgan" },
  note: "Proposed copy with fictional sample data. Production copy is unchanged. Initials are a preview treatment; profile photo data is not yet connected to email delivery.",
  footerLink: "Manage notification preferences",
  ...extra,
});
const access = (id, name, title, paragraphs, extra) =>
  base(id, name, "Access", title, paragraphs, {
    footerLink: undefined,
    footer: "If you didn’t request this email, you can ignore it.",
    icon: "lock",
    ...extra,
  });
const workspace = (id, name, title, paragraphs, extra) =>
  base(id, name, "Workspace", title, paragraphs, {
    footerLink: undefined,
    footer: "Sent to you as an administrator of Acme.",
    ...extra,
  });
const activity = (id, name, title, text, extra = {}) =>
  base(id, name, "Activity", title, [], {
    preheader: text || title,
    person: sam,
    updates: [{ id: "ACM-142", title: "Improve the onboarding flow", text }],
    action: action("View story", "stories/ACM-142"),
    source: `${rules}rules_story_updates.go`,
    ...extra,
  });
const maya = (id, name, title, paragraphs, extra = {}) =>
  base(id, name, "Maya", title, paragraphs, {
    maya: true,
    footerLink: undefined,
    footer: "You’re receiving this reply in your conversation with Maya.",
    source: replies,
    ...extra,
  });

export const reviewEmails = [
  access(
    "login-link",
    "Sign-in link + code",
    "Sign in to FortyOne",
    [
      "Use the button below to sign in, or enter this code in FortyOne.",
      "Your link and code expire in 10 minutes.",
    ],
    {
      code: "482 916",
      action: action("Sign in", "auth/verify"),
      source: `${templates}auth/verification.html`,
    },
  ),
  access(
    "login-link-only",
    "Sign-in link",
    "Sign in to FortyOne",
    ["Your sign-in link is ready. It expires in 10 minutes."],
    {
      action: action("Sign in", "auth/verify"),
      source: `${templates}auth/verification.html`,
      note: "Variant when no OTP is supplied. Expiry must match the authentication producer.",
    },
  ),
  access(
    "login-mobile",
    "Mobile verification code",
    "Your sign-in code",
    [
      "Enter this code in FortyOne to sign in.",
      "This code expires in 10 minutes.",
    ],
    { code: "482 916", source: `${templates}auth/verification_mobile.html` },
  ),
  access(
    "portal-verification",
    "Feedback portal verification",
    "Verify your email",
    [
      "Verify your email to continue to Acme’s feedback portal.",
      "Use the button below or enter this code. Both expire in 10 minutes.",
    ],
    {
      code: "482 916",
      action: action("Verify email", "feedback/verify"),
      source: `${templates}feedback/verification.html`,
    },
  ),
  workspace(
    "invitation",
    "Workspace invitation",
    "Your team is ready for you",
    [
      "Sam invited you to join Acme on FortyOne. Join your team to plan projects, track stories, and keep work moving.",
    ],
    {
      subject: "Sam invited you to Acme",
      preheader:
        "Join your team on FortyOne. Your invitation expires in 7 days.",
      image: "invitation.png",
      imageAlt: "A warm illustration of a shared team workspace",
      person: { ...sam, role: "Invited you" },
      action: action("Join Acme", "invitations/accept"),
      helper: "Your invitation expires in 7 days.",
      footer: "If you weren’t expecting this invitation, you can ignore it.",
      source: `${templates}invites/invitation.html`,
    },
  ),
  workspace(
    "invitation-accepted",
    "Invitation accepted",
    "Alex has joined Acme",
    ["Alex accepted your invitation and is now part of your workspace."],
    {
      image: "invitation-accepted.png",
      imageAlt: "A warm illustration welcoming a new teammate",
      person: { ...alex, role: "New teammate" },
      action: action("Open Acme"),
      footer: "You invited Alex to join Acme.",
      source: `${templates}invites/acceptance.html`,
    },
  ),
  workspace(
    "account-inactivity",
    "Account inactivity warning",
    "Keep your FortyOne account active",
    [
      "You haven’t signed in for a while. Your account will be deactivated in 30 days if it remains inactive.",
      "Sign in before then to keep using FortyOne.",
    ],
    {
      icon: "clock",
      action: action("Sign in", "login"),
      footer: "This notice concerns your FortyOne account.",
      source: `${templates}users/inactivity_warning.html`,
    },
  ),
  workspace(
    "workspaces-inactivity_warning",
    "Workspace inactivity warning",
    "Keep your Acme workspace",
    [
      "Acme has been inactive and is scheduled for deletion in 30 days.",
      "Open Acme before then to keep the workspace and its data.",
    ],
    {
      icon: "clock",
      warning: {
        title: "If the workspace stays inactive",
        text: "It will be deleted, and its data will be permanently removed.",
      },
      action: action("Open Acme"),
      source: `${templates}workspaces/inactivity_warning.html`,
    },
  ),
  workspace(
    "deletion-requested",
    "Deletion scheduled · requester",
    "Acme is scheduled for deletion",
    [
      "You requested to delete Acme. Its data will be permanently deleted after the time below.",
      "Changed your mind? Cancel the deletion in workspace settings before then.",
    ],
    {
      icon: "clock",
      rows: [["Deletion scheduled", "7 September 2026, 14:00 UTC"]],
      action: action("Manage deletion", "settings/workspace"),
      source: `${templates}workspaces/deletion_scheduled_confirmation.html`,
    },
  ),
  workspace(
    "deletion-admin",
    "Deletion scheduled · other admins",
    "Acme is scheduled for deletion",
    [
      "Sam requested to delete Acme. Its data will be permanently deleted after the time below.",
      "As a workspace administrator, you can cancel the deletion in workspace settings before then.",
    ],
    {
      icon: "clock",
      rows: [["Deletion scheduled", "7 September 2026, 14:00 UTC"]],
      action: action("Manage deletion", "settings/workspace"),
      source: `${templates}workspaces/deletion_scheduled_notification.html`,
    },
  ),
  workspace(
    "restored-requester",
    "Workspace restored · requester",
    "Acme has been restored",
    [
      "You restored Acme. The workspace is available again, and its scheduled deletion has been cancelled.",
    ],
    {
      icon: "check",
      action: action("Open Acme"),
      source: `${templates}workspaces/restored_confirmation.html`,
    },
  ),
  workspace(
    "restored-admin",
    "Workspace restored · other admins",
    "Acme has been restored",
    [
      "Sam restored Acme. The workspace is available again, and its scheduled deletion has been cancelled.",
    ],
    {
      icon: "check",
      action: action("Open Acme"),
      source: `${templates}workspaces/restored_notification.html`,
    },
  ),
  ...[
    [
      "assigned",
      "Story assigned",
      "Sam assigned you a story",
      "You’re now assigned to this story.",
    ],
    [
      "reassigned",
      "Story reassigned",
      "Your story has a new assignee",
      "Sam reassigned this story from you to Jordan.",
    ],
    [
      "unassigned",
      "Story unassigned",
      "You’re no longer assigned",
      "Sam removed you as the assignee of this story.",
    ],
    [
      "collaborator-added",
      "Added as collaborator",
      "You’re now a collaborator",
      "Sam added you as a collaborator on this story.",
    ],
    [
      "collaborator-removed",
      "Removed as collaborator",
      "Your collaboration role changed",
      "Sam removed you as a collaborator on this story.",
    ],
    [
      "priority",
      "Priority changed",
      "A story’s priority changed",
      "Sam changed the priority from Medium to High.",
    ],
    [
      "status",
      "Status changed",
      "A story is ready for review",
      "Sam changed the status from In progress to In review.",
    ],
    [
      "description",
      "Description updated",
      "A story’s description changed",
      "Sam updated the description. Open the story to see the latest requirements.",
    ],
    [
      "due-date-set",
      "Due date added",
      "A story now has a due date",
      "Sam set the due date to 11 September 2026.",
    ],
    [
      "due-date-changed",
      "Due date changed",
      "A story’s due date changed",
      "Sam moved the due date from 11 September to 14 September 2026.",
    ],
    [
      "due-date-removed",
      "Due date removed",
      "A story’s due date was removed",
      "Sam removed the due date. This story no longer has a deadline.",
    ],
    [
      "start-date",
      "Start date changed",
      "A story’s start date changed",
      "Sam moved the start date from 7 September to 8 September 2026.",
    ],
    [
      "sprint",
      "Sprint changed",
      "A story moved to another sprint",
      "Sam moved this story from Sprint 12 to Sprint 13.",
    ],
    [
      "estimate",
      "Estimate changed",
      "A story’s estimate changed",
      "Sam changed the estimate from 3 to 5 points.",
    ],
    [
      "collaborators",
      "Collaborators updated",
      "A story’s collaborators changed",
      "Sam updated the collaborators on this story.",
    ],
    [
      "renamed",
      "Story renamed",
      "A story has a new name",
      "Sam renamed “Refresh onboarding” to “Improve the onboarding flow”.",
    ],
    [
      "story-update",
      "Other story updates",
      "Sam updated a story",
      "Open the story to review the latest changes.",
    ],
  ].map(([id, name, title, text]) => activity(id, name, title, text)),
  activity(
    "maya-assigned",
    "Story assigned by Maya",
    "Maya assigned you a story",
    "This story matches your frontend experience and your current availability.",
    {
      person: undefined,
      source: `${rules}rules.go`,
      note: "Assignment reason is sample content. Use the supplied Maya reason, never a made-up explanation.",
    },
  ),
  ...[
    [
      "scheduled",
      "Story scheduled",
      "A story has been scheduled",
      "Scheduled for 8 September, 09:00–10:30.",
    ],
    [
      "schedule-day",
      "Scheduled day changed",
      "A story moved to another day",
      "Moved from 8 September to 9 September, 09:00–10:30.",
    ],
    [
      "schedule-time",
      "Scheduled time changed",
      "A story’s scheduled time changed",
      "Moved from 09:00–10:30 to 11:00–12:30 on 8 September.",
    ],
    [
      "needs-time",
      "Scheduling · needs time",
      "A story needs more time",
      "There isn’t enough time allocated to finish this story. Review its estimate and schedule.",
    ],
    [
      "at-risk",
      "Scheduling · at risk",
      "A story may miss its deadline",
      "The current schedule puts this story’s deadline at risk. Review the work and available time.",
    ],
    [
      "cannot-fit",
      "Scheduling · cannot fit",
      "A story couldn’t fit in your schedule",
      "There isn’t enough available time before the deadline. Review the estimate, deadline, or your availability.",
    ],
  ].map(([id, name, title, text]) =>
    activity(id, name, title, text, {
      person: undefined,
      icon: "clock",
      source: `${rules}rules_story_schedule.go`,
      note: "Scheduling example. Render actual dates and times in the recipient’s timezone; preserve the supplied scheduling reason.",
    }),
  ),
  activity("comment", "New story comment", "Sam commented on your story", "", {
    updates: [
      {
        id: "ACM-142",
        title: "Improve the onboarding flow",
        text: "",
        quote:
          "The updated screens are ready. Could you review the welcome step before we hand this over?",
      },
    ],
  }),
  activity(
    "conversation",
    "Conversation reply",
    "Sam replied to your comment",
    "The updated screens are ready for review.",
    {
      updates: undefined,
      conversation: {
        original: "Can we make the welcome step shorter?",
        author: "Sam Taylor",
        reply:
          "Yes, I’ve reduced it to two steps and updated the designs. Could you take another look?",
      },
      action: action("View conversation", "stories/ACM-142/comments"),
      source: `${rules}rules.go`,
    },
  ),
  activity("mention", "Mention", "Sam mentioned you", "", {
    updates: [
      {
        id: "ACM-142",
        title: "Improve the onboarding flow",
        text: "",
        quote:
          "Alex, could you confirm whether these screens cover the sign-in flow?",
      },
    ],
    action: action("View mention", "stories/ACM-142/comments"),
  }),
  base(
    "objective-update",
    "Objective update",
    "Strategy",
    "Sam updated an objective",
    [],
    {
      person: sam,
      preheader: "Improve activation is now on track.",
      updates: [
        {
          title: "Improve activation",
          text: "Health changed from At risk to On track.",
        },
      ],
      action: action("View objective", "objectives/activation"),
      source: `${rules}rules.go`,
    },
  ),
  base(
    "key-result-update",
    "Key result update",
    "Strategy",
    "A key result has new progress",
    [],
    {
      person: sam,
      preheader: "First-week activation increased from 42% to 48%.",
      updates: [
        {
          title: "Increase first-week activation to 60%",
          text: "Progress changed from 42% to 48%.",
        },
      ],
      action: action("View key result", "objectives/activation"),
      source: `${rules}rules.go`,
    },
  ),
  base(
    "strategy-planning",
    "Planning reminder",
    "Strategy",
    "Finish Acme’s strategy for Q4",
    [
      "Your Q4 strategy is missing a few details. Add them to give your team a clear plan for the quarter.",
    ],
    {
      rows: [["Still to add", "Objectives and key results"]],
      action: action("Continue planning", "strategy"),
      source: `${jobs}strategy_communications.go`,
    },
  ),
  base(
    "strategy-check-in",
    "Strategy check-in reminder",
    "Strategy",
    "Two objectives need a check-in",
    [
      "Add a short update so your team can see what’s on track and where support is needed.",
    ],
    {
      updates: [
        { title: "Improve activation", text: "No check-in in 14 days." },
        {
          title: "Reduce support response time",
          text: "At risk · Last checked in 10 days ago.",
        },
      ],
      action: action("Review objectives", "objectives"),
      source: `${jobs}strategy_weekly_check_ins.go`,
    },
  ),
  base(
    "strategy-monthly",
    "Monthly strategy summary",
    "Strategy",
    "Acme’s August strategy review",
    ["Here’s how your objectives and delivery progressed in August."],
    {
      rows: [
        ["Objectives", "3 on track · 1 at risk"],
        ["Stories completed", "24"],
        ["Still needs attention", "Support response time"],
      ],
      action: action("Review strategy", "strategy"),
      source: `${jobs}strategy_communications.go`,
      note: "Illustrative summary values. Only include counts and comparisons available in the job’s snapshot.",
    },
  ),
  ...[
    [
      "feedback-comment",
      "Feedback comment",
      "A new reply to your feedback",
      "Sam replied to “Add saved views”",
      "Thanks for the suggestion. Which filters would you like to save?",
    ],
    [
      "feedback-status",
      "Feedback status changed",
      "Your feedback is now planned",
      "Add saved views",
      "The status changed from Under review to Planned.",
    ],
    [
      "feedback-published",
      "Feedback update published",
      "There’s an update on your feedback",
      "Add saved views",
      "Saved views are now available. You can save your filters and return to them from the sidebar.",
    ],
    [
      "feedback-merged",
      "Feedback merged",
      "Your feedback was merged",
      "Add saved views",
      "Your feedback was merged into “Save and share filtered views”. Open that item to follow future updates.",
    ],
  ].map(([id, name, title, item, text]) =>
    base(id, name, "Feedback", title, [], {
      preheader: text || title,
      person: id === "feedback-comment" ? sam : undefined,
      updates: [{ title: item, text }],
      action: action("View feedback", "feedback/saved-views"),
      source: `${rules}rules.go`,
    }),
  ),
  base(
    "activity-digest",
    "Grouped activity updates",
    "Digests",
    "Three updates in Acme",
    ["Here’s what changed in the work you follow."],
    {
      updates: [
        {
          id: "ACM-142",
          title: "Improve the onboarding flow",
          text: "Sam moved this story to In review.",
        },
        {
          id: "ACM-156",
          title: "Add saved views",
          text: "Jordan assigned this story to you.",
        },
        {
          id: "ACM-161",
          title: "Update the billing page",
          text: "Sam mentioned you in a comment.",
        },
      ],
      action: action("View updates", "notifications"),
      source: "apps/server/internal/taskhandlers/notification_handlers.go",
    },
  ),
  base(
    "weekly-digest",
    "Weekly digest · FortyOne",
    "Digests",
    "Your week in Acme",
    [
      "Here’s a quick look at your team’s progress and what needs attention next.",
    ],
    {
      updates: [
        {
          title: "12 stories completed",
          text: "The onboarding refresh and saved views are ready.",
        },
        {
          title: "Three stories need attention",
          text: "Review overdue work before planning next week.",
        },
      ],
      action: action("Open Acme"),
      source: `${jobs}weekly_digest.go`,
    },
  ),
  base(
    "overdue-stories",
    "Overdue stories reminder",
    "Digests",
    "Two stories are overdue",
    ["Review these stories and update their status or due dates."],
    {
      icon: "clock",
      updates: [
        {
          id: "ACM-142",
          title: "Improve the onboarding flow",
          text: "Due 3 September · In progress",
        },
        {
          id: "ACM-156",
          title: "Add saved views",
          text: "Due 4 September · In review",
        },
      ],
      action: action("Review stories", "my-work"),
      source: `${jobs}overdue_stories.go`,
    },
  ),
  base(
    "overdue-objectives",
    "Overdue objectives reminder",
    "Digests",
    "An objective needs your attention",
    ["The due date has passed. Review progress and update the plan."],
    {
      icon: "clock",
      updates: [
        {
          title: "Improve activation",
          text: "Due 31 August · At risk",
          quote:
            "Key result: Increase first-week activation to 60%. Current progress: 48%.",
        },
      ],
      action: action("Review objective", "objectives/activation"),
      source: `${jobs}objective_overdue_email.go`,
    },
  ),
  base(
    "feedback-digest",
    "Feedback digest",
    "Digests",
    "Updates on your feedback",
    ["There’s new activity on two items you follow."],
    {
      updates: [
        {
          title: "Add saved views",
          text: "Sam asked which filters you’d like to save.",
        },
        {
          title: "Export reports as PDF",
          text: "The status changed to Planned.",
        },
      ],
      action: action("View feedback", "feedback"),
      source: `${jobs}feedback_digest.go`,
    },
  ),
  maya(
    "maya-weekly",
    "Weekly note",
    "Your week in Acme",
    [
      "Hi Alex,",
      "Your team completed 12 stories this week. The onboarding refresh is ready for review, and three stories still need attention.",
    ],
    {
      subject: "Your weekly note from Maya",
      updates: [
        {
          title: "Ready for review",
          text: "Sam finished the onboarding screens. Your feedback will help move them into development.",
        },
        {
          title: "Needs attention",
          text: "Three stories are overdue. Review what’s blocked before planning next week.",
        },
      ],
      closing:
        "Reply with what you’d like to look into, and I’ll help you work through it.",
      action: action("Open Acme"),
      source: `${jobs}weekly_digest.go`,
      note: "Weekly content is generated from actual workspace facts. Reply invitation is shown only when the reply thread is enabled.",
    },
  ),
  maya(
    "maya-answer",
    "Answer to a question",
    "Here’s where activation stands",
    [
      "First-week activation is at 48%, against a target of 60%.",
      "The onboarding refresh is in review. The next step is to review Sam’s screens and confirm what’s ready to build.",
    ],
    { action: action("View objective", "objectives/activation") },
  ),
  maya(
    "maya-clarify",
    "Clarification needed",
    "Which story should I update?",
    [
      "I found two stories called “Update the welcome email”. Reply with the story ID you mean.",
    ],
    {
      rows: [
        ["ACM-142", "Update the welcome email · Product"],
        ["ACM-208", "Update the welcome email · Growth"],
      ],
    },
  ),
  maya(
    "maya-confirmation",
    "Proposed change · confirm",
    "Ready for your confirmation",
    [
      "I can change the health of “Improve activation” from At risk to On track.",
      "I haven’t applied this change yet.",
    ],
    {
      rows: [
        ["Objective", "Improve activation"],
        ["Current health", "At risk"],
        ["Proposed health", "On track"],
      ],
      confirmation: true,
      note: "Exactly one proposed mutation. Preserve the literal CONFIRM and CANCEL commands and the exact target and values. Never imply an unconfirmed change has been applied.",
    },
  ),
  maya(
    "maya-applied",
    "Change applied",
    "The objective is now on track",
    [
      "I changed the health of “Improve activation” from At risk to On track.",
      "Reply if you’d like help with another update.",
    ],
    { icon: "check", source: `${replies}processor_proposals.go` },
  ),
  maya(
    "maya-already-applied",
    "Change already applied",
    "That change is already complete",
    [
      "“Improve activation” is already set to On track. No further change was made.",
    ],
    { icon: "check", source: `${replies}processor_proposals.go` },
  ),
  maya(
    "maya-cancelled",
    "Change cancelled",
    "The proposed change is cancelled",
    [
      "I cancelled the proposed health change for “Improve activation”. I haven’t applied it.",
      "Reply if you’d like to try a different update.",
    ],
    {
      note: "Use the proposal’s recorded values only when still valid. If current health is not rechecked, omit the sentence asserting its present value.",
      source: `${replies}processor_proposals.go`,
    },
  ),
  maya(
    "maya-expired",
    "Proposal no longer pending",
    "Let’s prepare a fresh preview",
    [
      "That proposed change is no longer pending. I haven’t applied it.",
      "Reply with the update you want, and I’ll prepare a new preview.",
    ],
    { icon: "clock", source: `${replies}processor_proposals.go` },
  ),
  maya(
    "maya-conflict",
    "Item or access changed",
    "I couldn’t apply that change",
    [
      "The item or your access changed after the preview, so I couldn’t apply the update.",
      "Reply with what you want to change, and I’ll check it again.",
    ],
    { source: `${replies}processor_proposals.go` },
  ),
  maya(
    "maya-preview-unavailable",
    "Earlier preview unavailable",
    "I need to create a new preview",
    [
      "I couldn’t restore the earlier preview. I haven’t changed anything.",
      "Reply with the update you want, and I’ll prepare a fresh preview.",
    ],
    { source: `${replies}processor.go` },
  ),
  maya(
    "maya-refusal",
    "Unsupported request",
    "I can’t do that by email",
    [
      "I can’t delete a workspace from an email reply. You can manage workspace deletion in FortyOne’s workspace settings.",
    ],
    {
      source: "apps/server/internal/modules/emailagent/service/prompt.go",
      note: "Example of the refuse intent. Only name an alternative action that exists and is appropriate to the user’s access.",
    },
  ),
  maya(
    "maya-resumed",
    "Resume proposed change",
    "Your proposed change is ready",
    ["Here’s the change we were discussing. I haven’t applied it yet."],
    {
      rows: [
        ["Objective", "Improve activation"],
        ["Change health", "At risk → On track"],
      ],
      confirmation: true,
      source: `${replies}processor_helpers.go`,
    },
  ),
];
