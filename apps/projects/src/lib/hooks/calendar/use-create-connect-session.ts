import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { createCalendarConnectSessionAction } from "@/lib/actions/calendar/create-connect-session";

export const useCreateCalendarConnectSession = () => {
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: () => createCalendarConnectSessionAction(workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Calendar", { description: res.error.message });
        return;
      }
      const authUrl = res.data?.authUrl;
      if (authUrl) {
        window.location.href = authUrl;
      }
    },
  });
};
