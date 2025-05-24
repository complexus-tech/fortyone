import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useSprintStories = (sprintId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: storyKeys.sprint(sprintId),
    queryFn: () => getStories(session!, { sprintId }),
  });
};
