const sprints = ["sprint_created", "sprint_updated", "sprint_deleted"] as const;
const stories = [
  "story_created",
  "story_updated",
  "story_deleted",
  "story_duplicated",
] as const;
const objectives = [
  "objective_created",
  "objective_updated",
  "objective_deleted",
  "objective_archived",
  "objective_restored",
] as const;

const teams = ["team_created", "team_updated", "team_deleted"] as const;
const keyResults = ["key_result_created", "key_result_updated"] as const;
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
  ...teams,
  ...keyResults,
  ...search,
] as const;

export type TrackingEvent = (typeof trackingEvents)[number];
