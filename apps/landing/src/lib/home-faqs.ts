export type HomeFaq = {
  question: string;
  answer: string;
};

export const homeFaqs: HomeFaq[] = [
  {
    question: "What makes FortyOne different from a task manager?",
    answer:
      "Most task managers start after the work has already been chosen. FortyOne starts with the request: customers submit and vote on feedback, your team decides what should move forward, planned feedback becomes linked project work, and customers can follow delivery on the public roadmap. The original request stays attached, so the team never loses why the work exists.",
  },
  {
    question: "How does customer feedback work?",
    answer:
      "Customers can submit requests, vote, and comment in a public portal. Feedback boards route each request to the team that owns the work. The team can review it, close it, or turn it into planned work.",
  },
  {
    question: "What can FortyOne's AI do?",
    answer:
      "AI can suggest an owner, fill in an estimate, find time for the work, and surface delivery risk from the project context already in FortyOne.",
  },
  {
    question: "Can we review AI actions before they are applied?",
    answer:
      "Yes. Teams can review and edit important suggestions before they change the project plan.",
  },
  {
    question: "Is the free plan actually free?",
    answer:
      "Yes. There is no credit card and no trial expiry. The Hobby plan supports one team and up to five members, enough to run a real sprint and decide whether FortyOne should scale with you.",
  },
  {
    question: "Can FortyOne plan around my team's calendar?",
    answer:
      "Yes. Google Calendar integration lets FortyOne sync availability so AI can recommend better schedules and work windows without storing private event details unnecessarily.",
  },
  {
    question: "Which tools does FortyOne work with?",
    answer:
      "FortyOne works with tools including Google Calendar, Slack, GitHub, GitLab, Figma, and Google Drive. These connections bring availability and source context into the project plan.",
  },
];
