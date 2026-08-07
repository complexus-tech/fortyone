import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { revokeCalendarConnectionAction } from "@/lib/actions/calendar/revoke-connection";
import type { CalendarSchedule } from "@/lib/queries/calendar/types";
import type { CalendarIntegration } from "@/modules/settings/workspace/integrations/calendar/types";

export const useRevokeCalendarConnection = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const allQueryKey = calendarKeys.all(workspaceSlug);
  const integrationQueryKey = calendarKeys.integration(workspaceSlug);
  const eventQueryKey = calendarKeys.events(workspaceSlug);
  const scheduleQueryKey = calendarKeys.schedules(workspaceSlug);

  return useMutation({
    mutationFn: (connectionId: string) =>
      revokeCalendarConnectionAction(workspaceSlug, connectionId),
    onSuccess: async (res, connectionId) => {
      if (res.error?.message) {
        toast.error("Calendar", { description: res.error.message });
        return;
      }

      await queryClient.cancelQueries({ queryKey: allQueryKey });
      queryClient.setQueryData<CalendarIntegration>(
        integrationQueryKey,
        (integration) =>
          integration
            ? {
                ...integration,
                connections: integration.connections.filter(
                  (connection) => connection.id !== connectionId,
                ),
              }
            : integration,
      );
      queryClient.setQueriesData<CalendarSchedule>(
        { queryKey: scheduleQueryKey },
        (schedule) =>
          schedule ? { ...schedule, busyWindows: [], events: [] } : schedule,
      );
      queryClient.removeQueries({ queryKey: eventQueryKey });
      toast.success("Calendar disconnected");
      await queryClient.invalidateQueries({ queryKey: allQueryKey });
    },
  });
};
