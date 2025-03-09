const sprints = ["created_sprint", "updated_sprint", "deleted_sprint"] as const;
const stories = ["created_story", "updated_story", "deleted_story"] as const;
const objectives = [
  "created_objective",
  "updated_objective",
  "deleted_objective",
] as const;
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
  ...search,
] as const;

export type TrackingEvent = (typeof trackingEvents)[number];
