import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { getCalendarIntegration } from "@/lib/queries/calendar/get-integration";

export const useCalendarIntegration = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: calendarKeys.integration(workspaceSlug),
    queryFn: () => getCalendarIntegration({ session: session!, workspaceSlug }),
  });
};
