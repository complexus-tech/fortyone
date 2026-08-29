import type { MarketingDetail } from "@/components/shared/marketing-detail-page";

export type Integration = MarketingDetail;

export const integrations: Integration[] = [
  {
    slug: "slack",
    label: "Slack",
    heroTitle: "Turn Slack conversations into coordinated work",
    metaTitle: "FortyOne for Slack | Turn Conversations Into Project Work",
    metaDescription:
      "Connect FortyOne to Slack to create stories from conversations, ask Maya about project work, and keep task context visible where your team communicates.",
    updatedDate: "2026-08-09",
    updatedLabel: "August 9, 2026",
    intro: [
      "Important work often starts in Slack, long before it has a clear owner, estimate, status, or place in the project plan. Copying that context into another tool slows the team down and makes it easy to lose the reason behind the request.",
      "The FortyOne Slack integration turns conversations into structured work. Create a story with the /fortyone command or a message shortcut, choose the right team and workflow details, and keep a link back to the Slack source.",
      "Maya also works in direct messages, mentions, and connected threads, so people can ask about project work without leaving Slack. FortyOne still applies the workspace and team permissions attached to each linked account.",
    ],
    benefits: [
      [
        "Faster work intake",
        "Turn a Slack message or slash command into a structured FortyOne story while the context is still fresh.",
      ],
      [
        "Less tool switching",
        "Ask Maya about work, owners, priorities, and next steps from the conversations where coordination already happens.",
      ],
      [
        "Context that stays connected",
        "Keep source conversations, story links, rich previews, and supported thread updates tied to the work they created.",
      ],
      [
        "Permission-aware answers",
        "Slack identities are linked to FortyOne accounts so responses and actions respect workspace, team, and channel access.",
      ],
    ],
    previewCards: [
      {
        heading: "Create from Slack",
        subheading: "Conversation turned into a story",
        badge: "Created",
        rows: [
          { label: "Story", value: "Prepare onboarding handoff" },
          { label: "Team", value: "Operations" },
        ],
      },
      {
        heading: "Ask Maya",
        subheading: "Project context in the thread",
        badge: "Ready",
        rows: [
          { label: "Question", value: "What is blocking the launch?" },
          { label: "Answer", value: "Two decisions need review" },
        ],
      },
    ],
    sections: [
      {
        id: "create-work",
        title: "Create project work from the conversation",
        paragraphs: [
          "Use /fortyone to open the create form with a title already in mind, or choose Create a story from a Slack message to carry the message into the new story as source context.",
          "Before the story is created, you can select the FortyOne team, workflow status, priority, assignee, labels, and objective available to your linked account. The result is real project work, not a disconnected copy of the Slack message.",
        ],
        rows: [
          [
            "/fortyone",
            "Open a structured story form from any supported Slack conversation",
          ],
          [
            "Message shortcut",
            "Use the selected message to prefill the title, description, and source",
          ],
          [
            "Project fields",
            "Choose the team, status, priority, assignee, labels, and objective",
          ],
          [
            "Source link",
            "Keep a direct path between the new story and the Slack conversation",
          ],
        ],
        tableHead: ["Slack action", "What FortyOne does"],
      },
      {
        id: "maya",
        title: "Ask Maya about work without leaving Slack",
        paragraphs: [
          "Message Maya directly or mention the FortyOne app in a channel to ask questions about work in your workspace. Once a thread is connected, replies can continue with the context from that conversation for a limited period.",
          "Workspace administrators can add custom guidance for terminology, response style, and operating rules. Answers stay constrained by the FortyOne teams the linked person can access and by the channel audience configured for the integration.",
        ],
        cards: [
          {
            heading: "Project question",
            subheading: "Answered with workspace context",
            rows: [
              {
                label: "Blocked",
                value: "Checkout review is waiting on Product",
              },
              { label: "Due next", value: "Launch checklist · Friday" },
            ],
          },
          {
            heading: "Controlled action",
            subheading: "Review important changes first",
            badge: "Confirm",
            rows: [
              { label: "Proposal", value: "Move OPS-42 to In progress" },
              { label: "Control", value: "Confirm or cancel in Slack" },
            ],
          },
        ],
      },
      {
        id: "previews",
        title: "Keep FortyOne work readable in Slack",
        paragraphs: [
          "When someone shares a supported FortyOne story link, Slack can show a rich preview with the details the viewer is allowed to see. The preview makes the story easier to discuss without exposing work to people who cannot access it in FortyOne.",
          "Connected request threads can also keep supported discussion tied to the corresponding work, reducing the split between a decision in Slack and the record the delivery team follows.",
        ],
        rows: [
          [
            "Story preview",
            "Show status, priority, ownership, dates, and other useful work context",
          ],
          [
            "Account check",
            "Prompt an unlinked person to securely connect the correct FortyOne account",
          ],
          [
            "Access check",
            "Withhold private story details when the viewer is not authorized",
          ],
          [
            "Thread context",
            "Keep supported request discussions connected to the work record",
          ],
        ],
        tableHead: ["Capability", "How it behaves"],
      },
    ],
    questions: [
      [
        "Who can connect a Slack workspace?",
        "A FortyOne workspace administrator starts the Slack installation. Slack may also require approval from someone who is allowed to install apps in that Slack workspace.",
      ],
      [
        "Does every Slack member get access to FortyOne data?",
        "No. People must use a linked FortyOne account, and FortyOne applies their existing workspace and team access before returning project details or accepting actions.",
      ],
      [
        "What Slack data does FortyOne process?",
        "FortyOne processes installation and channel identifiers, linked-user identifiers, and the content people deliberately send through supported commands, message actions, direct messages, mentions, subscribed threads, or shared FortyOne links. More detail is available in the Privacy Policy.",
      ],
      [
        "What happens when Slack is disconnected?",
        "Slack commands, message actions, previews, and Maya responses stop. FortyOne removes the active connection credentials and associated connection records, subject to the limited security and legal retention described in the Privacy Policy.",
      ],
    ],
  },
  {
    slug: "google-calendar",
    label: "Google Calendar",
    heroTitle: "Plan work around your Google Calendar",
    metaTitle: "FortyOne for Google Calendar | Plan Around Real Availability",
    metaDescription:
      "Connect your primary Google Calendar to see meetings beside FortyOne work and plan schedules around real availability without exposing private event details.",
    updatedDate: "2026-08-09",
    updatedLabel: "August 9, 2026",
    intro: [
      "A project plan can look achievable while the people doing the work are already committed to meetings, reviews, and focused work elsewhere. Without calendar context, those conflicts only become visible after a deadline starts slipping.",
      "FortyOne connects to your primary Google Calendar and brings its availability into the FortyOne Calendar. You can see meetings beside scheduled work and make planning decisions with a more realistic picture of the time you actually have.",
      "Detailed event information remains visible only to the calendar owner. Private events stay marked as Busy, while teammates, managers, capacity planning, and Maya receive availability rather than meeting content.",
    ],
    benefits: [
      [
        "A realistic daily plan",
        "See meetings and scheduled project work together before committing to another task window.",
      ],
      [
        "Fewer hidden conflicts",
        "Availability signals help planning avoid time that is already occupied on the connected calendar.",
      ],
      [
        "Private by design",
        "Event details are owner-only, private events remain Busy, and shared planning paths use availability instead of content.",
      ],
      [
        "A connection you control",
        "Reconnect, refresh, or disconnect your calendar from FortyOne settings whenever you need to.",
      ],
    ],
    previewCards: [
      {
        heading: "Today",
        subheading: "Meetings and project work together",
        badge: "Synced",
        rows: [
          { label: "09:00", value: "Product review" },
          { label: "11:00", value: "Focus block · OPS-42" },
        ],
      },
      {
        heading: "Availability check",
        subheading: "A clear window before assignment",
        badge: "Open",
        rows: [
          { label: "Busy", value: "14:00–14:45" },
          { label: "Suggested", value: "Start after 15:00" },
        ],
      },
    ],
    sections: [
      {
        id: "calendar-view",
        title: "See meetings beside the work you plan",
        paragraphs: [
          "FortyOne keeps a rolling view of your primary calendar, currently covering seven days back and ninety days ahead. That makes the Calendar useful for planning the current week while still showing the commitments that shape upcoming work.",
          "When event-detail access is available, the owner can see titles, locations, meeting links, descriptions, organizers, and attendees inside FortyOne. If only availability access is available, the integration still shows the time as Busy without storing those details.",
        ],
        rows: [
          [
            "Meetings",
            "Display connected calendar events alongside FortyOne schedule blocks",
          ],
          [
            "Availability",
            "Mark occupied windows so planned work can avoid obvious conflicts",
          ],
          [
            "Sync range",
            "Keep a rolling snapshot from seven days back to ninety days ahead",
          ],
          [
            "Refresh",
            "Sync the connection again from integration settings when needed",
          ],
        ],
        tableHead: ["Calendar signal", "How FortyOne uses it"],
      },
      {
        id: "privacy",
        title: "Keep meeting details with the calendar owner",
        paragraphs: [
          "Calendar content can include sensitive customer, hiring, financial, and personal information. FortyOne separates the detailed owner view from the availability signal used elsewhere in the workspace.",
          "Private and confidential events are stored without their descriptive details. Teammates and managers see that time is unavailable, while Maya and capacity-planning paths receive title-free busy windows rather than calendar content.",
        ],
        cards: [
          {
            heading: "Your calendar",
            subheading: "Owner-only event view",
            rows: [
              { label: "Title", value: "Quarterly planning" },
              { label: "Meeting link", value: "Available to you" },
            ],
          },
          {
            heading: "Shared planning",
            subheading: "Availability without meeting content",
            badge: "Private",
            rows: [
              { label: "Status", value: "Busy" },
              { label: "Details", value: "Not shared" },
            ],
          },
        ],
      },
      {
        id: "planning",
        title: "Make availability part of the plan",
        paragraphs: [
          "Availability is most useful when it changes a planning decision. FortyOne uses connected busy windows when showing workload and preparing schedule-aware recommendations, so a seemingly open day is not treated as entirely free.",
          "The integration is currently read-only: FortyOne reads the primary calendar to show events and availability, but it does not create, edit, or delete Google Calendar events.",
        ],
        rows: [
          [
            "Before assignment",
            "Check whether the proposed owner has a workable time window",
          ],
          [
            "Calendar view",
            "Compare scheduled stories with meetings on the same timeline",
          ],
          [
            "Maya planning",
            "Use availability signals without sending event titles into planning context",
          ],
          ["Google Calendar", "Leave the source calendar unchanged"],
        ],
        tableHead: ["Planning moment", "What the connection adds"],
      },
    ],
    questions: [
      [
        "Which Google Calendar does FortyOne connect?",
        "FortyOne currently connects the primary calendar for the Google account you authorize.",
      ],
      [
        "Can my teammates see my meeting details?",
        "No. Detailed event information is owner-only. Shared planning paths receive availability, and private or confidential events remain Busy without descriptive content.",
      ],
      [
        "Does FortyOne change events in Google Calendar?",
        "No. The current integration is read-only and does not create, edit, or delete Google Calendar events.",
      ],
      [
        "What happens when I disconnect?",
        "FortyOne clears the connection credentials, scopes, cached events, and availability windows for that connection and stops syncing the calendar.",
      ],
    ],
  },
  {
    slug: "github",
    label: "GitHub",
    heroTitle: "Keep project work in sync with GitHub",
    metaTitle: "FortyOne for GitHub | Connect Issues, Pull Requests and Work",
    metaDescription:
      "Connect FortyOne to GitHub to sync issues with teams, link branches, commits and pull requests, and automate workflow updates from engineering activity.",
    updatedDate: "2026-08-09",
    updatedLabel: "August 9, 2026",
    intro: [
      "Engineering work is easiest to follow when the project plan and the code tell the same story. Without a direct connection, teams spend time copying issue updates, pull request links, review state, and delivery progress between systems.",
      "The FortyOne GitHub integration connects GitHub organizations and repositories to your workspace. Link a repository to a FortyOne team, choose whether issues flow into FortyOne or sync in both directions, and keep code activity attached to the project work it advances.",
      "Branches, commits, pull requests, reviews, and checks can update the FortyOne record and drive configurable workflow rules, while branch formats and magic words keep the connection predictable for developers.",
    ],
    benefits: [
      [
        "One record for delivery",
        "Keep issues, branches, commits, pull requests, reviews, and checks connected to the corresponding FortyOne story.",
      ],
      [
        "Less duplicate updating",
        "Use inbound or bidirectional issue sync instead of manually copying changes between GitHub and FortyOne.",
      ],
      [
        "Automatic workflow signals",
        "Move work when branches, commits, pull requests, reviews, or checks reach the events your team configures.",
      ],
      [
        "Consistent developer handoff",
        "Use workspace branch formats, story identifiers, pull request context, and commit keywords across repositories.",
      ],
    ],
    previewCards: [
      {
        heading: "Pull request linked",
        subheading: "Engineering activity on the story",
        badge: "Open",
        rows: [
          { label: "Story", value: "ENG-214 · Improve search" },
          { label: "Review", value: "1 approval" },
        ],
      },
      {
        heading: "Issue sync",
        subheading: "Repository connected to a team",
        badge: "Two-way",
        rows: [
          { label: "Repository", value: "fortyone/projects" },
          { label: "Team", value: "Engineering" },
        ],
      },
    ],
    sections: [
      {
        id: "issue-sync",
        title: "Connect GitHub issues to the right team",
        paragraphs: [
          "After the FortyOne GitHub App is installed, workspace administrators can choose a repository and link it to a FortyOne team. New and updated issues can then become project work with repository context attached.",
          "Use inbound-only sync when GitHub should be the source for issue changes, or bidirectional sync when supported title, description, and state changes should move between GitHub and FortyOne.",
        ],
        rows: [
          [
            "GitHub App",
            "Connect an organization and the repositories approved during installation",
          ],
          [
            "Repository link",
            "Map one repository to the FortyOne team that owns its work",
          ],
          ["Inbound only", "Bring GitHub issue changes into FortyOne"],
          [
            "Bidirectional",
            "Keep supported issue and FortyOne story changes synchronized",
          ],
        ],
        tableHead: ["Setup choice", "What it controls"],
      },
      {
        id: "code-activity",
        title: "Keep code activity attached to the work",
        paragraphs: [
          "Include a FortyOne story identifier in a branch name, commit, or pull request context to connect engineering activity with the work item. The story can then show the links and delivery state without requiring a manual status report.",
          "FortyOne can also prefill pull request descriptions with story details, link commits using magic words, synchronize supported assignee and label changes, and record review or check state for the linked pull request.",
        ],
        cards: [
          {
            heading: "Branch and commits",
            subheading: "Linked through the story identifier",
            rows: [
              { label: "Branch", value: "joseph/ENG-214-improve-search" },
              { label: "Commit", value: "Part of ENG-214" },
            ],
          },
          {
            heading: "Pull request",
            subheading: "Delivery state visible in FortyOne",
            badge: "Checks passed",
            rows: [
              { label: "Review", value: "Approved" },
              { label: "Assignee", value: "Linked FortyOne user" },
            ],
          },
        ],
      },
      {
        id: "automation",
        title: "Let engineering events move the workflow",
        paragraphs: [
          "Every team can decide which GitHub events should change a story's workflow status. A branch can mark work as started, a review can move it into review, and a merged pull request or closing commit can complete it.",
          "The automation stays explicit: administrators choose the event, optional base-branch pattern, and target FortyOne status. That keeps repository behavior aligned with each team's workflow instead of applying one global rule everywhere.",
        ],
        rows: [
          [
            "Branch created",
            "Move linked work into the configured active status",
          ],
          [
            "Pull request opened",
            "Reflect that implementation is ready for review",
          ],
          [
            "Review or check activity",
            "Move work when the team's review conditions are reached",
          ],
          [
            "Merge or closing keyword",
            "Complete the story using the team's configured rule",
          ],
        ],
        tableHead: ["GitHub event", "Example workflow effect"],
      },
    ],
    questions: [
      [
        "Can I limit which repositories FortyOne can access?",
        "Yes. GitHub shows the permissions requested by the FortyOne GitHub App and lets the installer choose all repositories or selected repositories.",
      ],
      [
        "Do GitHub issues have to sync both ways?",
        "No. Each repository link can use inbound-only sync or bidirectional sync, depending on which system should accept supported updates.",
      ],
      [
        "Can GitHub activity update a FortyOne workflow?",
        "Yes. Teams can configure workflow rules for supported branch, commit, pull request, review, check, and merge events.",
      ],
      [
        "What GitHub data does FortyOne keep?",
        "FortyOne stores installation and repository metadata plus the issue, pull request, branch, commit, review, check, and comment data needed for linked work and automation. The Privacy Policy explains retention and deletion choices.",
      ],
    ],
  },
];

export function getIntegrationBySlug(slug: string) {
  return integrations.find((integration) => integration.slug === slug);
}
