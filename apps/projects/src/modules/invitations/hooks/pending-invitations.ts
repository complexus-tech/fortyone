import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { pendingInvitationsQueryOptions } from "../queries/options";

export const usePendingInvitations = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery(
    pendingInvitationsQueryOptions({ session: session!, workspaceSlug }),
  );
};
