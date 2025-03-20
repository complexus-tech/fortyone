import { useQuery } from "@tanstack/react-query";
import { linkKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getLinks } from "../queries/links/get-links";

export const useLinks = (storyId: string) => {
  return useQuery({
    queryKey: linkKeys.story(storyId),
    queryFn: () => getLinks(storyId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 30,
    gcTime: Number(DURATION_FROM_MILLISECONDS.HOUR),
  });
};
