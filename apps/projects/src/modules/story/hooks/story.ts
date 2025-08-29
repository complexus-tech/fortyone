import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStory } from "../queries/get-story";

export const useStoryById = (id: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: storyKeys.detail(id),
    queryFn: () => getStory(id, session!),
    staleTime: Number(DURATION_FROM_MILLISECONDS.MINUTE),
  });
};
