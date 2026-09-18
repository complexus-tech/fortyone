import type { DocumentCreate } from "./types";

export type DocumentTemplateIcon =
  | "blank"
  | "meeting"
  | "project"
  | "one-to-one"
  | "update"
  | "guide"
  | "research"
  | "event"
  | "checklist";

export type DocumentTemplateCategory =
  | "planning"
  | "teamwork"
  | "knowledge"
  | "campaigns";

export const documentTemplateCategories: readonly {
  id: DocumentTemplateCategory;
  label: string;
}[] = [
  { id: "planning", label: "Planning & strategy" },
  { id: "teamwork", label: "Teamwork" },
  { id: "knowledge", label: "Knowledge & processes" },
  { id: "campaigns", label: "Campaigns & events" },
];

export type DocumentTemplate = Required<
  Pick<DocumentCreate, "title" | "contentHtml" | "contentText">
> & {
  id: string;
  category: DocumentTemplateCategory | null;
  icon: DocumentTemplateIcon;
  label: string;
};

type TemplateSection = { heading: string; prompt: string };

function createTeamTemplate(
  id: string,
  label: string,
  icon: DocumentTemplateIcon,
  category: DocumentTemplateCategory,
  introduction: string,
  sections: readonly TemplateSection[],
): DocumentTemplate {
  return {
    id,
    label,
    title: label,
    icon,
    category,
    contentHtml: [
      `<p>${introduction}</p>`,
      ...sections.map(
        ({ heading, prompt }) => `<h2>${heading}</h2><p>${prompt}</p>`,
      ),
    ].join("\n"),
    contentText: [
      introduction,
      ...sections.flatMap(({ heading, prompt }) => [heading, prompt]),
    ].join("\n"),
  };
}

