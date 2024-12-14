import { useQuery } from "@tanstack/react-query";
import { getMembers } from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";

export const useMembers = () => {
  return useQuery({
    queryKey: memberKeys.lists(),
    queryFn: getMembers,
  });
};
