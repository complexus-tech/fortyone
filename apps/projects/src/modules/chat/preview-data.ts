import type { JSONContent } from "@tiptap/core";

export type PreviewPerson = {
  id: string;
  name: string;
};

export type PreviewTask = {
  id: string;
  title: string;
};

export type PreviewReaction = {
  emoji: string;
  count: number;
  reacted?: boolean;
};

export type PreviewMessage = {
  id: string;
  author: PreviewPerson;
  time: string;
  date?: string;
  text: string;
  richContent?: JSONContent | null;
  edited?: boolean;
  deleted?: boolean;
  reactions?: PreviewReaction[];
  replies?: PreviewMessage[];
  tasks?: PreviewTask[];
};

export type PreviewConversation = {
  id: string;
  name: string;
  kind: "team" | "dm";
  color?: string;
  description: string;
  members: PreviewPerson[];
  messages: PreviewMessage[];
};

export const PREVIEW_TODAY = "2026-10-16";

export const PREVIEW_YOU = { id: "you", name: "You" };
const SARAH = { id: "sarah", name: "Sarah Chen" };
const DANIEL = { id: "daniel", name: "Daniel Williams" };
const MAYA = { id: "maya", name: "Maya" };

// Sample review history for scrolling, date groups, and threaded UI states.
const PRODUCT_HISTORY: PreviewMessage[] = [
  {
    id: "product-20261012-01",
    author: SARAH,
    date: "2026-10-12",
    time: "09:08",
    text: "Morning! For this week’s product walkthrough, I’d like us to follow the journey from joining a workspace to finding a useful first task.",
  },
  {
    id: "product-20261012-02",
    author: DANIEL,
    date: "2026-10-12",
    time: "09:12",
    text: "Good place to start. I’ll look at the points where the interface asks someone to make a choice without much context.",
    reactions: [
      {
        emoji: "👀",
        count: 2,
      },
    ],
  },
  {
    id: "product-20261012-03",
    author: PREVIEW_YOU,
    date: "2026-10-12",
    time: "09:18",
    text: "I’ve put the review outline in one place.\n\nThree things to watch: what people notice first, where they hesitate, and whether they can tell what to do next.",
  },
  {
    id: "product-20261012-04",
    author: SARAH,
    date: "2026-10-12",
    time: "10:06",
    text: "One question for the review: when someone joins two teams, should their first screen suggest a team or show both equally?",
    replies: [
      {
        id: "product-20261012-04-reply-1",
        author: DANIEL,
        date: "2026-10-12",
        time: "10:11",
        text: "Show both, but make the invitation’s team the obvious starting point. We should avoid making the other team look unavailable.",
      },
      {
        id: "product-20261012-04-reply-2",
        author: PREVIEW_YOU,
        date: "2026-10-12",
        time: "10:19",
        text: "Agreed. A short “You were invited here” hint should be enough. Keep the switcher easy to find.",
      },
      {
        id: "product-20261012-04-reply-3",
        author: MAYA,
        date: "2026-10-12",
        time: "10:20",
        text: "Decision from this thread: keep both teams visible, use the invitation to suggest a starting point, and let people switch immediately. Open question: where the invitation hint belongs.",
      },
    ],
  },
  {
    id: "product-20261012-05",
    author: DANIEL,
    date: "2026-10-12",
    time: "10:11",
    text: "The empty team state is worth including too. It should feel intentional even when there are no tasks yet.",
  },
  {
    id: "product-20261012-06",
    author: SARAH,
    date: "2026-10-12",
    time: "11:24",
    text: "Yes. I’ll sketch that next to a team with real work so we can compare the two states side by side.",
    reactions: [
      {
        emoji: "👍",
        count: 1,
      },
    ],
  },
  {
    id: "product-20261012-07",
    author: PREVIEW_YOU,
    date: "2026-10-12",
    time: "14:10",
    text: "Small wording thought: “Create your first task” is clearer than “Get started” when that is the action we actually want someone to take.",
  },
  {
    id: "product-20261012-08",
    author: SARAH,
    date: "2026-10-12",
    time: "16:36",
    text: "Leaving the first pass here for tomorrow: a short introduction, one clear next step, and the team’s recent work below it. No extra setup screen.",
    reactions: [
      {
        emoji: "🙌",
        count: 3,
      },
    ],
  },
  {
    id: "product-20261013-01",
    author: DANIEL,
    date: "2026-10-13",
    time: "09:04",
    text: "Picked up the walkthrough notes. The invitation hint makes the first screen much easier to understand.",
  },
  {
    id: "product-20261013-02",
    author: SARAH,
    date: "2026-10-13",
    time: "09:08",
    text: "Nice. I’m trying the same layout with a longer team name this morning.",
  },
  {
    id: "product-20261013-03",
    author: PREVIEW_YOU,
    date: "2026-10-13",
    time: "09:17",
    text: "Please include a workspace with no profile photos too. Initials need to feel just as considered as the photo version.",
    reactions: [
      {
        emoji: "✅",
        count: 2,
      },
    ],
  },
  {
    id: "product-20261013-04",
    author: SARAH,
    date: "2026-10-13",
    time: "10:02",
    text: "For an empty team, do we show a sample task or keep the list genuinely empty?",
    replies: [
      {
        id: "product-20261013-04-reply-1",
        author: PREVIEW_YOU,
        date: "2026-10-13",
        time: "10:08",
        text: "I’d keep the list empty. We can explain what belongs there without making sample content look like the team’s work.",
      },
      {
        id: "product-20261013-04-reply-2",
        author: DANIEL,
        date: "2026-10-13",
        time: "10:14",
        text: "That also keeps selection and filtering predictable. A short example in the description should do the job.",
      },
      {
        id: "product-20261013-04-reply-3",
        author: SARAH,
        date: "2026-10-13",
        time: "10:18",
        text: "Makes sense. I’ll use the helper text for the example and keep the action focused on creating a task.",
      },
    ],
  },
  {
    id: "product-20261013-05",
    author: DANIEL,
    date: "2026-10-13",
    time: "10:06",
    text: "Another edge case: a team can have work, but nothing assigned to the person who just joined. Those two empty states need different copy.",
  },
  {
    id: "product-20261013-06",
    author: SARAH,
    date: "2026-10-13",
    time: "11:45",
    text: "Here’s the comparison sheet for those states. The layout stays the same; the explanation and suggested action change.",
  },
  {
    id: "product-20261013-07",
    author: PREVIEW_YOU,
    date: "2026-10-13",
    time: "13:18",
    text: "The second version reads better.\n\n“Nothing assigned to you yet” tells me what happened. Then “Explore team work” gives me somewhere useful to go instead of ending the journey.",
  },
  {
    id: "product-20261013-08",
    author: DANIEL,
    date: "2026-10-13",
    time: "15:20",
    text: "I’ll include the narrow-screen version in the next walkthrough. The action should stay near its explanation when the text wraps.",
    reactions: [
      {
        emoji: "👀",
        count: 1,
      },
    ],
  },
  {
    id: "product-20261013-09",
    author: SARAH,
    date: "2026-10-13",
    time: "16:12",
    text: "Updated the copy in the draft. Thanks for catching the difference between an empty team and an empty personal list.",
    reactions: [
      {
        emoji: "🙌",
        count: 2,
      },
    ],
  },
  {
    id: "product-20261014-01",
    author: SARAH,
    date: "2026-10-14",
    time: "08:55",
    text: "Today I’m looking at how team priorities sit beside the welcome content. I want the page to remain useful after the first visit.",
  },
  {
    id: "product-20261014-02",
    author: DANIEL,
    date: "2026-10-14",
    time: "09:01",
    text: "Could the welcome section shrink once someone has taken the first step? The team overview can then carry the page.",
  },
  {
    id: "product-20261014-03",
    author: PREVIEW_YOU,
    date: "2026-10-14",
    time: "09:16",
    text: "Yes, as long as we don’t move everything while they are using it. A settled state on the next visit would feel calmer.",
    reactions: [
      {
        emoji: "👍",
        count: 2,
      },
    ],
  },
  {
    id: "product-20261014-04",
    author: SARAH,
    date: "2026-10-14",
    time: "10:14",
    text: "I’ve tried three priority rows rather than a large card for each one. It leaves more room for task titles and owners.",
  },
  {
    id: "product-20261014-05",
    author: DANIEL,
    date: "2026-10-14",
    time: "11:32",
    text: "That helps with longer titles too. We should keep the status and due date readable without making every row taller.",
  },
  {
    id: "product-20261014-06",
    author: PREVIEW_YOU,
    date: "2026-10-14",
    time: "13:05",
    text: "@Maya can you recap the decisions about first visits and returning visits? Keep the open questions separate.",
  },
  {
    id: "product-20261014-07",
    author: MAYA,
    date: "2026-10-14",
    time: "13:06",
    text: "Decisions so far:\n• Keep the first visit focused on a clear next step.\n• Give returning teammates more room for current team work.\n• Keep the layout stable during a visit.\n\nStill open: what marks the welcome step as complete, and whether people can bring the introduction back.",
  },
  {
    id: "product-20261014-08",
    author: SARAH,
    date: "2026-10-14",
    time: "16:22",
    text: "That is the right split. I’ll bring both states to tomorrow’s review, including a way to reopen the introduction.",
    reactions: [
      {
        emoji: "✅",
        count: 2,
      },
      {
        emoji: "🙌",
        count: 1,
      },
    ],
  },
  {
    id: "product-20261015-01",
    author: SARAH,
    date: "2026-10-15",
    time: "09:03",
    text: "The two page states are ready for a review. I’ve kept the same navigation and spacing so the change feels familiar.",
  },
  {
    id: "product-20261015-02",
    author: DANIEL,
    date: "2026-10-15",
    time: "09:07",
    text: "I’ll take the keyboard walkthrough this time. Starting from the team switcher, then through the welcome action and priority list.",
  },
  {
    id: "product-20261015-03",
    author: SARAH,
    date: "2026-10-15",
    time: "09:22",
    text: "Keeping the FortyOne reference here while we review the wording. The experience should use the same straightforward language throughout.",
  },
  {
    id: "product-20261015-04",
    author: DANIEL,
    date: "2026-10-15",
    time: "10:40",
    text: "The welcome action is easy to reach, but the priority cards need a clearer focus state. Can we check that before the next pass?",
    replies: [
      {
        id: "product-20261015-04-reply-1",
        author: PREVIEW_YOU,
        date: "2026-10-15",
        time: "10:42",
        text: "Yes. Keep the same focus ring we use elsewhere so it is familiar. The whole interactive area should be obvious.",
      },
      {
        id: "product-20261015-04-reply-2",
        author: SARAH,
        date: "2026-10-15",
        time: "10:48",
        text: "I’ll check the light and dark versions together. The ring needs enough separation from the card border.",
      },
      {
        id: "product-20261015-04-reply-3",
        author: MAYA,
        date: "2026-10-15",
        time: "10:50",
        text: "Here is a task preview for that follow-up. The scope includes keyboard navigation and both themes.",
        tasks: [
          {
            id: "PRO-121",
            title: "Make keyboard focus clear on team priority cards",
          },
        ],
      },
    ],
  },
  {
    id: "product-20261015-05",
    author: PREVIEW_YOU,
    date: "2026-10-15",
    time: "10:43",
    text: "Also check what happens when the last item disappears from a filtered list. Focus should land somewhere useful, not get lost.",
    reactions: [
      {
        emoji: "👀",
        count: 2,
      },
    ],
  },
  {
    id: "product-20261015-06",
    author: SARAH,
    date: "2026-10-15",
    time: "11:58",
    text: "Copy pass is done for the draft.\n\nI shortened the introduction, made the empty states more specific, and removed a second action that was competing with the main one.",
  },
  {
    id: "product-20261015-07",
    author: DANIEL,
    date: "2026-10-15",
    time: "13:24",
    text: "The smaller-screen layout feels better with that shorter introduction. There is room to see the first item without scrolling.",
  },
  {
    id: "product-20261015-08",
    author: PREVIEW_YOU,
    date: "2026-10-15",
    time: "15:03",
    text: "For the next discussion, let’s focus on the welcome checklist and team overview as separate pieces. That should make the handoff easier.",
  },
  {
    id: "product-20261015-09",
    author: MAYA,
    date: "2026-10-15",
    time: "16:41",
    text: "Review recap:\n• The first action is clearer and the empty-state copy is more specific.\n• Team priorities have more room on smaller screens.\n• Keyboard focus needs one more pass.\n\nNext discussion: split the welcome checklist from the team overview, then agree on a small test.",
    reactions: [
      {
        emoji: "✅",
        count: 3,
      },
    ],
  },
];

