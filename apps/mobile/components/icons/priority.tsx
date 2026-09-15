import type { StoryPriority } from "@/modules/stories/types";
import { useColorScheme } from "react-native";
import { themeColors } from "@/constants/colors";
import { priorityIconGeometry } from "./task-icon-geometry";
import { TaskIcon } from "./task-icon";

type PriorityIconProps = {
  priority: StoryPriority;
  size?: number;
};

export function PriorityIcon({
  priority = "No Priority",
  size = 16,
}: PriorityIconProps) {
  const dark = useColorScheme() === "dark";
  const color = themeColors[dark ? "dark" : "light"].textMuted;
  return (
    <TaskIcon size={size} geometry={priorityIconGeometry(priority, color)} />
  );
}
