import { useQuery } from "@tanstack/react-query";
import { userKeys } from "@/constants/keys";
import { getProfile } from "../queries/users/profile";

export const useProfile = () => {
  return useQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(),
  });
};
