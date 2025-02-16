const sprints = ["Created Sprint", "Updated Sprint", "Deleted Sprint"] as const;

const trackingEvents = [...sprints] as const;

export type TrackingEvent = (typeof trackingEvents)[number];
