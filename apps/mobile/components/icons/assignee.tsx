import { useColorScheme } from "react-native";
import { themeColors } from "@/constants/colors";
import { assigneeIconGeometry } from "./task-icon-geometry";
import { TaskIcon } from "./task-icon";

export function AssigneeIcon({
  size = 18,
  color,
}: {
  size?: number;
  color?: string;
}) {
  const theme = useColorScheme() === "dark" ? "dark" : "light";
  return (
    <TaskIcon
      size={size}
      geometry={assigneeIconGeometry(color ?? themeColors[theme].icon)}
    />
  );
}
