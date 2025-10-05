import { useQuery } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { getStoryAttachments } from "../queries/get-attachments";

export const useStoryAttachments = (storyId: string) => {
  return useQuery({
    queryKey: storyKeys.attachments(storyId),
    queryFn: () => getStoryAttachments(storyId),
    enabled: Boolean(storyId),
  });
};
