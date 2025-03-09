const sprints = ["sprint_created", "sprint_updated", "sprint_deleted"] as const;
const stories = ["story_created", "story_updated", "story_deleted"] as const;
const objectives = [
  "objective_created",
  "objective_updated",
  "objective_deleted",
  "objective_archived",
  "objective_restored",
] as const;
const okrs = ["okr_created", "okr_updated", "okr_deleted"] as const;
const search = [
  "search_performed",
  "search_abandoned",
  "search_filtered",
  "search_no_results",
  "search_result_clicked",
] as const;

const trackingEvents = [
  ...sprints,
  ...stories,
  ...objectives,
  ...okrs,
  ...search,
] as const;

export type TrackingEvent = (typeof trackingEvents)[number];