export const documentTemplates: readonly DocumentTemplate[] = [
  {
    id: "blank",
    category: null,
    icon: "blank",
    label: "Blank document",
    title: "Untitled document",
    contentHtml: "",
    contentText: "",
  },
  {
    id: "meeting-notes",
    category: "teamwork",
    icon: "meeting",
    label: "Meeting notes",
    title: "Meeting notes",
    contentHtml: `
      <p>Keep the purpose, discussion, and follow-up from this meeting in one place.</p>
      <h2>Meeting details</h2>
      <table>
        <tbody>
          <tr><th>Date</th><td>Add date</td></tr>
          <tr><th>Participants</th><td>Add participants</td></tr>
          <tr><th>Purpose</th><td>What should this meeting accomplish?</td></tr>
        </tbody>
      </table>
      <h2>Agenda</h2>
      <ul><li>Add the first topic</li><li>Add another topic</li></ul>
      <h2>Notes</h2>
      <p>Capture the important context and discussion here.</p>
      <h2>Decisions</h2>
      <ul><li>Add a decision</li></ul>
      <h2>Action items</h2>
      <ul data-type="taskList"><li data-type="taskItem" data-checked="false"><label><input type="checkbox"><span></span></label><div><p>Add an owner and next step</p></div></li></ul>
    `,
    contentText: `Keep the purpose, discussion, and follow-up from this meeting in one place.
Meeting details
Date
Add date
Participants
Add participants
Purpose
What should this meeting accomplish?
Agenda
Add the first topic
Add another topic
Notes
Capture the important context and discussion here.
Decisions
Add a decision
Action items
Add an owner and next step`,
  },
  {
    id: "project-brief",
    category: "planning",
    icon: "project",
    label: "Project brief",
    title: "Project brief",
    contentHtml: `
      <p>Give everyone a clear view of what this project is for and how the work will move forward.</p>
      <h2>Overview</h2>
      <p>Describe the opportunity, problem, or change this project addresses.</p>
      <h2>Desired outcomes</h2>
      <ul><li>Add an outcome</li><li>Add another outcome</li></ul>
      <h2>Scope</h2>
      <table>
        <tbody>
          <tr><th>Included</th><td>What is part of this project?</td></tr>
          <tr><th>Not included</th><td>What is intentionally outside the scope?</td></tr>
        </tbody>
      </table>
      <h2>Plan</h2>
      <ol><li>Add the first milestone</li><li>Add the next milestone</li></ol>
      <h2>Risks and dependencies</h2>
      <ul><li>Add a risk or dependency</li></ul>
      <h2>Next steps</h2>
      <ul data-type="taskList"><li data-type="taskItem" data-checked="false"><label><input type="checkbox"><span></span></label><div><p>Add an owner and next step</p></div></li></ul>
    `,
    contentText: `Give everyone a clear view of what this project is for and how the work will move forward.
Overview
Describe the opportunity, problem, or change this project addresses.
Desired outcomes
Add an outcome
Add another outcome
Scope
Included
What is part of this project?
Not included
What is intentionally outside the scope?
Plan
Add the first milestone
Add the next milestone
Risks and dependencies
Add a risk or dependency
Next steps
Add an owner and next step`,
  },
  {
    id: "one-to-one",
    category: "teamwork",
    icon: "one-to-one",
    label: "One-to-one",
    title: "One-to-one",
    contentHtml: `
      <p>Use this shared space for a focused, useful conversation.</p>
      <h2>Check-in</h2>
      <p>How are things going right now?</p>
      <h2>Recent wins</h2>
      <ul><li>Add a highlight or achievement</li></ul>
      <h2>Discussion topics</h2>
      <ul><li>Add a topic</li><li>Add another topic</li></ul>
      <h2>Support needed</h2>
      <p>Capture decisions, feedback, or help that would make a difference.</p>
      <h2>Follow-up</h2>
      <ul data-type="taskList"><li data-type="taskItem" data-checked="false"><label><input type="checkbox"><span></span></label><div><p>Add a follow-up action</p></div></li></ul>
    `,
    contentText: `Use this shared space for a focused, useful conversation.
Check-in
How are things going right now?
Recent wins
Add a highlight or achievement
Discussion topics
Add a topic
Add another topic
Support needed
Capture decisions, feedback, or help that would make a difference.
Follow-up
Add a follow-up action`,
  },
  {
    id: "weekly-update",
    category: "teamwork",
    icon: "update",
    label: "Weekly update",
    title: "Weekly update",
    contentHtml: `
      <p>Share progress, priorities, and anything that needs attention.</p>
      <h2>Summary</h2>
      <p>Give a short overview of the week.</p>
      <h2>Progress</h2>
      <ul><li>Add a completed or meaningful piece of work</li></ul>
      <h2>Next priorities</h2>
      <ol><li>Add the most important priority</li><li>Add another priority</li></ol>
      <h2>Blockers and decisions</h2>
      <ul><li>Add a blocker, risk, or decision needed</li></ul>
      <h2>Key measures</h2>
      <table>
        <tbody>
          <tr><th>Measure</th><th>Current</th><th>Previous</th></tr>
          <tr><td>Add a measure</td><td>—</td><td>—</td></tr>
        </tbody>
      </table>
    `,
    contentText: `Share progress, priorities, and anything that needs attention.
Summary
Give a short overview of the week.
Progress
Add a completed or meaningful piece of work
Next priorities
Add the most important priority
Add another priority
Blockers and decisions
Add a blocker, risk, or decision needed
Key measures
Measure
Current
Previous
Add a measure
—
—`,
  },
  createTeamTemplate(
    "proposal",
    "Proposal",
    "project",
    "planning",
    "Make the case for an idea and the decision needed to move it forward.",
    [
      {
        heading: "Opportunity",
        prompt:
          "What problem or opportunity are we addressing, and who benefits?",
      },
      {
        heading: "Proposed approach",
        prompt: "Describe what you recommend and the alternatives considered.",
      },
      {
        heading: "Budget and resources",
        prompt: "Estimate costs, people, and time required.",
      },
      {
        heading: "Delivery plan",
        prompt: "Outline milestones, owners, and key risks.",
      },
      {
        heading: "Approval",
        prompt: "Who needs to decide, by when, and what are the next steps?",
      },
    ],
  ),
  createTeamTemplate(
    "campaign-plan",
    "Campaign plan",
    "update",
    "campaigns",
    "Plan a campaign around a clear audience, message, and outcome.",
    [
      {
        heading: "Goal and audience",
        prompt: "What should change, and who are we trying to reach?",
      },
      {
        heading: "Message",
        prompt:
          "Write the core message and the action you want people to take.",
      },
      {
        heading: "Channels and activities",
        prompt:
          "Choose channels, deliverables, and an owner for each activity.",
      },
      {
        heading: "Schedule and budget",
        prompt: "Set launch dates, review deadlines, and spending limits.",
      },
      {
        heading: "Success measures",
        prompt: "Choose the measures, targets, and review date.",
      },
    ],
  ),
  createTeamTemplate(
    "process-guide",
    "Process guide",
    "guide",
    "knowledge",
    "Document a repeatable way to get work done.",
    [
      {
        heading: "Purpose and scope",
        prompt: "When should someone use this process, and what does it cover?",
      },
      {
        heading: "Before you start",
        prompt: "List the access, information, tools, and approvals needed.",
      },
      {
        heading: "Steps",
        prompt: "Describe each step in order, including who is responsible.",
      },
      {
        heading: "Checks and exceptions",
        prompt:
          "How do you verify the result, and what happens if something goes wrong?",
      },
      {
        heading: "Ownership and review",
        prompt: "Name the process owner and the next review date.",
      },
    ],
  ),
  createTeamTemplate(
    "research-summary",
    "Research summary",
    "research",
    "knowledge",
    "Turn evidence into findings the team can act on.",
    [
      {
        heading: "Research questions",
        prompt: "What did we need to learn, and why?",
      },
      {
        heading: "Method and sources",
        prompt:
          "Describe how evidence was collected, the sources used, and any limitations.",
      },
      {
        heading: "Key findings",
        prompt:
          "Summarise what you learned and the evidence supporting each finding.",
      },
      {
        heading: "Recommendations",
        prompt: "Explain what the findings suggest we should do.",
      },
      {
        heading: "Open questions",
        prompt: "What remains uncertain, and what should we investigate next?",
      },
    ],
  ),
  createTeamTemplate(
    "onboarding-plan",
    "Onboarding plan",
    "checklist",
    "knowledge",
    "Help someone settle in with the right people, context, and first steps.",
    [
      {
        heading: "Welcome and contacts",
        prompt: "Introduce the role, manager, buddy, and key people.",
      },
      {
        heading: "Before day one",
        prompt:
          "List access, equipment, documents, and arrangements to prepare.",
      },
      {
        heading: "First week",
        prompt:
          "Plan introductions, essential learning, and a manageable first task.",
      },
      {
        heading: "First month",
        prompt: "Agree on priorities, milestones, and the support available.",
      },
      {
        heading: "Check-ins",
        prompt:
          "Schedule conversations to review progress, questions, and feedback.",
      },
    ],
  ),
  createTeamTemplate(
    "event-plan",
    "Event plan",
    "event",
    "campaigns",
    "Keep the programme, people, and practical arrangements in one place.",
    [
      {
        heading: "Event overview",
        prompt:
          "Define the purpose, audience, date, location, and expected attendance.",
      },
      {
        heading: "Programme",
        prompt: "Outline sessions, timings, speakers, and activity owners.",
      },
      {
        heading: "Logistics",
        prompt:
          "Plan the venue, registration, equipment, catering, and accessibility.",
      },
      {
        heading: "Budget and promotion",
        prompt:
          "Estimate costs and explain how people will hear about the event.",
      },
      {
        heading: "Readiness and follow-up",
        prompt:
          "Record contingency plans, final checks, feedback, and follow-up owners.",
      },
    ],
  ),
  createTeamTemplate(
    "team-goals",
    "Team goals",
    "project",
    "planning",
    "Agree on what matters and how the team will recognise progress.",
    [
      {
        heading: "Time period and context",
        prompt:
          "Which period are these goals for, and what should guide our choices?",
      },
      {
        heading: "Priorities",
        prompt: "Describe the few outcomes that matter most.",
      },
      {
        heading: "Success measures",
        prompt: "Give each outcome a starting point, target, and due date.",
      },
      {
        heading: "Responsibilities and support",
        prompt:
          "Name an owner for each goal and the dependencies or resources needed.",
      },
      {
        heading: "Review rhythm",
        prompt:
          "Set check-in dates and record progress, decisions, and adjustments.",
      },
    ],
  ),
  createTeamTemplate(
    "retrospective",
    "Retrospective",
    "meeting",
    "teamwork",
    "Reflect on recent work and agree on practical improvements.",
    [
      {
        heading: "Context",
        prompt:
          "Which period, project, or event are we reviewing, and who is taking part?",
      },
      {
        heading: "What worked",
        prompt: "Capture successes and practices worth continuing.",
      },
      {
        heading: "What was difficult",
        prompt:
          "Describe challenges and contributing factors without assigning blame.",
      },
      {
        heading: "What we learned",
        prompt: "Record lessons and ideas to try next time.",
      },
      {
        heading: "Actions",
        prompt:
          "Choose a small number of improvements, each with an owner and review date.",
      },
    ],
  ),
];
