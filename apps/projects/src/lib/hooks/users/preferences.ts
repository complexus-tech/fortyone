import { useQuery } from "@tanstack/react-query";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { userKeys } from "@/constants/keys";

export const useAutomationPreferences = () => {
  return useQuery({
    queryKey: userKeys.automationPreferences(),
    queryFn: () => getAutomationPreferences(),
  });
};
