import { useQuery } from "@tanstack/react-query";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { invitationKeys } from "@/constants/keys";
import { getMyInvitations } from "../queries/my-invitations";

export const useMyInvitations = () => {
  return useQuery({
    queryKey: invitationKeys.mine,
    queryFn: () => getMyInvitations(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
