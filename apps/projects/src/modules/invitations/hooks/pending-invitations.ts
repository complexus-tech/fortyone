import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { invitationKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getPendingInvitations } from "../queries/pending-invitations";

export const usePendingInvitations = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: invitationKeys.pending(workspaceSlug),
    queryFn: () => getPendingInvitations({ session: session!, workspaceSlug }),
  });
};
