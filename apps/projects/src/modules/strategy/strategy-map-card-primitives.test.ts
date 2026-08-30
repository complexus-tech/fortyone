/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { KeyResult, Objective } from "@/modules/objectives/types";
import {
  getCompleteStrategyMapAverageProgress,
  getObjectiveProgress,
} from "./strategy-map-progress";
import { getStrategyDescriptionPreview } from "./strategy-map-card-primitives";

const createObjective = (overrides: Partial<Objective> = {}) =>
  ({
    keyResultCount: 0,
    stats: {
      backlog: 0,
      cancelled: 0,
      completed: 3,
      started: 0,
      total: 4,
      unstarted: 1,
    },
    ...overrides,
  }) as Objective;

const createKeyResult = (overrides: Partial<KeyResult> = {}) =>
  ({
    currentValue: 25,
    measurementType: "percentage",
    startValue: 0,
    targetValue: 100,
    ...overrides,
  }) as KeyResult;

describe("strategy map card primitives", () => {
  it("creates a concise plain-text description preview", () => {
    expect(
      getStrategyDescriptionPreview(
        "<p>Ship&nbsp;the <strong>strategy</strong> &amp; learn &gt; repeat.</p>",
      ),
    ).toBe("Ship the strategy & learn > repeat.");
    expect(getStrategyDescriptionPreview(null)).toBe("");
  });

  it("uses key-result progress when key results are configured", () => {
    expect(
      getObjectiveProgress(createObjective({ keyResultCount: 2 }), [
        createKeyResult({ currentValue: 20 }),
        createKeyResult({ currentValue: 60 }),
      ]),
    ).toBe(40);
    expect(getObjectiveProgress(createObjective({ keyResultCount: 1 }))).toBe(
      0,
    );
  });

  it("falls back to story completion for objectives without key results", () => {
    expect(getObjectiveProgress(createObjective())).toBe(75);
  });

  it("does not report a misleading strategy average from a partial key-result window", () => {
    const keyResultObjective = createObjective({
      id: "objective-with-key-results",
      keyResultCount: 1,
    });
    const storyObjective = createObjective({ id: "story-objective" });

    expect(
      getCompleteStrategyMapAverageProgress(
        [keyResultObjective, storyObjective],
        new Map(),
        new Set(),
      ),
    ).toBeNull();
    expect(
      getCompleteStrategyMapAverageProgress(
        [keyResultObjective, storyObjective],
        new Map([
          [keyResultObjective.id, [createKeyResult({ currentValue: 25 })]],
        ]),
        new Set([keyResultObjective.id]),
      ),
    ).toBe(50);
  });
});
