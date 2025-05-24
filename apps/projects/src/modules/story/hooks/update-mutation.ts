import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { Story } from "@/modules/stories/types";
import type { DetailedStory } from "../types";
import { updateStoryAction } from "../actions/update-story";

export const useUpdateStoryMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();
  const params = useParams<{ storyId: string }>();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload),

    onMutate: ({ storyId, payload }) => {
      queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );

      const activeQueries = queryClient.getQueryCache().getAll();

      // update parent story if it exists
      if (params.storyId !== storyId) {
        const parentStory = queryClient.getQueryData<DetailedStory>(
          storyKeys.detail(params.storyId),
        );
        if (parentStory?.subStories) {
          queryClient.setQueryData<DetailedStory>(
            storyKeys.detail(params.storyId),
            {
              ...parentStory,
              subStories: parentStory.subStories.map((subStory) =>
                subStory.id === storyId
                  ? { ...subStory, ...payload }
                  : subStory,
              ),
            },
          );
        }
      }

      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          if (!queryKey.toLowerCase().includes("detail")) {
            queryClient.setQueryData<Story[]>(query.queryKey, (stories) =>
              stories?.map((story) =>
                story.id === storyId ? { ...story, ...payload } : story,
              ),
            );
          }
        }
      });

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          ...payload,
        });
        return { previousStory };
      }
    },
    onError: (error, variables, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(variables.storyId),
          context.previousStory,
        );
      }
      toast.error("Failed to update story", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res, { storyId, payload }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("story_updated", {
        storyId,
        ...payload,
      });

      if (params.storyId !== storyId) {
        queryClient.invalidateQueries({
          queryKey: storyKeys.detail(params.storyId),
        });
      }
      const activeQueries = queryClient.getQueryCache().getAll();
      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });
    },
  });

  return mutation;
};
