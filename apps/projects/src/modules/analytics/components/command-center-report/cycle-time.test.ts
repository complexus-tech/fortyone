import { summarizeCycleTime } from "./model";

describe("observed cycle time", () => {
  it("shows unavailable when completions have no recorded start history", () => {
    expect(
      summarizeCycleTime([
        {
          date: "2026-10-01",
          activeUsers: 0,
          storiesPerDay: 4,
          avgCycleTime: 0,
          cycleTimeSamples: 0,
        },
      ]),
    ).toEqual({ samples: 0, averageDays: null });
  });
  it("weights daily averages by observed completions", () => {
    const common = { activeUsers: 0, storiesPerDay: 0 };
    expect(
      summarizeCycleTime([
        {
          ...common,
          date: "2026-09-30",
          avgCycleTime: 2,
          cycleTimeSamples: 10,
        },
        { ...common, date: "2026-10-01", avgCycleTime: 5, cycleTimeSamples: 2 },
        { ...common, date: "2026-10-02", avgCycleTime: 0, cycleTimeSamples: 0 },
      ]),
    ).toEqual({ samples: 12, averageDays: 2.5 });
  });
  it("preserves an observed zero and excludes legacy points without sample counts", () => {
    const common = { date: "2026-10-01", activeUsers: 0, storiesPerDay: 0 };
    expect(
      summarizeCycleTime([{ ...common, avgCycleTime: 0, cycleTimeSamples: 1 }]),
    ).toEqual({ samples: 1, averageDays: 0 });
    expect(summarizeCycleTime([{ ...common, avgCycleTime: 20 }])).toEqual({
      samples: 0,
      averageDays: null,
    });
  });
});
