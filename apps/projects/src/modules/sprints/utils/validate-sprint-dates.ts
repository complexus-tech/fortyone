import { isAfter, isBefore, isEqual } from "date-fns";
import { type Sprint } from "../types";

export const validateSprintDates = (
  startDate: string,
  endDate: string,
  existingSprints: Sprint[],
  currentSprintId?: string,
) => {
  const newStartDate = new Date(startDate);
  const newEndDate = new Date(endDate);

  // Validate that deadline is after start date
  if (isBefore(newEndDate, newStartDate) || isEqual(newEndDate, newStartDate)) {
    return {
      isValid: false,
      error: "Deadline must be after start date.",
    };
  }

  // Filter out the current sprint if we're updating
  const sprints = currentSprintId
    ? existingSprints.filter((sprint) => sprint.id !== currentSprintId)
    : existingSprints;

  // Check for overlaps with existing sprints
  const hasOverlap = sprints.some((sprint) => {
    const sprintStartDate = new Date(sprint.startDate);
    const sprintEndDate = new Date(sprint.endDate);

    // Check if the new sprint overlaps with an existing sprint
    const isStartDateOverlap =
      (isEqual(newStartDate, sprintStartDate) ||
        isAfter(newStartDate, sprintStartDate)) &&
      (isEqual(newStartDate, sprintEndDate) ||
        isBefore(newStartDate, sprintEndDate));

    const isEndDateOverlap =
      (isEqual(newEndDate, sprintStartDate) ||
        isAfter(newEndDate, sprintStartDate)) &&
      (isEqual(newEndDate, sprintEndDate) ||
        isBefore(newEndDate, sprintEndDate));

    const isSprintEnclosed =
      isBefore(newStartDate, sprintStartDate) &&
      isAfter(newEndDate, sprintEndDate);

    return isStartDateOverlap || isEndDateOverlap || isSprintEnclosed;
  });

  if (hasOverlap) {
    return {
      isValid: false,
      error:
        "Sprint dates overlap with an existing sprint. Please choose different dates.",
    };
  }

  return { isValid: true };
};
