import { useQuery } from "@tanstack/react-query";
import { invitationKeys } from "@/constants/keys";
import { getMyInvitations } from "../queries/my-invitations";

export const useMyInvitations = () => {
  return useQuery({
    queryKey: invitationKeys.mine,
    queryFn: getMyInvitations,
  });
};
