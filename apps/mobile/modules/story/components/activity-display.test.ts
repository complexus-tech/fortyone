import type { StoryActivity, StoryActivityUser } from "@/modules/stories/types";
import type { ActivityDisplayContext } from "./activity-display.ts";
import assert from "node:assert/strict";
import test from "node:test";
import { FORMER_USER_ID } from "@/lib/former-user";
import {
  formatActivityTimestamp,
  formatScheduleActivityValue,
  getActivityDisplay,
  indexActivityPeople,
  indexActivityReferences,
  resolveActivityActor,
} from "./activity-display.ts";

const MEMBER_ID = "d1ae2e34-8e3a-4d36-8b6e-ade2bd49732a";
const MAYA_ID = "b77c3c1a-0d0c-4f5a-841b-389f1e7a6605";
const OTHER_ID = "fbd8adf1-136e-4c47-81f9-6f9aacb84933";
const STATUS_ID = "7f35cbb3-70ab-498a-9c8b-05f9e2e34d55";
const NEW_STATUS_ID = "3bd133d2-015c-4316-95e3-00fa48d4371a";
const REFERENCE_ID = "9f01260d-765c-4910-ae93-799c6d026db7";

const person: StoryActivityUser = {
  id: MEMBER_ID,
  fullName: "Jane Doe",
  username: "jane",
  avatarUrl: "",
  isActive: true,
  isSystem: false,
};
const maya: StoryActivityUser = {
  ...person,
  id: MAYA_ID,
  fullName: "Maya",
  username: "maya",
  isSystem: true,
};
const context: ActivityDisplayContext = {
  members: indexActivityReferences([person]),
  statuses: indexActivityReferences([
    { id: STATUS_ID, name: "Todo" },
    { id: NEW_STATUS_ID, name: "In progress" },
  ]),
  sprints: indexActivityReferences([{ id: REFERENCE_ID, name: "September" }]),
  objectives: indexActivityReferences([{ id: REFERENCE_ID, name: "Launch" }]),
  labels: indexActivityReferences([{ id: REFERENCE_ID, name: "Mobile" }]),
  currentUserId: OTHER_ID,
  timezone: "Africa/Harare",
  storyTerm: "task",
  sprintTerm: "sprint",
  objectiveTerm: "objective",
  keyResultTerm: "key result",
  estimateScheme: "tshirt",
};
const activity: StoryActivity = {
  id: "activity-1",
  storyId: "story-1",
  userId: MAYA_ID,
  user: maya,
  type: "update",
  field: "auto_scheduling_time",
  currentValue: "21 Aug 2026 at 08:23 UTC",
  oldValue: null,
  newValue: "2026-08-21T08:23:00Z",
  createdAt: "2026-08-20T12:00:00Z",
};
const display = (patch: Partial<StoryActivity>, config = context) =>
  getActivityDisplay({ ...activity, newValue: undefined, ...patch }, config);

test("retained history uses Former user even with missing members or a stale system summary", () => {
  assert.deepEqual(display({ userId: FORMER_USER_ID, user: null }).actor, {
    name: "Former user",
    isSystem: false,
  });
  assert.deepEqual(
    display({ userId: FORMER_USER_ID, user: { ...maya, id: FORMER_USER_ID } })
      .actor,
    {
      name: "Former user",
      isSystem: false,
    },
  );
  assert.equal(
    display({ field: "assignee_id", currentValue: FORMER_USER_ID }).message,
    "assigned the task to Former user",
  );
});

test("uses the API's embedded Maya actor even when absent from workspace members", () => {
  const result = getActivityDisplay(activity, context);
  assert.deepEqual(result.actor, { name: "Maya", isSystem: true });
  assert.equal(result.message, "scheduled work for 21 Aug 2026 at 10:23");
});

test("embedded identity remains authoritative for inactive or departed accounts", () => {
  assert.deepEqual(
    resolveActivityActor(
      {
        userId: MEMBER_ID,
        user: { ...person, fullName: "Jane Smith", isActive: false },
      },
      context,
    ),
    { name: "Jane Smith", isSystem: false },
  );
});

