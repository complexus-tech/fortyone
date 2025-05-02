import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { invitationKeys } from "@/constants/keys";
import { getMyInvitations } from "../queries/my-invitations";

export const useMyInvitations = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: invitationKeys.mine,
    queryFn: () => getMyInvitations(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
