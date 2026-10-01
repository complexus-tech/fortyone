import type { CustomField } from "./types";
import {
  customFieldReportSchema,
  customFieldSchema,
  storyCustomFieldsSchema,
} from "./types";
import {
  formatCustomFieldValue,
  formatExactDecimal,
  prepareCustomFieldValues,
  validateCustomFieldValue,
  createStoryValueBatches,
  getCustomFieldInputLabel,
} from "./value-utils";

const FIELD_ID = "d350e64b-06cd-43e5-8bc8-66d756d54378";
const OPTION_ID = "8241188a-71d2-47e2-98f0-465d4fce5558";
const base: CustomField = {
  id: FIELD_ID,
  teamId: "c5f7e92d-3810-45a8-86fc-940835532701",
  name: "Budget",
  type: "money",
  currency: "USD",
  options: [],
  showOnCreate: false,
  archivedAt: null,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
};

describe("exact custom field amounts", () => {
  it("includes the immutable currency in money input labels", () => {
    expect(getCustomFieldInputLabel(base)).toBe("Budget (USD)");
    expect(
      getCustomFieldInputLabel({ ...base, type: "number", currency: null }),
    ).toBe("Budget");
  });
  it("formats money beyond the safe integer range without precision loss", () => {
    expect(formatExactDecimal("9007199254740993.29", "USD")).toBe(
      "USD 9,007,199,254,740,993.29",
    );
    expect(formatExactDecimal("999999999999999999.995", "USD")).toBe(
      "USD 1,000,000,000,000,000,000.00",
    );
    expect(formatExactDecimal("-123.4567", "KWD")).toBe("KWD −123.457");
    expect(formatExactDecimal("125.9", "JPY")).toBe("JPY 126");
  });
  it("retains exact number precision and keeps zero distinct from missing", () => {
    expect(formatExactDecimal("123456.000000001")).toBe("123,456.000000001");
    expect(formatCustomFieldValue(base, "0")).toBe("USD 0.00");
    expect(formatCustomFieldValue(base, null)).toBe("Not set");
    expect(prepareCustomFieldValues([base], { [FIELD_ID]: "0" })).toEqual([
      { fieldId: FIELD_ID, value: "0" },
    ]);
    expect(prepareCustomFieldValues([base], { [FIELD_ID]: " " })).toEqual([
      { fieldId: FIELD_ID, value: null },
    ]);
  });
});

describe("bounded custom field reads", () => {
  it("deduplicates, sorts and excludes optimistic story IDs before batching", () => {
    const ids = Array.from(
      { length: 205 },
      (_, index) =>
        `d350e64b-06cd-43e5-8bc8-${index.toString().padStart(12, "0")}`,
    );
    const batches = createStoryValueBatches([
      ...ids.reverse(),
      ids[0],
      "123-optimistic",
    ]);
    expect(batches.map((batch) => batch.length)).toEqual([100, 100, 5]);
    expect(new Set(batches.flat()).size).toBe(205);
    expect(batches.flat()).toEqual([...batches.flat()].sort());
  });
});

describe("custom field value validation", () => {
  it("rejects overflow-prone numeric notation and impossible dates", () => {
    expect(validateCustomFieldValue(base, "1e309")).toBeTruthy();
    expect(validateCustomFieldValue(base, "1,250")).toBeTruthy();
    const date = { ...base, type: "date" as const, currency: null };
    expect(validateCustomFieldValue(date, "2026-02-30")).toBeTruthy();
    expect(validateCustomFieldValue(date, "2028-02-29")).toBeNull();
  });
  it("omits stale template definitions and options while preserving active values", () => {
    const select: CustomField = {
      ...base,
      type: "select",
      currency: null,
      options: [
        { id: OPTION_ID, name: "Enterprise", archivedAt: "2026-09-30" },
      ],
    };
    expect(
      prepareCustomFieldValues([select], { [FIELD_ID]: OPTION_ID }),
    ).toEqual([]);
    expect(formatCustomFieldValue(select, OPTION_ID)).toBe("Enterprise");
    expect(validateCustomFieldValue(select, OPTION_ID)).toBe(
      "Choose an active option.",
    );
    expect(
      prepareCustomFieldValues([{ ...base, archivedAt: "2026-09-30" }], {
        [FIELD_ID]: "17.25",
      }),
    ).toEqual([]);
    expect(
      prepareCustomFieldValues([base], { "other-team-field": "17.25" }),
    ).toEqual([]);
  });
});

describe("custom field runtime contracts", () => {
  it("rejects inconsistent currency types, duplicate options and numeric value coercion", () => {
    expect(
      customFieldSchema.safeParse({ ...base, currency: null }).success,
    ).toBe(false);
    expect(
      customFieldSchema.safeParse({
        ...base,
        options: [
          { id: OPTION_ID, name: "A" },
          { id: OPTION_ID, name: "B" },
        ],
      }).success,
    ).toBe(false);
    expect(
      storyCustomFieldsSchema.safeParse({
        fields: [base],
        values: [{ fieldId: FIELD_ID, value: 123.45 }],
      }).success,
    ).toBe(false);
  });
  it("retains report decimals exactly and rejects malformed coverage counts", () => {
    const report = {
      field: base,
      aggregation: "sum",
      groupBy: "none",
      currency: "USD",
      totalCount: 2,
      valuedCount: 1,
      missingCount: 1,
      rows: [
        {
          key: "all",
          label: "All work",
          value: "9007199254740993.29",
          count: 1,
        },
      ],
    };
    expect(customFieldReportSchema.parse(report).rows[0]?.value).toBe(
      "9007199254740993.29",
    );
    expect(
      customFieldReportSchema.safeParse({ ...report, missingCount: -1 })
        .success,
    ).toBe(false);
    expect(
      customFieldReportSchema.safeParse({
        ...report,
        rows: [{ ...report.rows[0], value: "NaN" }],
      }).success,
    ).toBe(false);
  });
});
