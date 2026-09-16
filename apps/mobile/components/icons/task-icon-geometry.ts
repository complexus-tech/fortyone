import type { StoryPriority } from "../../modules/stories/types";
import type { StatusCategory } from "../../types/statuses";
import { colors } from "../../constants/colors";

type IconStroke = {
  fill?: string;
  stroke?: string;
  strokeWidth?: number;
  strokeLinecap?: "round";
  strokeLinejoin?: "round";
  strokeDasharray?: string;
  opacity?: number;
};

export type TaskIconGeometry = {
  viewBox: string;
  circles?: (IconStroke & { cx: number; cy: number; r: number })[];
  paths?: (IconStroke & { d: string })[];
  rects?: (IconStroke & {
    x: number;
    y: number;
    width: number;
    height: number;
    rx: number;
  })[];
};

// Plain SVG geometry is shared by React Native and the editor's DOM surface.
export function statusIconGeometry(
  category: StatusCategory = "unstarted",
  color: string,
): TaskIconGeometry {
  const paths: NonNullable<TaskIconGeometry["paths"]> = [];
  if (category === "started")
    paths.push({ d: "M10 4.5a5.5 5.5 0 0 1 0 11Z", fill: color });
  if (category === "paused")
    paths.push({
      d: "M8 7v6m4-6v6",
      stroke: color,
      strokeWidth: 1.85,
      strokeLinecap: "round",
    });
  if (category === "completed")
    paths.push({
      d: "m6.3 10 2.4 2.4 5-5",
      stroke: colors.white,
      strokeWidth: 1.85,
      strokeLinecap: "round",
      strokeLinejoin: "round",
    });
  if (category === "cancelled")
    paths.push({
      d: "m7.5 7.5 5 5m0-5-5 5",
      stroke: color,
      strokeWidth: 1.85,
      strokeLinecap: "round",
    });

  return {
    viewBox: "0 0 20 20",
    circles: [
      {
        cx: 10,
        cy: 10,
        r: 7.5,
        stroke: color,
        strokeWidth: 1.85,
        strokeDasharray: category === "backlog" ? "2.6 2.6" : undefined,
        fill: category === "completed" ? color : "none",
      },
    ],
    paths,
  };
}

const PRIORITY_BARS = [
  { x: 1, y: 8, height: 6 },
  { x: 6, y: 5, height: 9 },
  { x: 11, y: 2, height: 12 },
];
const FILLED_PRIORITY_BARS = { Low: 1, Medium: 2, High: 3 } as const;

// Match apps/projects/src/components/ui/priority-icon.tsx. Both native and
// editor DOM icons use this renderer so priority colors stay consistent.
const PRIORITY_COLORS = {
  Urgent: colors.danger,
  High: colors.warning,
  Medium: colors.success,
  Low: colors.info,
} satisfies Record<Exclude<StoryPriority, "No Priority">, string>;

export function priorityIconGeometry(
  priority: StoryPriority,
  mutedColor: string,
): TaskIconGeometry {
  const color =
    priority === "No Priority" ? mutedColor : PRIORITY_COLORS[priority];
  if (priority === "Urgent")
    return {
      viewBox: "0 0 16 16",
      rects: [{ x: 1, y: 1, width: 14, height: 14, rx: 3, fill: color }],
      paths: [
        {
          d: "M8 4.5v4m0 3h.01",
          stroke: colors.white,
          strokeWidth: 1.7,
          strokeLinecap: "round",
        },
      ],
    };

  return {
    viewBox: "0 0 16 16",
    rects:
      priority === "No Priority"
        ? [1, 6, 11].map((x) => ({
            x,
            y: 7,
            width: 3,
            height: 2.5,
            rx: 0.5,
            fill: color,
            opacity: 0.9,
          }))
        : PRIORITY_BARS.map((bar, index) => ({
            ...bar,
            width: 3,
            rx: 0.7,
            fill: color,
            opacity: index < FILLED_PRIORITY_BARS[priority] ? 1 : 0.4,
          })),
  };
}

export function assigneeIconGeometry(color: string): TaskIconGeometry {
  return {
    viewBox: "0 0 20 20",
    circles: [
      {
        cx: 10,
        cy: 10,
        r: 8,
        stroke: color,
        strokeWidth: 1.3,
        strokeDasharray: "2 2",
      },
      { cx: 10, cy: 7.5, r: 2.5, fill: color },
    ],
    paths: [{ d: "M5 15a5 5 0 0 1 10 0Z", fill: color }],
  };
}
