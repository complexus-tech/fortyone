import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { getRunningSprints } from "../queries/get-running-sprints";

export const useRunningSprints = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: sprintKeys.running(),
    queryFn: () => getRunningSprints(session!),
  });
};
