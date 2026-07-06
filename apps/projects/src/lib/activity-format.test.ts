import { format } from "date-fns";
import { formatActivityReasonDates } from "./activity-format";

describe("activity formatting", () => {
  it("formats ISO timestamps embedded in activity reasons", () => {
    const startAt = "2026-06-15T14:00:00Z";
    const endAt = "2026-06-15T16:30:00Z";

    expect(
      formatActivityReasonDates(
        `Maya found an available calendar slot from ${startAt} to ${endAt}.`,
      ),
    ).toBe(
      `Maya found an available calendar slot from ${format(new Date(startAt), "PP 'at' p")} to ${format(new Date(endAt), "PP 'at' p")}.`,
    );
  });

  it("leaves plain text reasons unchanged", () => {
    expect(
      formatActivityReasonDates("Assigned because they have capacity."),
    ).toBe("Assigned because they have capacity.");
  });
});
