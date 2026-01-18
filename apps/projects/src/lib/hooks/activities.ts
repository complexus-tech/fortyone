import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getActivities } from "../queries/activities/get-activities";

export const useActivities = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: ["activities"],
    queryFn: () => getActivities({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
