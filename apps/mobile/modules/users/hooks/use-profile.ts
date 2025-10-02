import { useQuery } from "@tanstack/react-query";
import { getProfile } from "../queries/get-profile";
import { userKeys } from "@/constants/keys";

export const useProfile = () => {
  return useQuery({
    queryKey: userKeys.profile(),
    queryFn: getProfile,
  });
};
