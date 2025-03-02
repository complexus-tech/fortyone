import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { Story } from "@/modules/stories/types";
import { labelKeys } from "@/constants/keys";
import type { DetailedStory } from "../types";
import { updateLabelsAction } from "../actions/update-labels";

export const useUpdateLabelsMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ storyId, labels }: { storyId: string; labels: string[] }) =>
      updateLabelsAction(storyId, labels),
    onError: (error, variables) => {
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
          queryClient.setQueryData<Story[]>(query.queryKey, (stories) =>
            stories?.map((story) =>
              story.id === storyId ? { ...story, labels } : story,
            ),
          );
        }
      });
    },
    onSettled: (storyId) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId!) });
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
      queryClient.invalidateQueries({ queryKey: labelKeys.lists() });
    },
  });

  return mutation;
};
