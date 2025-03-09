import { useCurrentWorkspace } from "@/lib/hooks/workspaces";

export const useUserRole = () => {
  const { workspace } = useCurrentWorkspace();
  return {
    userRole: workspace?.userRole,
  };
};
