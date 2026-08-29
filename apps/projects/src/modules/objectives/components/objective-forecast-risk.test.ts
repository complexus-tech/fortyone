/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Objective } from "../types";
import {
  getObjectiveForecastRiskCopy,
  isObjectiveForecastAtRisk,
} from "./objective-forecast-risk-utils";

const createObjective = (overrides: Partial<Objective> = {}): Objective =>
  ({
    endDate: "2026-08-21",
    forecastCauseStory: {
      id: "story-42",
      sequenceId: 42,
      source: "planning",
      title: "Prepare the launch brief",
    },
    forecastDaysDelta: 8,
    forecastEndDate: "2026-08-29",
    scheduleStatus: "at_risk",
    ...overrides,
  }) as Objective;

describe("objective forecast risk", () => {
  it("explains the forecast delta and the work driving it", () => {
    const objective = createObjective();

    expect(isObjectiveForecastAtRisk(objective)).toBe(true);
    expect(getObjectiveForecastRiskCopy(objective, "Story")).toEqual({
      description:
        "Linked work is forecast for Aug 29, 2026; the target is Aug 21, 2026. Story 42, Prepare the launch brief, is currently driving the forecast.",
      headline: "Forecast is 8 days past target",
      shortLabel: "Forecast +8d",
    });
  });

  it("does not present manual health or on-track delivery as forecast risk", () => {
    expect(
      getObjectiveForecastRiskCopy(
        createObjective({
          forecastDaysDelta: 0,
          health: "Off Track",
          scheduleStatus: "on_track",
        }),
      ),
    ).toBeNull();
  });
});
