import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { linkKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getLinks } from "../queries/links/get-links";

export const useLinks = (storyId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: linkKeys.story(storyId),
    queryFn: () => getLinks(storyId, session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
