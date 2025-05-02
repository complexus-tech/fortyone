import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryComments } from "../queries/get-comments";

export const useStoryComments = (id: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: storyKeys.comments(id),
    queryFn: () => getStoryComments(id, session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
