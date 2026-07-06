import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { revokeCalendarConnectionAction } from "@/lib/actions/calendar/revoke-connection";

export const useRevokeCalendarConnection = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = calendarKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (connectionId: string) =>
      revokeCalendarConnectionAction(workspaceSlug, connectionId),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Calendar", { description: res.error.message });
        return;
      }
      toast.success("Calendar disconnected");
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
