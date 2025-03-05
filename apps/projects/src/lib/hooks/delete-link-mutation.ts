import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { linkKeys } from "@/constants/keys";
import type { Link } from "@/types";
import { deleteLinkAction } from "../actions/links/delete-link";

export const useDeleteLinkMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ linkId }: { linkId: string; storyId: string }) =>
      deleteLinkAction(linkId),

    onMutate: ({ linkId, storyId }) => {
      const previousLinks = queryClient.getQueryData<Link[]>(
        linkKeys.story(storyId),
      );
      if (previousLinks) {
        queryClient.setQueryData<Link[]>(
          linkKeys.story(storyId),
          previousLinks.filter((link) => link.id !== linkId),
        );
      }
      return { previousLinks };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        linkKeys.story(variables.storyId),
        context?.previousLinks,
      );
      toast.error("Failed to delete link", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: linkKeys.story(variables.storyId),
      });
    },
    onSettled: (_, __, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: linkKeys.story(storyId),
      });
    },
  });

  return mutation;
};
