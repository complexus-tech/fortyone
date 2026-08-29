import type { DetailedStory } from "@/modules/story/types";
import { formatTimeNeeded } from "@/lib/time-needed";

type StoryCalendarState = Pick<
  DetailedStory,
  | "autoSchedulingEnabled"
  | "autoSchedulingStatus"
  | "endDate"
  | "estimatedDurationMinutes"
>;

const toDateOnly = (value: string | null) => value?.slice(0, 10) ?? null;
const formatStoryCount = (count: number) =>
  `${count} ${count === 1 ? "story" : "stories"}`;
const formatScopedStoryCount = (count: number, total: number) =>
  count === total ? `all ${formatStoryCount(count)}` : formatStoryCount(count);

export const getStoryCalendarImpact = (story: StoryCalendarState) => {
  if (!story.autoSchedulingEnabled) {
    return "Calendar scheduling is off, so Maya will not reserve focus time for it.";
  }

  if (!story.estimatedDurationMinutes) {
    return "Calendar scheduling is on, but Maya still needs the time needed before it can reserve focus time.";
  }

  const duration = formatTimeNeeded(story.estimatedDurationMinutes, "full");
  const deliveryDate = toDateOnly(story.endDate);
  if (!deliveryDate) {
    return `Calendar scheduling is on for ${duration}, but no delivery date is set.`;
  }

  if (story.autoSchedulingStatus === "needs_owner") {
    return `Calendar scheduling is on for ${duration} before ${deliveryDate}, but Maya still needs a calendar owner.`;
  }

  return `Maya will reserve ${duration} of focus time on the assignee's calendar before ${deliveryDate}.`;
};

export const getBulkStoryCalendarImpact = (stories: StoryCalendarState[]) => {
  if (stories.length === 0) {
    return "No calendar time was reserved because no stories were created.";
  }

  const scheduledStories = stories.filter(
    (story) => story.autoSchedulingEnabled,
  );
  const unscheduledCount = stories.length - scheduledStories.length;

  if (scheduledStories.length === 0) {
    return `Calendar scheduling is off for all ${stories.length} stories. Add each story's time needed and delivery details manually, or provide those details for Maya to schedule selected stories.`;
  }

  const enabledCount = scheduledStories.length;
  const countStatus = (status: StoryCalendarState["autoSchedulingStatus"]) =>
    scheduledStories.filter((story) => story.autoSchedulingStatus === status)
      .length;
  const impact: string[] = [];

  if (unscheduledCount === 0) {
    impact.push(`Calendar scheduling is on for all ${stories.length} stories.`);
  } else {
    impact.push(
      `Calendar scheduling is on for ${formatStoryCount(enabledCount)} and off for ${formatStoryCount(unscheduledCount)}.`,
    );
  }

  const reservedCount = countStatus("scheduled") + countStatus("locked");
  if (reservedCount > 0) {
    impact.push(
      `Focus time is reserved for ${formatScopedStoryCount(reservedCount, stories.length)}.`,
    );
  }

  const planningCount = countStatus("planning");
  if (planningCount > 0) {
    impact.push(
      `Maya is planning focus time for ${formatScopedStoryCount(planningCount, stories.length)}; no reservation is confirmed yet.`,
    );
  }

  const needsOwnerCount = countStatus("needs_owner");
  if (needsOwnerCount > 0) {
    impact.push(
      `${formatStoryCount(needsOwnerCount)} still ${needsOwnerCount === 1 ? "needs" : "need"} a calendar owner before Maya can plan focus time; no reservation is confirmed for ${needsOwnerCount === 1 ? "it" : "them"}.`,
    );
  }

  const needsTimeCount = countStatus("needs_time");
  if (needsTimeCount > 0) {
    impact.push(
      `${formatStoryCount(needsTimeCount)} still ${needsTimeCount === 1 ? "needs" : "need"} time needed before Maya can plan focus time; no reservation is confirmed for ${needsTimeCount === 1 ? "it" : "them"}.`,
    );
  }

  const cannotFitCount = countStatus("cannot_fit");
  if (cannotFitCount > 0) {
    impact.push(
      `Maya could not fit ${formatScopedStoryCount(cannotFitCount, stories.length)} before ${cannotFitCount === 1 ? "its" : "their"} delivery ${cannotFitCount === 1 ? "date" : "dates"}.`,
    );
  }

  const atRiskCount = countStatus("at_risk");
  if (atRiskCount > 0) {
    impact.push(
      `${formatStoryCount(atRiskCount)} ${atRiskCount === 1 ? "has" : "have"} planned calendar work at risk of missing ${atRiskCount === 1 ? "its" : "their"} delivery ${atRiskCount === 1 ? "date" : "dates"}.`,
    );
  }

  const waitingCount = countStatus("off");
  if (waitingCount > 0) {
    impact.push(
      `${formatStoryCount(waitingCount)} ${waitingCount === 1 ? "is" : "are"} still waiting for scheduling to start; no reservation is confirmed yet.`,
    );
  }

  return impact.join(" ");
};
