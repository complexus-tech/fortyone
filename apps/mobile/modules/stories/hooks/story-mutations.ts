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

    onMutate: async (storyId) => {
      await queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });

      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId)
      );

      // Optimistically update the story
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          archivedAt: new Date().toISOString(),
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
      Alert.alert("Error", "Failed to archive story");
    },

    onSuccess: () => {
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

    onMutate: async (storyId) => {
      await queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });

      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId)
      );

      // Optimistically update the story
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          archivedAt: null,
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
      Alert.alert("Error", "Failed to unarchive story");
    },

    onSuccess: () => {
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
      storyId,
      hardDelete,
    }: {
      storyId: string;
      hardDelete?: boolean;
    }) => deleteStory(storyId, hardDelete),

    onMutate: async ({ storyId }) => {
      await queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });

      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId)
      );

      // Optimistically update the story
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          deletedAt: new Date().toISOString(),
        });
      }

      return { previousStory };
    },

    onError: (error, { storyId }, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(storyId),
          context.previousStory
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Error", "Failed to delete story");
    },

    onSuccess: () => {
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
      await queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });

      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId)
      );

      // Optimistically update the story
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

    onSuccess: () => {
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

    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      Alert.alert("Success", "Story duplicated successfully");
    },
  });

  return mutation;
};
