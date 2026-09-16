import assert from "node:assert/strict";
import test from "node:test";
import { getTeamTabs, nextTeamPage, resolveTeamTab } from "./team-tabs.ts";

test("team tabs use workspace terminology and hide unavailable sections", () => {
  assert.deepEqual(
    getTeamTabs("work items", "Goals", {
      objectives: false,
      feedback: false,
      intake: false,
    }),
    [{ value: "all", label: "All work items" }],
  );
  assert.deepEqual(
    getTeamTabs("tickets", "Projects", {
      objectives: true,
      feedback: true,
      intake: true,
    }),
    [
      { value: "all", label: "All tickets" },
      { value: "objectives", label: "Projects" },
      { value: "feedback", label: "Feedback" },
      { value: "intake", label: "Intake" },
    ],
  );
});

test("each team feature is gated independently", () => {
  for (const section of ["objectives", "feedback", "intake"] as const) {
    const tabs = getTeamTabs("stories", "Objectives", {
      objectives: section === "objectives",
      feedback: section === "feedback",
      intake: section === "intake",
    });
    assert.deepEqual(
      tabs.map((tab) => tab.value),
      ["all", section],
    );
    assert.equal(resolveTeamTab(section, tabs), section);
  }
});

test("disabling or removing the selected section returns to all work", () => {
  const tabs = getTeamTabs("stories", "Objectives", {
    objectives: false,
    feedback: false,
    intake: false,
  });
  for (const selection of [
    "objectives",
    "feedback",
    "intake",
    "sprints",
    "active",
    "unknown",
  ])
    assert.equal(resolveTeamTab(selection, tabs), "all");
});

test("pagination advances only when the server supplies a later page", () => {
  assert.equal(nextTeamPage({ page: 1, nextPage: 2, hasMore: true }), 2);
  assert.equal(
    nextTeamPage({ page: 2, nextPage: 3, hasMore: false }),
    undefined,
  );
  assert.equal(
    nextTeamPage({ page: 2, nextPage: 2, hasMore: true }),
    undefined,
  );
  assert.equal(
    nextTeamPage({ page: 2, nextPage: 1, hasMore: true }),
    undefined,
  );
});
