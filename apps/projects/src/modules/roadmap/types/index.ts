export const ROADMAP_LAYOUTS = ["kanban", "gantt", "list"] as const;

export type RoadmapLayoutType = (typeof ROADMAP_LAYOUTS)[number];

const ROADMAP_LAYOUT_LABELS: Record<RoadmapLayoutType, string> = {
  gantt: "Timeline",
  kanban: "Board",
  list: "List",
};

export const getRoadmapLayoutLabel = (layout: RoadmapLayoutType) =>
  ROADMAP_LAYOUT_LABELS[layout];
