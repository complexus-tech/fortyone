export type HomeFaq = {
  question: string;
  answer: string;
};

export const homeFaqs: HomeFaq[] = [
  {
    question: "What makes FortyOne different from a task manager?",
    answer:
      "Most task managers start after the work has already been chosen. FortyOne connects the customer request, company strategy, project documents, calendar plan, and delivery work, so the team can trace why something matters from the first signal to the finished result.",
  },
  {
    question: "How does customer feedback work?",
    answer:
      "Customers can submit requests, vote, and comment in a public portal. Feedback boards route each request to the team that owns the work. The team can review it, close it, or turn it into planned work.",
  },
  {
    question: "What can FortyOne's AI do?",
    answer:
      "Maya can prepare a delivery brief, suggest an owner and estimate, propose the first work window, and surface delivery risk from the project context already in FortyOne.",
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
      "Yes. Google Calendar meetings and FortyOne work can appear in one weekly view. Teams can schedule assigned work around existing commitments, protect focus time, and spot conflicts. When Maya auto-assignment is enabled, Maya can choose an owner and place the first work block using workload and calendar availability.",
  },
  {
    question: "How does FortyOne connect strategy to daily work?",
    answer:
      "Start with an ultimate goal, organise the strategy into pillars, and connect each objective to measurable key results. Stories can then link to the objective and key result they support, while the roadmap shows ownership, health, progress, and dates.",
  },
  {
    question: "Can documents stay connected to project work?",
    answer:
      "Yes. Teams can write shared project documents, attach Stories and Objectives through Related Work, and turn selected text into a new Story or Objective without losing the surrounding context.",
  },
  {
    question: "Which tools does FortyOne work with?",
    answer:
      "FortyOne works with tools including Google Calendar, Slack, GitHub, GitLab, Figma, and Google Drive. These connections bring availability and source context into the project plan.",
  },
];
