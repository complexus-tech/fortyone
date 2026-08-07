import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { syncCalendarConnectionAction } from "@/lib/actions/calendar/sync-connection";

type SyncCalendarConnectionInput = {
  connectionId: string;
  silent?: boolean;
};

export const useSyncCalendarConnection = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const integrationQueryKey = calendarKeys.integration(workspaceSlug);
  const eventQueryKey = calendarKeys.events(workspaceSlug);
  const scheduleQueryKey = calendarKeys.schedules(workspaceSlug);

  return useMutation({
    mutationFn: ({ connectionId }: SyncCalendarConnectionInput) =>
      syncCalendarConnectionAction(workspaceSlug, connectionId),
    onSuccess: async (res, { silent }) => {
      if (res.error?.message) {
        if (!silent) {
          toast.error("Calendar", { description: res.error.message });
        }
        await queryClient.invalidateQueries({ queryKey: integrationQueryKey });
        return;
      }
      if (!silent) {
        toast.success("Calendar synced");
      }
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: integrationQueryKey }),
        queryClient.invalidateQueries({ queryKey: eventQueryKey }),
        queryClient.invalidateQueries({ queryKey: scheduleQueryKey }),
      ]);
    },
  });
};
