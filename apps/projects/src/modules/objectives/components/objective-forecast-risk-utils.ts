import { format, parseISO } from "date-fns";
import type { Objective } from "../types";

export const isObjectiveForecastAtRisk = (
  objective: Pick<Objective, "forecastDaysDelta" | "scheduleStatus">,
) => objective.scheduleStatus === "at_risk" && objective.forecastDaysDelta > 0;

export const getObjectiveForecastRiskCopy = (
  objective: Pick<
    Objective,
    | "endDate"
    | "forecastCauseStory"
    | "forecastDaysDelta"
    | "forecastEndDate"
    | "scheduleStatus"
  >,
  storyTerm = "work item",
) => {
  if (!isObjectiveForecastAtRisk(objective)) return null;

  const dayLabel = objective.forecastDaysDelta === 1 ? "day" : "days";
  const forecastDate = objective.forecastEndDate
    ? format(parseISO(objective.forecastEndDate), "MMM d, yyyy")
    : null;
  const targetDate = objective.endDate
    ? format(parseISO(objective.endDate), "MMM d, yyyy")
    : null;
  const timing =
    forecastDate && targetDate
      ? `Linked work is forecast for ${forecastDate}; the target is ${targetDate}.`
      : `Linked work is currently forecast ${objective.forecastDaysDelta} ${dayLabel} beyond the target.`;
  const cause = objective.forecastCauseStory
    ? ` ${storyTerm} ${objective.forecastCauseStory.sequenceId}, ${objective.forecastCauseStory.title}, is currently driving the forecast.`
    : "";

  return {
    description: `${timing}${cause}`,
    headline: `Forecast is ${objective.forecastDaysDelta} ${dayLabel} past target`,
    shortLabel: `Forecast +${objective.forecastDaysDelta}d`,
  };
};
