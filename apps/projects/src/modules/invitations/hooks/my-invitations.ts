import { useQuery } from "@tanstack/react-query";
import { myInvitationsQueryOptions } from "../queries/options";

export const useMyInvitations = () => {
  return useQuery(myInvitationsQueryOptions());
};
