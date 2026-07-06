import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { syncCalendarConnectionAction } from "@/lib/actions/calendar/sync-connection";

export const useSyncCalendarConnection = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const integrationQueryKey = calendarKeys.integration(workspaceSlug);
  const scheduleQueryKey = ["calendar", workspaceSlug, "schedule"] as const;

  return useMutation({
    mutationFn: (connectionId: string) =>
      syncCalendarConnectionAction(workspaceSlug, connectionId),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Calendar", { description: res.error.message });
        return;
      }
      toast.success("Calendar availability synced");
      queryClient.invalidateQueries({ queryKey: integrationQueryKey });
      queryClient.invalidateQueries({ queryKey: scheduleQueryKey });
    },
  });
};