test("old cached rows use members; missing identities never become fabricated Maya actors", () => {
  assert.equal(
    display({ user: undefined, userId: MEMBER_ID }).actor.name,
    "Jane Doe",
  );
  assert.deepEqual(display({ user: null }).actor, {
    name: "Someone",
    isSystem: false,
  });
  assert.deepEqual(
    display(
      { userId: OTHER_ID, user: maya },
      {
        ...context,
        currentUserId: null,
      },
    ).actor,
    { name: "Someone", isSystem: false },
  );
});

test("self, username-only and unnamed system accounts have deliberate labels", () => {
  assert.equal(display({ userId: OTHER_ID, user: null }).actor.name, "You");
  assert.equal(display({ user: { ...maya, fullName: "" } }).actor.name, "maya");
  assert.equal(
    display({ user: { ...maya, username: "", fullName: "" } }).actor.name,
    "System",
  );
});

test("first scheduling and rescheduling use different verbs and typed timestamps", () => {
  const result = getActivityDisplay(
    { ...activity, oldValue: "2026-08-20T07:00:00Z" },
    context,
  );
  assert.equal(result.message, "rescheduled work to 21 Aug 2026 at 10:23");
  assert.equal(
    getActivityDisplay({ ...activity, oldValue: "<nil>" }, context).message,
    "scheduled work for 21 Aug 2026 at 10:23",
  );
});

test("schedule timestamps honor the profile timezone across date boundaries and DST", () => {
  assert.equal(
    formatScheduleActivityValue(
      "",
      "2026-09-15T01:30:00Z",
      "America/Los_Angeles",
    ),
    "14 Sept 2026 at 18:30",
  );
  assert.equal(
    formatScheduleActivityValue(
      "",
      "2026-01-15T01:30:00Z",
      "America/Los_Angeles",
    ),
    "14 Jan 2026 at 17:30",
  );
});

test("legacy schedule text survives missing timestamps and invalid profile settings", () => {
  for (const timezone of [undefined, "invalid/timezone"]) {
    assert.equal(
      formatScheduleActivityValue(
        "21 Aug 2026 at 10:23 CAT",
        activity.newValue,
        timezone,
      ),
      "21 Aug 2026 at 10:23",
    );
  }
  assert.equal(
    formatScheduleActivityValue(
      "21 Aug 2026 at 10:23 +0200",
      "invalid",
      "Africa/Harare",
    ),
    "21 Aug 2026 at 10:23",
  );
  assert.equal(
    formatScheduleActivityValue("<nil>", "invalid", "Africa/Harare"),
    "an unavailable time",
  );
});

test("scheduling status changes humanize both values and avoid fake transitions", () => {
  assert.equal(
    display({
      field: "auto_scheduling_status",
      currentValue: "At risk",
      oldValue: "scheduled",
    }).message,
    "changed auto-scheduling from Scheduled to At risk",
  );
  assert.equal(
    display({
      field: "auto_scheduling_status",
      currentValue: "At risk",
      oldValue: "at_risk",
    }).message,
    "updated auto-scheduling status to At risk",
  );
  assert.equal(
    display({
      field: "auto_scheduling_status",
      currentValue: "cannot_fit",
      oldValue: null,
    }).message,
    "changed auto-scheduling to Cannot fit",
  );
});

test("schedule toggles respect explicit false and do not infer state from malformed values", () => {
  assert.equal(
    display({
      field: "auto_scheduling_enabled",
      currentValue: "true",
      newValue: false,
    }).message,
    "paused auto-scheduling",
  );
  assert.equal(
    display({ field: "auto_scheduling_enabled", currentValue: "true" }).message,
    "enabled auto-scheduling",
  );
  assert.equal(
    display({ field: "auto_scheduling_locked", currentValue: "false" }).message,
    "unlocked the auto-scheduled calendar blocks",
  );
  assert.equal(
    display({ field: "auto_scheduling_locked", currentValue: "true" }).message,
    "locked the auto-scheduled calendar blocks",
  );
  assert.equal(
    display({ field: "auto_scheduling_locked", currentValue: "broken" })
      .message,
    "updated auto-scheduling",
  );
});

