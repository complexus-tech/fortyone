import { formatDocumentRelativeTime } from "./relative-time";

describe("formatDocumentRelativeTime", () => {
  beforeAll(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date("2026-08-07T12:00:00.000Z"));
  });

  afterAll(() => {
    jest.useRealTimers();
  });

  it.each([
    ["2026-08-07T11:59:58.000Z", "Just now"],
    ["2026-08-07T11:59:00.000Z", "1m"],
    ["2026-08-07T11:10:00.000Z", "50m"],
    ["2026-08-07T11:00:00.000Z", "1h"],
    ["2026-08-07T10:00:00.000Z", "2h"],
    ["2026-08-06T12:00:00.000Z", "1d"],
    ["2026-08-05T12:00:00.000Z", "2d"],
  ])("formats %s as %s", (value, expected) => {
    expect(formatDocumentRelativeTime(value)).toBe(expected);
  });

  it("returns an empty label for an invalid timestamp", () => {
    expect(formatDocumentRelativeTime("not-a-date")).toBe("");
  });
});