const PRODUCT_LATEST: PreviewMessage[] = [
  {
    id: "product-today-01",
    author: DANIEL,
    date: "2026-10-16",
    time: "10:03",
    text: "The split works for me. I can review the team overview while Sarah takes the checklist.",
  },
  {
    id: "product-today-02",
    author: SARAH,
    date: "2026-10-16",
    time: "10:06",
    text: "Great. I’ll keep the checklist to three steps and bring both the empty and populated versions to the walkthrough.",
    reactions: [
      {
        emoji: "🙌",
        count: 2,
      },
    ],
  },
  {
    id: "product-today-03",
    author: PREVIEW_YOU,
    date: "2026-10-16",
    time: "10:11",
    text: "One last detail for the handoff: include the long-title, missing-photo, and narrow-screen examples. Those are useful cases to review together.",
  },
  {
    id: "product-today-04",
    author: DANIEL,
    date: "2026-10-16",
    time: "11:20",
    text: "I have the review cases lined up.\n\nThe overview should still make sense when there is one priority, when there are several, and when the team has not added any yet.",
  },
  {
    id: "product-today-05",
    author: SARAH,
    date: "2026-10-16",
    time: "11:24",
    text: "Perfect. Let’s leave any new feedback in the relevant thread so the decisions stay easy to find.",
    reactions: [
      {
        emoji: "👍",
        count: 1,
      },
    ],
  },
  {
    id: "product-today-06",
    author: MAYA,
    date: "2026-10-16",
    time: "12:05",
    text: "Next up: Sarah has the welcome checklist, Daniel has the team overview, and the review will include empty states, long titles, missing photos, and a narrow screen. The keyboard-focus follow-up remains in scope.",
  },
];

