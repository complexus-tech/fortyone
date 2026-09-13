import { useMutation, useQueryClient } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { createComment } from "../actions/create-comment";

export const useCreateCommentMutation = () => {
  const client = useQueryClient();
  return useMutation({
    mutationFn: createComment,
    retry: false,
    onSuccess: async (_comment, { storyId }) => {
      await Promise.all([
        client.invalidateQueries({
          queryKey: storyKeys.commentsInfinite(storyId),
        }),
        client.invalidateQueries({
          queryKey: storyKeys.activitiesInfinite(storyId),
        }),
      ]);
    },
  });
};
