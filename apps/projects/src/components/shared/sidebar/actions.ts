import { logout } from "auth";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";

export const logOut = async () => {
  try {
    await logout();
  } catch {
    // Best-effort cookie clearing; continue logout flow.
  }
};

export const changeWorkspace = async (workspaceId: string) => {
  await switchWorkspace(workspaceId);
};
