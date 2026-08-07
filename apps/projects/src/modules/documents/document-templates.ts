import type { DocumentCreate } from "./types";

export type DocumentTemplateIcon =
  | "blank"
  | "meeting"
  | "project"
  | "one-to-one"
  | "update";

export type DocumentTemplate = Required<
  Pick<DocumentCreate, "title" | "contentHtml" | "contentText">
> & {
  id: string;
  icon: DocumentTemplateIcon;
  label: string;
};

export const documentTemplates: readonly DocumentTemplate[] = [
  {
    id: "blank",
    icon: "blank",
    label: "Blank document",
    title: "Untitled document",
    contentHtml: "",
    contentText: "",
  },
  {
    id: "meeting-notes",
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
];
