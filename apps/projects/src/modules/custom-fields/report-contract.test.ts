import type { CustomField, CustomFieldReport } from "./types";
import { customFieldReportSchema } from "./types";

const amount: CustomField = {
  id: "d350e64b-06cd-43e5-8bc8-66d756d54378",
  teamId: "c5f7e92d-3810-45a8-86fc-940835532701",
  name: "Deal amount",
  type: "money",
  currency: "USD",
  options: [],
  showOnCreate: true,
  archivedAt: null,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
};

const average: CustomFieldReport = {
  field: amount,
  aggregation: "average",
  groupBy: "status",
  currency: "USD",
  totalCount: 4,
  valuedCount: 2,
  missingCount: 2,
  rows: [
    {
      key: "backlog",
      label: "Backlog",
      value: "750.0000000000000000",
      count: 2,
    },
    { key: "started", label: "Started", value: "", count: 0 },
  ],
};

describe("custom field aggregate response contracts", () => {
  it("decodes two valued deals alongside an empty status group without coercing decimals", () => {
    const result = customFieldReportSchema.parse(average);
    expect(result.rows).toEqual(average.rows);
    expect(result.rows[0]?.value).toBe("750.0000000000000000");
    expect(result.rows[1]?.value).toBe("");
  });

  it.each(["average", "min", "max"] as const)(
    "preserves an entirely empty %s report and distinguishes a valued zero",
    (aggregation) => {
      const empty = {
        ...average,
        aggregation,
        totalCount: 2,
        valuedCount: 0,
        missingCount: 2,
        rows: [{ key: "all", label: "All work", value: "", count: 0 }],
      };
      expect(customFieldReportSchema.parse(empty).rows[0]?.value).toBe("");
      const zero = {
        ...empty,
        valuedCount: 1,
        missingCount: 1,
        rows: [{ ...empty.rows[0], value: "0.0000000000000000", count: 1 }],
      };
      expect(customFieldReportSchema.parse(zero).rows[0]?.value).toBe(
        "0.0000000000000000",
      );
    },
  );

  it.each(["sum", "count"] as const)(
    "requires a numeric zero for an empty %s group",
    (aggregation) => {
      const empty = {
        ...average,
        aggregation,
        currency: aggregation === "count" ? null : "USD",
        totalCount: 2,
        valuedCount: 0,
        missingCount: 2,
        rows: [{ key: "all", label: "All work", value: "0", count: 0 }],
      };
      expect(customFieldReportSchema.parse(empty).rows[0]?.value).toBe("0");
      expect(
        customFieldReportSchema.safeParse({
          ...empty,
          rows: [{ ...empty.rows[0], value: "" }],
        }).success,
      ).toBe(false);
    },
  );

  it("rejects an absent result for valued work and a fabricated zero for an empty average", () => {
    expect(
      customFieldReportSchema.safeParse({
        ...average,
        rows: [{ ...average.rows[0], value: "" }, average.rows[1]],
      }).success,
    ).toBe(false);
    expect(
      customFieldReportSchema.safeParse({
        ...average,
        rows: [average.rows[0], { ...average.rows[1], value: "0" }],
      }).success,
    ).toBe(false);
  });

  it("rejects inconsistent report coverage and group counts", () => {
    expect(
      customFieldReportSchema.safeParse({ ...average, missingCount: 1 })
        .success,
    ).toBe(false);
    expect(
      customFieldReportSchema.safeParse({
        ...average,
        rows: [{ ...average.rows[0], count: 1 }, average.rows[1]],
      }).success,
    ).toBe(false);
  });

  it.each([null, 750, "NaN", "1e10"])(
    "rejects a malformed aggregate value %p",
    (value) => {
      expect(
        customFieldReportSchema.safeParse({
          ...average,
          rows: [{ ...average.rows[0], value }, average.rows[1]],
        }).success,
      ).toBe(false);
    },
  );
});
