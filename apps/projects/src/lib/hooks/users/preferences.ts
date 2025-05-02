import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { userKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useAutomationPreferences = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: userKeys.automationPreferences(),
    queryFn: () => getAutomationPreferences(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
