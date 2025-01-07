import { useQuery } from "@tanstack/react-query";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "../queries/links/get-links";

export const useLinks = (storyId: string) => {
  return useQuery({
    queryKey: linkKeys.story(storyId),
    queryFn: () => getLinks(storyId),
  });
};
