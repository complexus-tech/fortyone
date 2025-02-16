const sprints = ["Created Sprint", "Updated Sprint", "Deleted Sprint"] as const;
const stories = ["Created Story", "Updated Story", "Deleted Story"] as const;
const objectives = [
  "Created Objective",
  "Updated Objective",
  "Deleted Objective",
] as const;
const search = ["Search Performed"] as const;

const trackingEvents = [
  ...sprints,
  ...stories,
  ...objectives,
  ...search,
] as const;

export type TrackingEvent = (typeof trackingEvents)[number];
