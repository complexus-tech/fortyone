import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Alert } from "react-native";
import { storyKeys } from "@/constants/keys";
import {
  archiveStory,
  unarchiveStory,
  deleteStory,
  restoreStory,
  duplicateStory,
} from "../actions/story-actions";
import type { DetailedStory } from "../types";

// Archive mutation
export const useArchiveStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: archiveStory,

    onMutate: async (storyIds) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: storyKeys.all });

      // Store previous stories for rollback
      const previousStories: Record<string, DetailedStory> = {};

      // Update each story optimistically
      storyIds.forEach((storyId) => {
        const previousStory = queryClient.getQueryData<DetailedStory>(
          storyKeys.detail(storyId)
        );

        if (previousStory) {
          previousStories[storyId] = previousStory;
          queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
            ...previousStory,
            archivedAt: new Date().toISOString(),
          });
        }
      });

      return { previousStories };
    },

    onError: (error, storyIds, context) => {
      // Rollback optimistic updates
      if (context?.previousStories) {
        Object.entries(context.previousStories).forEach(
          ([storyId, previousStory]) => {
            queryClient.setQueryData<DetailedStory>(
              storyKeys.detail(storyId),
              previousStory
            );
          }
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Error", "Failed to archive story");
    },

    onSuccess: (res, storyIds) => {
      if (res.error?.message) {
        Alert.alert("Error", res.error.message);
        return;
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Success", "Story archived successfully");
    },
  });

  return mutation;
};

// Unarchive mutation
export const useUnarchiveStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: unarchiveStory,

    onMutate: async (storyIds) => {
      await queryClient.cancelQueries({ queryKey: storyKeys.all });

      // Store previous stories for rollback
      const previousStories: Record<string, DetailedStory> = {};

      storyIds.forEach((storyId) => {
        const previousStory = queryClient.getQueryData<DetailedStory>(
          storyKeys.detail(storyId)
        );

        if (previousStory) {
          previousStories[storyId] = previousStory;
          queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
            ...previousStory,
            archivedAt: null,
          });
        }
      });

      return { previousStories };
    },

    onError: (error, storyIds, context) => {
      // Rollback optimistic updates
      if (context?.previousStories) {
        Object.entries(context.previousStories).forEach(
          ([storyId, previousStory]) => {
            queryClient.setQueryData<DetailedStory>(
              storyKeys.detail(storyId),
              previousStory
            );
          }
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Error", "Failed to unarchive story");
    },

    onSuccess: (res, storyIds) => {
      if (res.error?.message) {
        Alert.alert("Error", res.error.message);
        return;
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Success", "Story unarchived successfully");
    },
  });

  return mutation;
};

// Delete mutation
export const useDeleteStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      storyIds,
      hardDelete,
    }: {
      storyIds: string[];
      hardDelete?: boolean;
    }) => deleteStory(storyIds, hardDelete),

    onMutate: async ({ storyIds }) => {
      await queryClient.cancelQueries({ queryKey: storyKeys.all });

      // Store previous stories for rollback
      const previousStories: Record<string, DetailedStory> = {};

      storyIds.forEach((storyId) => {
        const previousStory = queryClient.getQueryData<DetailedStory>(
          storyKeys.detail(storyId)
        );

        if (previousStory) {
          previousStories[storyId] = previousStory;
          queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
            ...previousStory,
            deletedAt: new Date().toISOString(),
          });
        }
      });

      return { previousStories };
    },

    onError: (error, { storyIds }, context) => {
      // Rollback optimistic updates
      if (context?.previousStories) {
        Object.entries(context.previousStories).forEach(
          ([storyId, previousStory]) => {
            queryClient.setQueryData<DetailedStory>(
              storyKeys.detail(storyId),
              previousStory
            );
          }
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Error", "Failed to delete story");
    },

    onSuccess: (res, { storyIds }) => {
      if (res.error?.message) {
        Alert.alert("Error", res.error.message);
        return;
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Success", "Story deleted successfully");
    },
  });

  return mutation;
};

// Restore mutation
export const useRestoreStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: restoreStory,

    onMutate: async (storyId) => {
      await queryClient.cancelQueries({ queryKey: storyKeys.all });

      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId)
      );

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          deletedAt: null,
        });
      }

      return { previousStory };
    },

    onError: (error, storyId, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(storyId),
          context.previousStory
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Error", "Failed to restore story");
    },

    onSuccess: (res, storyId) => {
      if (res.error?.message) {
        Alert.alert("Error", res.error.message);
        return;
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Success", "Story restored successfully");
    },
  });

  return mutation;
};

// Duplicate mutation
export const useDuplicateStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: duplicateStory,

    onError: (error) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Error", "Failed to duplicate story");
    },

    onSuccess: (res, storyId) => {
      if (res.error?.message) {
        Alert.alert("Error", res.error.message);
        return;
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Success", "Story duplicated successfully");
    },
  });

  return mutation;
};
