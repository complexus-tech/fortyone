import assert from "node:assert/strict";
import test from "node:test";
import { formatContextDates, getSprintTiming } from "./context-dates";

test("missing or malformed dates render useful context without throwing", () => {
  assert.equal(formatContextDates(null, undefined), "No dates set");
  assert.equal(
    formatContextDates("broken", "2026-09-15T00:00:00Z"),
    "Due Sep 15",
  );
  assert.equal(formatContextDates("2026-02-31", null), "No dates set");
  assert.equal(
    formatContextDates("2026-09-14T00:00:00Z", "2026-09-15"),
    "Sep 14 – Sep 15",
  );
});

test("sprints stay in progress through their final calendar day", () => {
  assert.equal(
    getSprintTiming("2026-09-01", "2026-09-15", new Date(2026, 8, 15, 23, 59)),
    "In progress",
  );
  assert.equal(
    getSprintTiming("2026-09-01", "2026-09-15", new Date(2026, 8, 16)),
    "Completed",
  );
  assert.equal(
    getSprintTiming("2026-09-17", "2026-09-20", new Date(2026, 8, 16)),
    "Upcoming",
  );
  assert.equal(getSprintTiming("2026-09-20", "2026-09-17"), null);
});
