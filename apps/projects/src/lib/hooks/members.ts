import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { getMembers } from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useMembers = (search = "") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useQuery({
    queryKey: [...memberKeys.lists(workspaceSlug), normalizedSearch],
    queryFn: () =>
      getMembers({ session: session!, workspaceSlug }, normalizedSearch),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
