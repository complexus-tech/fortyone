/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Objective } from "@/modules/objectives/types";
import { getStrategyKeyResultQueryObjectives } from "./use-strategy-key-results";

const objective = (id: string, keyResultCount: number) =>
  ({ id, keyResultCount }) as Objective;

describe("strategy key-result query window", () => {
  it("creates queries only for active objectives that can have key results", () => {
    const activeObjective = objective("active", 2);
    const zeroKeyResultObjective = objective("zero", 0);
    const distantObjective = objective("distant", 3);

    expect(
      getStrategyKeyResultQueryObjectives(
        [activeObjective, zeroKeyResultObjective, distantObjective],
        new Set([activeObjective.id, zeroKeyResultObjective.id]),
      ),
    ).toEqual([activeObjective]);
  });
});
