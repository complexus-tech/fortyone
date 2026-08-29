/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Objective } from "./types";
import { toKeyResultCreateInput } from "./key-result-generation";
import { keyResultGenerationSchema } from "./schemas/key-result-generation";

const objective = {
  id: "objective-1",
  startDate: "2026-08-01T00:00:00.000Z",
  endDate: "2026-09-30T00:00:00.000Z",
} as Objective;

describe("toKeyResultCreateInput", () => {
  it("rejects timestamp-shaped AI dates before they reach the API", () => {
    const result = keyResultGenerationSchema.safeParse({
      keyResults: [
        {
          name: "Increase activation",
          measurementType: "number",
          startValue: 10,
          targetValue: 25,
          startDate: "2026-08-05T00:00:00.000Z",
          endDate: "2026-09-15T00:00:00.000Z",
        },
      ],
    });

    expect(result.success).toBe(false);
  });

  it("maps a generated suggestion to the key-result create contract", () => {
    expect(
      toKeyResultCreateInput(
        {
          name: "  Increase activation  ",
          measurementType: "number",
          startValue: 10,
          targetValue: 25,
          startDate: "2026-08-05",
          endDate: "2026-09-15",
        },
        objective,
      ),
    ).toEqual({
      contributors: [],
      currentValue: 10,
      endDate: "2026-09-15",
      lead: null,
      measurementType: "number",
      name: "Increase activation",
      objectiveId: "objective-1",
      startDate: "2026-08-05",
      startValue: 10,
      targetValue: 25,
    });
  });

  it("keeps dates inside the objective range", () => {
    const result = toKeyResultCreateInput(
      {
        name: "Reach the target",
        measurementType: "percentage",
        startValue: -10,
        targetValue: 120,
        startDate: "2026-07-01",
        endDate: "2026-12-31",
      },
      objective,
    );

    expect(result).toMatchObject({
      currentValue: 0,
      endDate: "2026-09-30",
      startDate: "2026-08-01",
      startValue: 0,
      targetValue: 100,
    });
  });

  it("normalizes boolean values to the persisted representation", () => {
    const result = toKeyResultCreateInput(
      {
        name: "Launch",
        measurementType: "boolean",
        startValue: 0,
        targetValue: 12,
        startDate: "2026-08-01",
        endDate: "2026-09-30",
      },
      objective,
    );

    expect(result).toMatchObject({
      currentValue: 0,
      startValue: 0,
      targetValue: 1,
    });
  });
});