// Deliberately separate sample content from future chat queries and mutations.
export const PREVIEW_CONVERSATIONS: PreviewConversation[] = [
  {
    id: "product",
    name: "Product",
    kind: "team",
    color: "#6366f1",
    description: "Ideas, decisions, and what we’re building next.",
    members: [SARAH, DANIEL, PREVIEW_YOU, MAYA],
    messages: [
      ...PRODUCT_HISTORY,
      {
        id: "product-1",
        date: PREVIEW_TODAY,
        author: SARAH,
        time: "09:41",
        text: "Morning team! I’ve been thinking about the first-time experience. Can we make it easier for someone to find their next step?",
      },
      {
        id: "product-2",
        date: PREVIEW_TODAY,
        author: DANIEL,
        time: "09:43",
        text: "Yes. A short welcome checklist would help. I’d also bring the team’s current priorities into the overview, so people can see where to contribute.",
        reactions: [{ emoji: "🙌", count: 2 }],
      },
      {
        id: "product-3",
        date: PREVIEW_TODAY,
        author: SARAH,
        time: "09:45",
        text: "Let’s keep the checklist to three steps. We can test the flow with a few new teammates before rolling it out.",
      },
      {
        id: "product-4",
        date: PREVIEW_TODAY,
        author: PREVIEW_YOU,
        time: "09:47",
        text: "@Maya can you turn this discussion into tasks? Keep the welcome checklist and team overview separate.",
        replies: [
          {
            id: "product-maya",
            date: PREVIEW_TODAY,
            author: MAYA,
            time: "09:47",
            text: "Two clear next steps from this conversation. I’ve kept the three-step limit and the small-group test in the scope, and left ownership open for the team.",
            tasks: [
              { id: "PRO-128", title: "Design a three-step welcome checklist" },
              {
                id: "PRO-129",
                title: "Bring team priorities into the overview",
              },
            ],
          },
          {
            id: "product-reply",
            date: PREVIEW_TODAY,
            author: SARAH,
            time: "09:49",
            text: "Looks good. I can take the checklist and share a first pass here.",
          },
        ],
      },
      ...PRODUCT_LATEST,
    ],
  },
  {
    id: "engineering",
    name: "Engineering",
    kind: "team",
    color: "#16a34a",
    description: "Shipping, technical decisions, and helping each other out.",
    members: [DANIEL, PREVIEW_YOU, MAYA],
    messages: [
      {
        id: "engineering-1",
        author: DANIEL,
        time: "10:12",
        text: "The latest build is ready for a walkthrough. Let’s collect anything that feels slow or unclear in this thread.",
      },
      {
        id: "engineering-2",
        author: PREVIEW_YOU,
        time: "10:15",
        text: "I’ll try the onboarding flow on a smaller screen and share what I find.",
        reactions: [{ emoji: "👍", count: 2 }],
      },
    ],
  },
  {
    id: "design",
    name: "Design",
    kind: "team",
    color: "#a855f7",
    description:
      "Work in progress, thoughtful feedback, and the little details.",
    members: [SARAH, PREVIEW_YOU, MAYA],
    messages: [
      {
        id: "design-1",
        author: SARAH,
        time: "08:52",
        text: "A small direction for the welcome flow: fewer choices up front, a clear next step, and more room for the actual work. I’ll share the screens here when they’re ready.",
      },
    ],
  },
  {
    id: "sarah",
    name: SARAH.name,
    kind: "dm",
    description: "Direct message",
    members: [SARAH, PREVIEW_YOU],
    messages: [
      {
        id: "sarah-1",
        author: SARAH,
        time: "09:52",
        text: "Do you have a few minutes to look at the welcome checklist together? I’d love a second pair of eyes before sharing it with everyone.",
      },
      {
        id: "sarah-2",
        author: PREVIEW_YOU,
        time: "09:54",
        text: "Of course. Send over the first pass when you’re ready.",
      },
    ],
  },
  {
    id: "daniel",
    name: DANIEL.name,
    kind: "dm",
    description: "Direct message",
    members: [DANIEL, PREVIEW_YOU],
    messages: [
      {
        id: "daniel-1",
        author: DANIEL,
        date: "2026-10-15",
        time: "16:24",
        text: "Thanks for the walkthrough earlier. The smaller-screen feedback was really helpful. I’ll pick those improvements up next.",
      },
    ],
  },
];
