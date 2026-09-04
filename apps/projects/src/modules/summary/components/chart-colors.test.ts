import {
  SUMMARY_CHART_PRIMARY,
  SUMMARY_STATUS_COLORS,
  getSummaryPriorityColor,
} from "./chart-colors";

describe("summary chart colors", () => {
  it.each([
    ["Urgent", "var(--color-danger)"],
    ["High", "var(--color-warning)"],
    ["Medium", "var(--color-success)"],
    ["Low", "var(--color-info)"],
    ["No Priority", "var(--color-text-muted)"],
  ])("maps %s to its semantic color", (priority, expected) => {
    expect(getSummaryPriorityColor(priority)).toBe(expected);
  });

  it("falls back to the product accent for custom priorities", () => {
    expect(getSummaryPriorityColor("Custom")).toBe(SUMMARY_CHART_PRIMARY);
  });

  it("uses distinct categorical colors for status segments", () => {
    expect(new Set(SUMMARY_STATUS_COLORS).size).toBe(
      SUMMARY_STATUS_COLORS.length,
    );
  });
});
