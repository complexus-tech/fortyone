export type PropertyMenuSelectionMode = "single" | "bulk";

export const shouldApplyPropertySelection = <T>(
  mode: PropertyMenuSelectionMode,
  currentValue: T,
  nextValue: T,
) => mode === "bulk" || !Object.is(currentValue, nextValue);

export const isPropertySelectionActive = <T>(
  mode: PropertyMenuSelectionMode,
  currentValue: T,
  candidateValue: T,
) => mode === "single" && Object.is(currentValue, candidateValue);

export const getNumberedMenuItem = <T>(items: T[], value: string) => {
  if (!/^\d+$/.test(value)) return undefined;

  return items.at(Number.parseInt(value, 10));
};
