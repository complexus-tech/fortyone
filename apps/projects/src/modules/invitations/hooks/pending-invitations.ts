import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { invitationKeys } from "@/constants/keys";
import { getPendingInvitations } from "../queries/pending-invitations";

export const usePendingInvitations = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: invitationKeys.pending,
    queryFn: () => getPendingInvitations(session!),
  });
};
