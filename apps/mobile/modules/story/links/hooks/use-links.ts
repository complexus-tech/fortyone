import { useQuery } from "@tanstack/react-query";
import { getLinks } from "../queries/get-links";

export const useLinks = (storyId: string) => {
  return useQuery({
    queryKey: ["links", storyId],
    queryFn: () => getLinks(storyId),
    enabled: Boolean(storyId),
  });
};
