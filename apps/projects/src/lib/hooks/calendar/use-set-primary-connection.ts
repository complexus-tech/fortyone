import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { setPrimaryCalendarConnectionAction } from "@/lib/actions/calendar/set-primary-connection";
import type { CalendarIntegration } from "@/modules/settings/workspace/integrations/calendar/types";

export const useSetPrimaryCalendarConnection = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const integrationQueryKey = calendarKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (connectionId: string) =>
      setPrimaryCalendarConnectionAction(workspaceSlug, connectionId),
    onSuccess: async (res, connectionId) => {
      if (res.error?.message) {
        toast.error("Calendar", { description: res.error.message });
        return;
      }

      await queryClient.cancelQueries({ queryKey: integrationQueryKey });
      queryClient.setQueryData<CalendarIntegration>(
        integrationQueryKey,
        (integration) =>
          integration
            ? {
                ...integration,
                connections: integration.connections.map((connection) => ({
                  ...connection,
                  isPrimary: connection.id === connectionId,
                })),
              }
            : integration,
      );
      await queryClient.invalidateQueries({ queryKey: integrationQueryKey });
    },
  });
};
