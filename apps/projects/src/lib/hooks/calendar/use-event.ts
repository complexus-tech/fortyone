import { useQuery } from "@tanstack/react-query";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getCalendarEvent } from "@/lib/queries/calendar/get-event";

export const useCalendarEvent = (eventId: string | null) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    enabled: Boolean(eventId && session),
    queryKey: calendarKeys.event(workspaceSlug, eventId ?? ""),
    queryFn: () =>
      getCalendarEvent({ session: session!, workspaceSlug }, eventId!),
    staleTime: 1000 * 60,
  });
};
