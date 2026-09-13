import { storyKeys } from "@/constants/keys";
import { useQuery } from "@tanstack/react-query";
import { getLinks } from "../queries/get-links";

export const useLinks = (storyId: string) => {
  return useQuery({
    queryKey: [...storyKeys.detail(storyId), "links"],
    queryFn: ({ signal }) => getLinks(storyId, signal),
    enabled: Boolean(storyId),
  });
};
