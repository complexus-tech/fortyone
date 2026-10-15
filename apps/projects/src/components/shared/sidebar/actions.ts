import { logout } from "auth";
import { flushDesktopAttachments } from "@/shell/desktop/attachments";
import { flushDesktopDrafts, retryDesktopDrafts } from "@/shell/desktop/drafts";
import {
  clearDesktopQueryCache,
  getActiveDesktopCacheScope,
} from "@/shell/desktop/desktop-query-persistence";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";

export const logOut = async () => {
  await flushDesktopAttachments();
  const scope = getActiveDesktopCacheScope();
  if (scope) await retryDesktopDrafts(scope);
  else await flushDesktopDrafts();
  try {
    await clearDesktopQueryCache();
  } catch {
    // Native reads still require the account cookie; logout must remain available.
  }
  try {
    await logout();
  } catch {
    // Best-effort cookie clearing; continue logout flow.
  }
};

export const changeWorkspace = async (
  workspaceId: string,
  signal?: AbortSignal,
) => {
  await switchWorkspace(workspaceId, signal);
};
