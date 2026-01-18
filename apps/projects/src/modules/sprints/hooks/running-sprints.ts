import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import { getRunningSprints } from "../queries/get-running-sprints";

export const useRunningSprints = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: sprintKeys.running(workspaceSlug),
    queryFn: () => getRunningSprints({ session: session!, workspaceSlug }),
  });
};
