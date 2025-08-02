import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/utils";
import { getProfile } from "../queries/profile";

export const userKeys = {
  all: ["users"] as const,
  profile: () => [...userKeys.all, "profile"] as const,
  automationPreferences: () =>
    [...userKeys.all, "automation-preferences"] as const,
};

export const useProfile = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
    enabled: Boolean(session),
  });
};
