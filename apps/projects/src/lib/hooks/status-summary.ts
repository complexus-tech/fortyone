import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStatusSummary } from "../queries/analytics/get-status-summary";

export const useStatusSummary = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: ["status-summary"],
    queryFn: () => getStatusSummary({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
