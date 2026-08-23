import type { CSSProperties } from "react";
import type { StoryPriority } from "@/modules/stories/types";

const PRIORITY_COLORS: Partial<Record<StoryPriority, string>> = {
  Urgent: "var(--color-danger)",
  High: "var(--color-warning)",
  Medium: "var(--color-success)",
  Low: "var(--color-info)",
};
const CATEGORY_TINT_PERCENTAGE = 5;

export const getCategoryHeaderStyle = ({
  statusColor,
  priority,
}: {
  statusColor?: string;
  priority?: StoryPriority | null;
}): CSSProperties | undefined => {
  const categoryColor =
    statusColor ?? (priority ? PRIORITY_COLORS[priority] : undefined);

  if (!categoryColor) return undefined;

  return {
    backgroundImage: `linear-gradient(to right, transparent, color-mix(in oklab, ${categoryColor} ${CATEGORY_TINT_PERCENTAGE}%, transparent))`,
    borderImageSlice: 1,
    borderImageSource: `linear-gradient(to right, var(--color-border), color-mix(in oklab, ${categoryColor} ${CATEGORY_TINT_PERCENTAGE}%, var(--color-border)))`,
  };
};
