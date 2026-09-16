import type { Objective, ObjectiveHealth } from "./types";
import type { StoryPriority } from "@/modules/stories/types";
import { useInfiniteQuery } from "@tanstack/react-query";
import {
  objectiveKeys,
  homeKeys,
  searchKeys,
  storyKeys,
} from "@/constants/keys";
import { readEntity, writeEntity } from "@/modules/entity-details/http";
import { useDetailMutation } from "@/modules/entity-details/use-detail-mutation";

export type ObjectiveActivity = {
  id: string;
  type: string;
  updateType: string;
  field: string;
  currentValue: string;
  comment: string;
  userId: string;
  createdAt: string;
  user?: {
    fullName: string;
    username: string;
    avatarUrl: string;
    isActive: boolean;
  };
};
type ActivityPage = {
  activities: ObjectiveActivity[];
  pagination: { page: number; hasMore: boolean };
};
export type ObjectivePatch = {
  name?: string;
  description?: string | null;
  statusId?: string;
  priority?: StoryPriority;
  leadUser?: string | null;
  health?: ObjectiveHealth;
  startDate?: string | null;
  endDate?: string | null;
  isPrivate?: boolean;
  comment?: string;
  expectedUpdatedAt?: string;
};
export const readObjective = (id: string) =>
  readEntity<Objective>(`objectives/${encodeURIComponent(id)}`);
export function useObjectiveActivity(id: string, enabled: boolean) {
  return useInfiniteQuery({
    queryKey: [...objectiveKeys.detail(id), "activities"],
    initialPageParam: 1,
    queryFn: ({ pageParam, signal }) =>
      readEntity<ActivityPage>(
        `objectives/${encodeURIComponent(id)}/activities?page=${pageParam}&pageSize=20`,
        signal,
      ),
    getNextPageParam: (page) =>
      page.pagination.hasMore ? page.pagination.page + 1 : undefined,
    enabled,
  });
}
export function useObjectiveCommands(id: string) {
  return useDetailMutation(
    (command: { type: "update"; patch: ObjectivePatch } | { type: "delete" }) =>
      writeEntity(
        command.type === "delete" ? "delete" : "put",
        `objectives/${encodeURIComponent(id)}`,
        command.type === "update" ? command.patch : undefined,
      ),
    [objectiveKeys.all, storyKeys.all, homeKeys.all, searchKeys.all],
  );
}