test("status, assignee and scope references resolve old and current values", () => {
  assert.equal(
    display({
      field: "status_id",
      currentValue: NEW_STATUS_ID,
      oldValue: STATUS_ID,
    }).message,
    "moved the task from Todo to In progress",
  );
  assert.equal(
    display({ field: "assignee_id", currentValue: MEMBER_ID }).message,
    "assigned the task to Jane Doe",
  );
  assert.equal(
    display({ field: "assignee_id", currentValue: MAYA_ID }).message,
    "assigned the task to Maya",
  );
  assert.equal(
    display({ field: "sprint_id", currentValue: REFERENCE_ID }).message,
    "moved the task to September",
  );
  assert.equal(
    display({ field: "objective_id", currentValue: REFERENCE_ID }).message,
    "moved the task to Launch",
  );
});

test("human assignment activities resolve the separate Maya account without replacing the actor", () => {
  const references = {
    ...context,
    members: indexActivityPeople([person], maya),
  };
  const humanActivity = {
    userId: MEMBER_ID,
    user: person,
    field: "assignee_id",
    currentValue: MAYA_ID,
  };
  const assigned = display(humanActivity, references);
  assert.deepEqual(assigned.actor, { name: "Jane Doe", isSystem: false });
  assert.equal(assigned.message, "assigned the task to Maya");
  assert.equal(
    display({ ...humanActivity, oldValue: MEMBER_ID }, references).message,
    "reassigned the task from Jane Doe to Maya",
  );
  assert.equal(
    display(
      { ...humanActivity, oldValue: MAYA_ID, currentValue: MEMBER_ID },
      references,
    ).message,
    "reassigned the task from Maya to Jane Doe",
  );
  assert.equal(context.members.has(MAYA_ID), false);
});

test("Maya lookup merges by ID and missing or non-system responses never fabricate an account", () => {
  const members = [person, { ...maya, fullName: "Outdated name" }];
  const indexed = indexActivityPeople(members, maya);
  assert.equal(indexed.size, 2);
  assert.equal(indexed.get(MAYA_ID)?.fullName, "Maya");
  assert.equal(members[1].fullName, "Outdated name");
  assert.equal(indexActivityPeople([person], null).has(MAYA_ID), false);
  assert.equal(
    indexActivityPeople([person], { ...maya, isSystem: false }).has(MAYA_ID),
    false,
  );
});

test("deleted/unavailable references do not expose IDs or claim the account was deleted", () => {
  for (const field of [
    "status_id",
    "assignee_id",
    "sprint_id",
    "objective_id",
    "key_result_id",
  ]) {
    const result = display({ field, currentValue: OTHER_ID });
    assert.match(result.message, /unavailable/);
    assert.ok(!result.message.includes(OTHER_ID));
    assert.ok(!result.message.includes("deleted"));
  }
});

test("removals use clear verbs; nil substrings in legitimate names are preserved", () => {
  assert.equal(
    display({
      field: "assignee_id",
      currentValue: "<nil>",
      oldValue: MEMBER_ID,
    }).message,
    "unassigned the task",
  );
  assert.equal(
    display({
      field: "sprint_id",
      currentValue: "<nil>",
      oldValue: REFERENCE_ID,
    }).message,
    "removed the task from September",
  );
  assert.equal(
    display({ field: "end_date", currentValue: "" }).message,
    "removed the deadline",
  );
  assert.equal(
    display({ field: "title", currentValue: "Vanilla launch" }).message,
    "renamed the task to Vanilla launch",
  );
  assert.equal(
    display({ field: "title", currentValue: "nil" }).message,
    "renamed the task to nil",
  );
});

