import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { StoryActivity } from "@/modules/stories/types";
import { commentStoryAction } from "../actions/comment-story";

export const useCommentStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: {
        comment: string;
        parentId?: string | null;
        userId: string;
      };
    }) =>
      commentStoryAction(storyId, {
        comment: payload?.comment,
        parentId: payload?.parentId,
      }),
    onError: (_, variables) => {
      toast.error("Failed to comment story", {
        description: "Your comment was not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(variables),
        },
      });
    },
    onMutate: ({ storyId, payload }) => {
      const previousActivities = queryClient.getQueryData<StoryActivity[]>(
        storyKeys.activities(storyId),
      );
      if (previousActivities) {
        queryClient.setQueryData<StoryActivity[]>(
          storyKeys.activities(storyId),
          [
            ...previousActivities,
            {
              id: payload?.userId,
              userId: payload?.userId,
              field: "comment",
              storyId: storyId,
              type: "comment",
              createdAt: new Date().toISOString(),
              currentValue: payload.comment,
              parentId: payload.parentId ?? null,
            },
          ],
        );
      }
    },
    onSettled: (_, __, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: storyKeys.activities(storyId!),
      });
    },
  });

  return mutation;
};
