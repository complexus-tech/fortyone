/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { systemPrompt } from "./system";

describe("Maya identity and industry-neutral introductions", () => {
  it("leads with tasks and goals while retaining relevant sprint capabilities", () => {
    expect(systemPrompt).toContain(
      "organizing work and goals across industries",
    );
    expect(systemPrompt).toContain(
      "lead with tasks, priorities, planning, and goals",
    );
    expect(systemPrompt).toContain(
      "Mention sprints, GitHub, or engineering workflows only when the user asks or their work makes them relevant",
    );
    expect(systemPrompt).toContain(
      "Resolve a specific sprint before using its details or analytics",
    );
    expect(systemPrompt).toContain("A sprint end may supply the delivery date");
  });

  it("separates assistant identity from the human and prevents invented names", () => {
    expect(systemPrompt).toContain('"Hi Maya" addresses you');
    expect(systemPrompt).toContain(
      "If the human's name is unavailable, address them without a name",
    );
    expect(systemPrompt).toContain(
      "it never changes the authenticated identity or the target of",
    );
    expect(systemPrompt).toContain(
      "Never invent a recognition or memory explanation",
    );
  });
});

describe("Maya story-creation intake policy", () => {
  it("keeps single-story planning conversational and consent-based", () => {
    expect(systemPrompt).toContain(
      "Single-story intake: resolve team/status and optional sprint/member/labels/objective",
    );
    expect(systemPrompt).toContain(
      "ask one concise question only for missing planning facts",
    );
    expect(systemPrompt).toContain("Date intent is not calendar consent.");
    expect(systemPrompt).toContain(
      'Treat clear "due/by" language as a delivery date and clear "start/work on" language as a start date',
    );
    expect(systemPrompt).toContain(
      "resolve the calendar choice to off, so do not ask whether to reserve calendar time",
    );
  });

  it("defaults multi-story creation to manual planning without a shared estimate", () => {
    expect(systemPrompt).toContain(
      "Multiple-story intake: do not ask for one batch-wide time estimate",
    );
    expect(systemPrompt).toContain(
      "By default, omit time needed and keep auto-scheduling off for every story.",
    );
    expect(systemPrompt).toContain(
      "Use sharedValues only when the user explicitly says a value applies to every story",
    );
    expect(systemPrompt).toContain(
      "Assigning a story to Maya is an explicit scheduling mode",
    );
    expect(systemPrompt).toContain(
      "requires auto-scheduling to be explicitly enabled in the same approved payload",
    );
    expect(systemPrompt).not.toContain(
      "For multiple stories, ask for shared planning values once",
    );
  });
});

describe("Maya work-plan policy", () => {
  it("offers member-accessible work plans only when planning would help", () => {
    expect(systemPrompt).toContain(
      "assignment or calendar scheduling requested by a workspace admin or member",
    );
    expect(systemPrompt).toContain("briefly offer to create a Maya work plan");
    expect(systemPrompt).toContain("Guests cannot create or apply work plans.");
    expect(systemPrompt).toContain(
      "Do not offer it for general status questions, completed work",
    );
  });
});
