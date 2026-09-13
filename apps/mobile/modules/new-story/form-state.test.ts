import assert from "node:assert/strict";
import test from "node:test";
import { formReducer, initialState, isFormState } from "./form-state";

test("changing team clears team-owned selections without losing content or submission identity", () => {
  const draft = {
    ...initialState,
    title: "Review onboarding",
    description: {
      html: "<p>Keep this draft</p>",
      text: "Keep this draft",
      mentions: [],
    },
    idempotencyKey: "same-story-on-retry",
    teamId: "old-team",
    statusId: "old-status",
    assigneeId: "old-member",
    labelIds: ["old-label"],
    activeSheet: "team" as const,
  };
  const next = formReducer(draft, { type: "setTeam", teamId: "new-team" });
  assert.deepEqual(next, {
    ...draft,
    teamId: "new-team",
    statusId: null,
    assigneeId: null,
    labelIds: [],
    activeSheet: null,
  });
  assert.equal(draft.teamId, "old-team");
});

test("label selection stays open for multiple choices and removes only the toggled label", () => {
  const draft = {
    ...initialState,
    labelIds: ["one"],
    activeSheet: "labels" as const,
  };
  const withTwo = formReducer(draft, { type: "toggleLabel", labelId: "two" });
  const withoutOne = formReducer(withTwo, {
    type: "toggleLabel",
    labelId: "one",
  });
  assert.deepEqual(withoutOne.labelIds, ["two"]);
  assert.equal(withoutOne.activeSheet, "labels");
  assert.deepEqual(draft.labelIds, ["one"]);
});

test("restored drafts require valid rich content, team selections, and priority", () => {
  assert.equal(
    isFormState({ ...initialState, idempotencyKey: "retry-key" }),
    true,
  );
  assert.equal(
    isFormState({ ...initialState, description: { html: "<p>Text</p>" } }),
    false,
  );
  assert.equal(isFormState({ ...initialState, statusId: 12 }), false);
  assert.equal(isFormState({ ...initialState, priority: "Unknown" }), false);
  assert.equal(isFormState({ ...initialState, activeSheet: "missing" }), false);
});