test("date-only changes stay on the selected day in western and eastern device timezones", () => {
  const previousTimezone = process.env.TZ;
  try {
    for (const timezone of ["America/Los_Angeles", "Pacific/Auckland"]) {
      process.env.TZ = timezone;
      assert.equal(
        display({ field: "start_date", currentValue: "2026-09-15T00:00:00Z" })
          .message,
        "set the start date to 15 Sep 2026",
      );
    }
  } finally {
    if (previousTimezone === undefined) delete process.env.TZ;
    else process.env.TZ = previousTimezone;
  }
  assert.equal(
    display({ field: "end_date", currentValue: "2026-02-30" }).message,
    "set the deadline to an unavailable date",
  );
  assert.equal(formatActivityTimestamp("invalid"), null);
  assert.equal(
    formatActivityTimestamp(
      "2026-09-15T12:00:00Z",
      new Date("2026-09-15T12:05:00Z"),
    ),
    "5 minutes ago",
  );
});

test("time estimates include units and complexity follows canonical values", () => {
  assert.equal(
    display({ field: "estimated_duration_minutes", currentValue: "90" })
      .message,
    "set the task time needed to 1 hour 30 minutes",
  );
  assert.equal(
    display({ field: "minimum_focus_block_minutes", currentValue: "30" })
      .message,
    "set the task minimum focus block to 30 minutes",
  );
  assert.equal(
    display({ field: "estimate_unit", currentValue: "5" }).message,
    "set the task complexity to L",
  );
  assert.equal(
    display(
      { field: "estimate_unit", currentValue: "5" },
      { ...context, estimateScheme: "points" },
    ).message,
    "set the task complexity to 5 points",
  );
});

test("collections handle API arrays, legacy arrays, missing members and clearing", () => {
  assert.equal(
    display({ field: "labels", currentValue: "", newValue: [REFERENCE_ID] })
      .message,
    "updated labels to Mobile",
  );
  assert.equal(
    display({ field: "labels", currentValue: "old", newValue: [] }).message,
    "updated labels to no labels",
  );
  assert.equal(
    display({
      field: "collaborator_ids",
      currentValue: `[${MEMBER_ID} ${OTHER_ID}]`,
    }).message,
    "updated collaborators to Jane Doe, an unavailable member",
  );
  assert.equal(
    display({ field: "labels", currentValue: "", newValue: [OTHER_ID] })
      .message,
    "updated labels to an unavailable label",
  );
});

test("description activities never display HTML and relationship codes stay internal", () => {
  assert.equal(
    display({
      field: "description_html",
      currentValue: "<p>Private description</p>",
    }).message,
    "updated the description",
  );
  const result = display({
    field: "related_id",
    currentValue: "Launch checklist",
    reason: "association_removed",
  });
  assert.equal(
    result.message,
    "removed the related to relationship with Launch checklist",
  );
  assert.equal("reason" in result, false);
});

test("mobile updates omit rationale while preserving the actor and scheduled event", () => {
  const result = getActivityDisplay(
    {
      ...activity,
      reason: "Moved because the assignee has a conflicting meeting.",
    },
    context,
  );
  assert.deepEqual(result, {
    actor: { name: "Maya", isSystem: true },
    message: "scheduled work for 21 Aug 2026 at 10:23",
  });
});

test("hidden rationale does not replace a human update or its values", () => {
  const result = display({
    userId: MEMBER_ID,
    user: person,
    field: "priority",
    oldValue: "Low",
    currentValue: "High",
    reason: "Raised after planning.",
  });
  assert.deepEqual(result, {
    actor: { name: "Jane Doe", isSystem: false },
    message: "changed priority from Low to High",
  });
});

test("future field names are human readable and workspace terminology is honored", () => {
  assert.equal(
    display({ field: "review_policy", currentValue: "Required" }).message,
    "changed review policy to Required",
  );
  assert.equal(
    display({ type: "create" }, { ...context, storyTerm: "ticket" }).message,
    "created the ticket",
  );
  assert.equal(
    display(
      { field: "sprint_id", currentValue: OTHER_ID },
      { ...context, sprintTerm: "cycle" },
    ).message,
    "moved the task to an unavailable cycle",
  );
});
