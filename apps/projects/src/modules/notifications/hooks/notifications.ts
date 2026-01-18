import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { notificationKeys } from "@/constants/keys";
import { getNotifications } from "../queries/get-notifications";

export const useNotifications = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: notificationKeys.all(workspaceSlug),
    queryFn: () => getNotifications({ session: session!, workspaceSlug }),
  });
};
