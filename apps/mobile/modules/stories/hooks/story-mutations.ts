import { useQueryClient } from "@tanstack/react-query";
import { useSessionMutation } from "@/lib/use-session-mutation";
import { toast } from "sonner-native";
import { storyKeys } from "@/constants/keys";
import {
  archiveStory,
  unarchiveStory,
  deleteStory,
  restoreStory,
  duplicateStory,
} from "../actions/story-actions";
import type { DetailedStory } from "../types";
import { optimisticallyUpdateStories, restoreStoryCache } from "../utils/cache";

type StoryUpdate = { storyId: string; patch: Partial<DetailedStory> };

const useStoryActionMutation = <TVariables, TData>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  message: string,
  updates?: (variables: TVariables) => StoryUpdate[],
) => {
  const client = useQueryClient();
  const queryKey = storyKeys.all;
  return useSessionMutation({
    mutationFn,
    onMutate: (variables) =>
      updates
        ? optimisticallyUpdateStories(client, queryKey, updates(variables))
        : Promise.resolve([]),
    onError: (error, _variables, snapshots) => {
      if (snapshots) restoreStoryCache(client, snapshots);
      toast.error("Could not save this change", { description: error.message });
    },
    onSuccess: () => {
      toast.success(message);
    },
    onSettled: () => client.invalidateQueries({ queryKey }),
  });
};

export const useArchiveStoryMutation = () =>
  useStoryActionMutation(archiveStory, "Task archived", (storyIds) =>
    storyIds.map((storyId) => ({
      storyId,
      patch: { archivedAt: new Date().toISOString() },
    })),
  );

export const useUnarchiveStoryMutation = () =>
  useStoryActionMutation(unarchiveStory, "Task unarchived", (storyIds) =>
    storyIds.map((storyId) => ({
      storyId,
      patch: { archivedAt: null },
    })),
  );

export const useDeleteStoryMutation = () =>
  useStoryActionMutation(deleteStory, "Task deleted", (storyId) => [
    { storyId, patch: { deletedAt: new Date().toISOString() } },
  ]);

export const useRestoreStoryMutation = () =>
  useStoryActionMutation(restoreStory, "Task restored", (storyId) => [
    { storyId, patch: { deletedAt: null } },
  ]);

export const useDuplicateStoryMutation = () =>
  useStoryActionMutation(duplicateStory, "Task duplicated");
