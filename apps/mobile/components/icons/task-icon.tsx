import Svg, { Circle, Path, Rect } from "react-native-svg";
import type { TaskIconGeometry } from "./task-icon-geometry";

export function TaskIcon({
  geometry,
  size,
}: {
  geometry: TaskIconGeometry;
  size: number;
}) {
  return (
    <Svg
      width={size}
      height={size}
      viewBox={geometry.viewBox}
      fill="none"
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
    >
      {geometry.circles?.map((circle) => <Circle key={circle.r} {...circle} />)}
      {geometry.rects?.map((rect) => <Rect key={rect.x} {...rect} />)}
      {geometry.paths?.map((path) => <Path key={path.d} {...path} />)}
    </Svg>
  );
}
