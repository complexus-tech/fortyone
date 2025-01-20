export type ObjectiveStatus =
  | "Not Started"
  | "In Progress"
  | "Completed"
  | "Cancelled";

export type MeasureType =
  | "Number"
  | "Percent (%)"
  | "Boolean (Complete/Incomplete)";

export type KeyResult = {
  id: string;
  name: string;
  measureType: MeasureType;
  startValue: number;
  targetValue: number;
};

export type NewObjective = {
  name: string;
  description: string;
  descriptionHTML: string;
  teamId: string;
  status: ObjectiveStatus;
  startDate: string | null;
  endDate: string | null;
  leadUserId: string | null;
  keyResults: KeyResult[];
};
