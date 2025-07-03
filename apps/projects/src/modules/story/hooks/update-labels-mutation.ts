import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { GroupedStoriesResponse } from "@/modules/stories/types";
import { labelKeys } from "@/constants/keys";
import type { DetailedStory } from "../types";
import { updateLabelsAction } from "../actions/update-labels";

export const useUpdateLabelsMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ storyId, labels }: { storyId: string; labels: string[] }) =>
      updateLabelsAction(storyId, labels),

    onMutate: ({ storyId, labels }) => {
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          labels,
        });
      }

      const activeQueries = queryClient.getQueryCache().getAll();

      activeQueries.forEach((query) => {
        if (
          query.queryKey.includes("stories") &&
          query.queryKey.includes("list")
        ) {
          queryClient.setQueryData<GroupedStoriesResponse>(
            query.queryKey,
            (data) => {
              if (!data) return data;

              return {
                ...data,
                groups: data.groups.map((group) => ({
                  ...group,
                  stories: group.stories.map((story) =>
                    story.id === storyId ? { ...story, labels } : story,
                  ),
                })),
              };
            },
          );
        }
      });

      return { previousStory };
    },
    onError: (error, variables, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(variables.storyId),
          context.previousStory,
        );
      }
      toast.error("Failed to update labels", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res, { storyId }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId) });
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
      queryClient.invalidateQueries({ queryKey: labelKeys.lists() });
    },
  });

  return mutation;
};
