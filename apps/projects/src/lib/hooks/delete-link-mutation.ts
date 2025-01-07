import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { linkKeys } from "@/constants/keys";
import { Link } from "@/types";
import { deleteLinkAction } from "../actions/links/delete-link";

export const useDeleteLinkMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ linkId, storyId }: { linkId: string; storyId: string }) =>
      deleteLinkAction(linkId, storyId),
    onError: (_, variables) => {
      toast.error("Failed to delete link", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(variables),
        },
      });
    },
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
    },
    onSettled: (_, __, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: linkKeys.story(storyId),
      });
    },
  });

  return mutation;
};
