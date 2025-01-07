import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { linkKeys } from "@/constants/keys";
import { createLinkAction, NewLink } from "../actions/links/create-link";
import { Link } from "@/types";

export const useCreateLinkMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (newLink: NewLink) => createLinkAction(newLink),
    onError: (_, variables) => {
      toast.error("Failed to create link", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(variables),
        },
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
          linkKeys.story(newLink?.storyId as string),
          [...previousLinks, optimisticLink],
        );
      }
    },
    onSettled: (newLink) => {
      queryClient.invalidateQueries({
        queryKey: linkKeys.story(newLink?.storyId as string),
      });
    },
  });

  return mutation;
};
