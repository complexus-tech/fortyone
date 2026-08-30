/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { format } from "date-fns";
import { getDueDateTone, parseDueDate } from "./due-date";

describe("story due-date presentation", () => {
  it("preserves the calendar day for date-only and legacy timestamp values", () => {
    expect(format(parseDueDate("2026-08-30"), "yyyy-MM-dd")).toBe("2026-08-30");
    expect(format(parseDueDate("2026-08-30T00:00:00.000Z"), "yyyy-MM-dd")).toBe(
      "2026-08-30",
    );
  });

  it("uses a neutral tone until the browser has hydrated the current time", () => {
    expect(getDueDateTone(new Date(2026, 7, 30), null)).toBe("neutral");
  });

  it("classifies overdue and due-soon deadlines after hydration", () => {
    const now = new Date(2026, 7, 30, 12);

    expect(getDueDateTone(new Date(2026, 7, 29), now)).toBe("overdue");
    expect(getDueDateTone(new Date(2026, 8, 4), now)).toBe("due-soon");
    expect(getDueDateTone(new Date(2026, 8, 8), now)).toBe("neutral");
  });
});
