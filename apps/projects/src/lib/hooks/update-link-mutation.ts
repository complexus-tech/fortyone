import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { linkKeys } from "@/constants/keys";
import type { Link } from "@/types";
import type { UpdateLink } from "../actions/links/update-link";
import { updateLinkAction } from "../actions/links/update-link";

export const useUpdateLinkMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      linkId,
      payload,
    }: {
      linkId: string;
      payload: UpdateLink;
      storyId: string;
    }) => updateLinkAction(linkId, payload),

    onMutate: (newLink) => {
      const previousLinks = queryClient.getQueryData<Link[]>(
        linkKeys.story(newLink.storyId),
      );
      if (previousLinks) {
        const updatedLinks = previousLinks.map((link) =>
          link.id === newLink.linkId ? { ...link, ...newLink.payload } : link,
        );
        queryClient.setQueryData<Link[]>(
          linkKeys.story(newLink.storyId),
          updatedLinks,
        );
      }
      return { previousLinks };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        linkKeys.story(variables.storyId),
        context?.previousLinks,
      );
      toast.error("Failed to update link", {
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
    onSuccess: (res, { storyId }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: linkKeys.story(storyId),
      });
    },
  });

  return mutation;
};
