import type { MayaSkillInput } from "./types";

export const MAYA_SKILL_STARTERS: MayaSkillInput[] = [
  {
    name: "Triage incoming work",
    description: "Review new requests and suggest clear next steps.",
    instructions:
      "Review new incoming requests for my team. Check the available details and related work before recommending a priority, owner, or next step. Identify possible duplicates and explain your recommendations briefly. Draft any suggested tasks for my review; ask only for information that is needed to make a decision.",
  },
  {
    name: "Plan my week",
    description: "Choose priorities that fit your deadlines and capacity.",
    instructions:
      "Help me plan this week using my open work, upcoming deadlines, dependencies, and available capacity. Recommend up to three priorities, explain the tradeoffs, and flag anything likely to slip. When assignment or protected calendar time would help, offer a work-plan preview before making changes. Do not invent effort estimates or reserve calendar time without my approval.",
  },
  {
    name: "Draft a weekly update",
    description: "Summarize progress, risks, and what comes next.",
    instructions:
      "Draft a weekly update for my team covering the last seven days. Use current workspace evidence to describe completed work, progress toward goals, blockers or risks, and next steps. Keep the update concise and useful for teammates. Distinguish confirmed facts from missing information. Show the draft here for review before publishing or posting it anywhere.",
  },
];

export const insertMayaSkillInstructions = (
  draft: string,
  instructions: string,
) =>
  !draft.trim() || draft.trim() === "/"
    ? instructions
    : `${draft}\n\n${instructions}`;
