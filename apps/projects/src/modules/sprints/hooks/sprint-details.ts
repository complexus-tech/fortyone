import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { getSprint } from "../queries/get-sprint-details";

export const useSprint = (sprintId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: sprintKeys.detail(sprintId),
    queryFn: () => getSprint(sprintId, session!),
  });
};
