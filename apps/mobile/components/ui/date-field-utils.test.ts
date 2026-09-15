import assert from "node:assert/strict";
import test from "node:test";
import { formatISO, isValid } from "date-fns";
import { calendarDate, dateValue } from "./date-field-utils";

test("date commits preserve the selected calendar day across UTC offsets", () => {
  const original = process.env.TZ;
  try {
    for (const timezone of [
      "America/Los_Angeles",
      "Africa/Harare",
      "Pacific/Auckland",
    ]) {
      process.env.TZ = timezone;
      const selected = new Date(2026, 2, 17, 16, 30);
      assert.equal(
        formatISO(calendarDate(selected), { representation: "date" }),
        "2026-03-17",
      );
      assert.equal(
        formatISO(dateValue("2026-03-17T00:00:00Z"), {
          representation: "date",
        }),
        "2026-03-17",
      );
    }
  } finally {
    if (original === undefined) delete process.env.TZ;
    else process.env.TZ = original;
  }
});

test("missing or malformed stored dates cannot pass Invalid Date to the native picker", () => {
  for (const value of [null, undefined, "", "not-a-date", "2026-99-99"]) {
    assert.equal(isValid(dateValue(value)), true);
  }
});
