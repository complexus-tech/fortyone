import { colors } from "lib";

const FALLBACK_OBJECTIVE_COLOR = "#FFE066";

export const getAvailableObjectiveColor = (usedColors: string[]) => {
  const normalizedUsedColors = new Set(
    usedColors.map((color) => color.toLowerCase()),
  );

  return (
    colors.find((color) => !normalizedUsedColors.has(color.toLowerCase())) ??
    FALLBACK_OBJECTIVE_COLOR
  );
};
