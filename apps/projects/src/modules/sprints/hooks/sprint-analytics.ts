import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { getSprintAnalytics } from "../queries/get-sprint-analytics";

export const useSprintAnalytics = (sprintId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: sprintKeys.analytics(sprintId),
    queryFn: () => getSprintAnalytics(sprintId, session!),
  });
};
