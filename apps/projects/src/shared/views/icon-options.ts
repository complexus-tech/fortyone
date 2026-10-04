import type { ComponentType } from "react";
import {
  AnalyticsIcon,
  BookIcon,
  CalendarIcon,
  ClockIcon,
  CodeIcon,
  DocsIcon,
  GoalIcon,
  KanbanIcon,
  ListIcon,
  RoadmapIcon,
  StarIcon,
  TagsIcon,
  TeamIcon,
  WorkflowIcon,
} from "icons";
import type { ViewIconKey } from "./metadata";

export const VIEW_ICON_OPTIONS: readonly {
  value: ViewIconKey;
  label: string;
  Icon: ComponentType<{ className?: string }>;
}[] = [
  { value: "list", label: "List", Icon: ListIcon },
  { value: "kanban", label: "Board", Icon: KanbanIcon },
  { value: "calendar", label: "Calendar", Icon: CalendarIcon },
  { value: "clock", label: "Clock", Icon: ClockIcon },
  { value: "star", label: "Star", Icon: StarIcon },
  { value: "analytics", label: "Analytics", Icon: AnalyticsIcon },
  { value: "goal", label: "Goal", Icon: GoalIcon },
  { value: "tags", label: "Tags", Icon: TagsIcon },
  { value: "code", label: "Code", Icon: CodeIcon },
  { value: "team", label: "Team", Icon: TeamIcon },
  { value: "book", label: "Book", Icon: BookIcon },
  { value: "workflow", label: "Workflow", Icon: WorkflowIcon },
  { value: "roadmap", label: "Roadmap", Icon: RoadmapIcon },
  { value: "docs", label: "Documents", Icon: DocsIcon },
];
