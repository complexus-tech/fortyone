import type { Status } from "@/types/statuses";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { statusIconGeometry } from "./task-icon-geometry";
import { TaskIcon } from "./task-icon";

export type StatusIconProps = {
  category?: Status["category"];
  color?: string;
  size?: number;
};

export function StatusIcon({
  category = "unstarted",
  color: customColor,
  size = 18,
}: StatusIconProps) {
  const { resolvedTheme } = useTheme();
  const color = customColor ?? themeColors[resolvedTheme].icon;
  return (
    <TaskIcon size={size} geometry={statusIconGeometry(category, color)} />
  );
}
