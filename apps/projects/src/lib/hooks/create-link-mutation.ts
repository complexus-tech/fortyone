import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { linkKeys } from "@/constants/keys";
import type { Link } from "@/types";
import type { NewLink } from "../actions/links/create-link";
import { createLinkAction } from "../actions/links/create-link";

export const useCreateLinkMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (newLink: NewLink) => createLinkAction(newLink),
    onError: (_, variables) => {
      toast.error("Failed to create link", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => { mutation.mutate(variables); },
        },
      });
      queryClient.invalidateQueries({
        queryKey: linkKeys.story(variables.storyId),
      });
    },
    onMutate: (newLink) => {
      const optimisticLink: Link = {
        ...newLink,
        title: newLink.title || "",
        id: "optimistic",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      const previousLinks = queryClient.getQueryData<Link[]>(
        linkKeys.story(newLink.storyId),
      );
      if (previousLinks) {
        queryClient.setQueryData<Link[]>(
          linkKeys.story(newLink.storyId),
          [...previousLinks, optimisticLink],
        );
      }
    },
    onSettled: (newLink) => {
      queryClient.invalidateQueries({
        queryKey: linkKeys.story(newLink?.storyId!),
      });
    },
  });

  return mutation;
};
