import { useQuery } from "@tanstack/react-query";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getCalendarSchedule } from "@/lib/queries/calendar/get-schedule";

export const useCalendarSchedule = (
  params: {
    startAt: string;
    endAt: string;
  },
  options?: {
    enabled?: boolean;
    refetchInterval?: number | false;
  },
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: calendarKeys.schedule(
      workspaceSlug,
      params.startAt,
      params.endAt,
    ),
    queryFn: () =>
      getCalendarSchedule({ session: session!, workspaceSlug }, params),
    enabled: options?.enabled,
    refetchInterval: options?.refetchInterval,
    staleTime: 1000 * 60,
  });
};
