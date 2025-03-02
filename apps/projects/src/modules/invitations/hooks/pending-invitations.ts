import { useQuery } from "@tanstack/react-query";
import { invitationKeys } from "@/constants/keys";
import { getPendingInvitations } from "../queries/pending-invitations";

export const usePendingInvitations = () => {
  return useQuery({
    queryKey: invitationKeys.pending,
    queryFn: getPendingInvitations,
  });
};
