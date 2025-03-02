import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import nProgress from "nprogress";
import { useRouter } from "next/navigation";
import { slugify } from "@/utils";
import type { Story } from "@/modules/stories/types";
import { storyKeys } from "@/modules/stories/constants";
import { createStoryAction } from "../actions/create-story";

export const useCreateStoryMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();

  const mutation = useMutation({
    mutationFn: createStoryAction,
    onError: (error, story) => {
      const previousStories = queryClient.getQueryData<Story[]>(
        storyKeys.lists(),
      );
      if (previousStories) {
        queryClient.setQueryData<Story[]>(
          storyKeys.lists(),
          previousStories.filter((s) => s.id !== story.id),
        );
      }
      toast.error(`Failed to create story: ${story.title}`, {
        description:
          error.message || "An error occurred while creating the story",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(story);
          },
        },
      });
    },
    onSuccess: (story) => {
      const previousStories = queryClient.getQueryData<Story[]>(
        storyKeys.lists(),
      );
      if (previousStories) {
        queryClient.setQueryData<Story[]>(storyKeys.lists(), [
          ...previousStories,
          story,
        ]);
      }
      toast.success("Success", {
        description: "Story created successfully",
        action: {
          label: "View story",
          onClick: () => {
            nProgress.start();
            router.push(`/story/${story.id}/${slugify(story.title)}`);
          },
        },
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
    },
  });

  return mutation;
};
