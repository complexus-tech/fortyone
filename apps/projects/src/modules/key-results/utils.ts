import type { KeyResult } from "@/modules/objectives/types";
import type { KeyResultWithTeam } from "./types";

export type ObjectiveKeyResultGroup = {
  averageProgress: number;
  keyResults: KeyResultWithTeam[];
  objectiveId: string;
  objectiveName: string;
  teamId: string;
  teamName: string;
};

const numberFormatter = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 2,
});

export const getKeyResultProgress = (
  keyResult: Pick<
    KeyResult,
    "measurementType" | "startValue" | "currentValue" | "targetValue"
  >,
) => {
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

export const formatKeyResultValue = (
  value: number,
  measurementType: KeyResult["measurementType"],
) => {
  if (measurementType === "boolean") {
    return value === 1 ? "Complete" : "Incomplete";
  }

  const formattedValue = numberFormatter.format(value);
  return measurementType === "percentage"
    ? `${formattedValue}%`
    : formattedValue;
};

export const getKeyResultReference = (
  teamCode: string | null | undefined,
  sequenceId: number | null | undefined,
) => {
  if (
    !teamCode?.trim() ||
    typeof sequenceId !== "number" ||
    !Number.isInteger(sequenceId) ||
    sequenceId <= 0
  ) {
    return null;
  }

  return `${teamCode.trim().toUpperCase()}-${sequenceId}`;
};

export const groupKeyResultsByObjective = (
  keyResults: KeyResultWithTeam[],
): ObjectiveKeyResultGroup[] => {
  const groups = new Map<
    string,
    Omit<ObjectiveKeyResultGroup, "averageProgress"> & {
      progressTotal: number;
    }
  >();

  for (const keyResult of keyResults) {
    const existingGroup = groups.get(keyResult.objectiveId);
    const progress = getKeyResultProgress(keyResult);

    if (existingGroup) {
      existingGroup.keyResults.push(keyResult);
      existingGroup.progressTotal += progress;
      continue;
    }

    groups.set(keyResult.objectiveId, {
      keyResults: [keyResult],
      objectiveId: keyResult.objectiveId,
      objectiveName: keyResult.objectiveName,
      progressTotal: progress,
      teamId: keyResult.teamId,
      teamName: keyResult.teamName,
    });
  }

  return Array.from(groups.values(), (group) => ({
    averageProgress: Math.round(
      group.progressTotal / Math.max(group.keyResults.length, 1),
    ),
    keyResults: group.keyResults,
    objectiveId: group.objectiveId,
    objectiveName: group.objectiveName,
    teamId: group.teamId,
    teamName: group.teamName,
  }));
};
