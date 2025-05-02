import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { workspaceKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getWorkspaceSettings } from "../../queries/workspaces/get-settings";

export const useWorkspaceSettings = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: workspaceKeys.settings(),
    queryFn: () => getWorkspaceSettings(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
