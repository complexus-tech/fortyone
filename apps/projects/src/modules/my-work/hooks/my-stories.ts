import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "@/modules/stories/constants";
import { getMyStories } from "../queries/get-stories";

export const useMyStories = () => {
  const { data: session } = useSession();

  return useQuery({
    queryKey: storyKeys.mine(),
    queryFn: () => getMyStories(session!),
    // refetchOnMount: true,
  });
};
