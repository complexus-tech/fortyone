export type KeyResultProgressInput = {
  currentValue: number;
  measurementType: "boolean" | "number" | "percentage";
  startValue: number;
  targetValue: number;
};

/**
 * Computes a bounded completion percentage for any key-result read model.
 *
 * This is feature-neutral domain math, so presentation layers can share it
 * without importing the Key Results feature implementation.
 */
export const getKeyResultProgress = (keyResult: KeyResultProgressInput) => {
  if (keyResult.measurementType === "boolean") {
    return keyResult.currentValue === 1 ? 100 : 0;
  }

  const targetChange = keyResult.targetValue - keyResult.startValue;
  if (targetChange === 0) {
    return keyResult.currentValue === keyResult.targetValue ? 100 : 0;
  }

  const currentChange = keyResult.currentValue - keyResult.startValue;
  const progress = Math.round((currentChange / targetChange) * 100);

  return Math.min(Math.max(progress, 0), 100);
};
