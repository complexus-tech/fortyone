import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { labelKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getLabels } from "../queries/labels/get-labels";

export const useLabels = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: labelKeys.lists(workspaceSlug),
    queryFn: () => getLabels({ session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
