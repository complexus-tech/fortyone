import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { linkKeys } from "@/constants/keys";
import { Link } from "@/types";
import { UpdateLink, updateLinkAction } from "../actions/links/update-link";

export const useUpdateLinkMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      linkId,
      payload,
      storyId,
    }: {
      linkId: string;
      payload: UpdateLink;
      storyId: string;
    }) => updateLinkAction(linkId, payload, storyId),
    onError: (_, variables) => {
      toast.error("Failed to update link", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(variables),
        },
      });
    },
    onMutate: (newLink) => {
      const previousLinks = queryClient.getQueryData<Link[]>(
        linkKeys.story(newLink.storyId),
      );
      if (previousLinks) {
        const updatedLinks = previousLinks.map((link) =>
          link.id === newLink.linkId ? { ...link, ...newLink.payload } : link,
        );
        queryClient.setQueryData<Link[]>(
          linkKeys.story(newLink?.storyId as string),
          updatedLinks,
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
