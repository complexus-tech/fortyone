import assert from "node:assert/strict";
import test from "node:test";
import type { ObjectiveActivity } from "./detail-data";
import { objectiveActivityRows } from "./activity-rows";

const activity: ObjectiveActivity = {
  id: "activity-1",
  type: "update",
  updateType: "objective",
  field: "comment",
  currentValue: "",
  comment: "**Ready** for review",
  userId: "member-1",
  createdAt: "2026-09-16T12:00:00Z",
};
const context = {
  members: [{ id: "member-1", fullName: "Rudo" }],
  statuses: [{ id: "status-1", name: "In progress" }],
  objectiveTerm: "objective",
  keyResultTerm: "key result",
};

test("standalone objective comments appear as Markdown comments, not property updates", () => {
  const rows = objectiveActivityRows({
    ...context,
    activities: [activity],
    tab: "comments",
  });
  assert.equal(rows[0].format, "markdown");
  assert.equal(rows[0].body, "**Ready** for review");
  assert.equal(rows[0].author, "Rudo");
  assert.deepEqual(
    objectiveActivityRows({
      ...context,
      activities: [activity],
      tab: "updates",
    }),
    [],
  );
});

test("legacy health comments retain literal text", () => {
  const rows = objectiveActivityRows({
    ...context,
    activities: [{ ...activity, field: "health" }],
    tab: "comments",
  });
  assert.equal(rows[0].format, "plain");
});

test("objective activity resolves status and lead names and does not expose description markup", () => {
  const rows = objectiveActivityRows({
    ...context,
    tab: "updates",
    activities: [
      { ...activity, field: "status_id", currentValue: "status-1" },
      { ...activity, field: "lead_user_id", currentValue: "member-1" },
      {
        ...activity,
        field: "description",
        currentValue: "<p>Private details</p>",
      },
      { ...activity, field: "status_id", currentValue: "unknown-status-id" },
    ],
  });
  assert.deepEqual(
    rows.map((row) => row.body),
    [
      "updated status to In progress",
      "updated lead to Rudo",
      "updated description",
      "updated status",
    ],
  );
});
