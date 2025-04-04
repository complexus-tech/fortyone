import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStoryAttachments } from "../queries/get-attachments";

export const useStoryAttachments = (id: string) => {
  return useQuery({
    queryKey: storyKeys.attachments(id),
    queryFn: () => getStoryAttachments(id),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
