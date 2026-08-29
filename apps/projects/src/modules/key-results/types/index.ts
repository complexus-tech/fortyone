import type { KeyResult } from "@/modules/objectives/types";

export type KeyResultWithTeam = KeyResult & {
  objectiveName: string;
  teamId: string;
  teamName: string;
  teamCode: string;
  workspaceId: string;
};

export type KeyResultListResponse = {
  keyResults: KeyResultWithTeam[];
  totalCount: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
};

export type KeyResultFilters = {
  objectiveIds?: string[];
  teamIds?: string[];
  measurementTypes?: KeyResult["measurementType"][];
  leadIds?: string[];
  createdAfter?: string;
  createdBefore?: string;
  endDateAfter?: string;
  endDateBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  page?: number;
  pageSize?: number;
  orderBy?: "name" | "created_at" | "updated_at" | "objective_name";
  orderDirection?: "asc" | "desc";
};

export type CreateKeyResultRequest = {
  objectiveId: string;
  name: string;
  measurementType: KeyResult["measurementType"];
  startValue?: number;
  currentValue?: number;
  targetValue?: number;
  lead?: string | null;
  contributors?: string[];
  startDate: string;
  endDate: string;
};

export type UpdateKeyResultRequest = {
  name?: string;
  measurementType?: KeyResult["measurementType"];
  startValue?: number;
  currentValue?: number;
  targetValue?: number;
  lead?: string | null;
  contributors?: string[];
  startDate?: string;
  endDate?: string;
};
