import { useQuery } from "@tanstack/react-query";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { userKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useAutomationPreferences = () => {
  return useQuery({
    queryKey: userKeys.automationPreferences(),
    queryFn: () => getAutomationPreferences(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
