import type { MediaScope } from "./media-node-view";
import type { CollaborationDraft } from "./collaboration";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import type {
  Bootstrap,
  DraftKind,
  JsonRecord,
  OperationInput,
  Settings,
  WorkspaceData,
  WorkspaceSettings,
} from "./types";

let boundAccountId: string | null = null;
export function bindNativeAccount(accountId: string | null) {
  boundAccountId = accountId;
}
export function getBoundAccountId() {
  if (!boundAccountId)
    throw new Error("Sign in before opening workspace data.");
  return boundAccountId;
}

// Each editor/effect captures its initiating identity. A delayed request from a
// closed account cannot accidentally use the identity of a newly signed-in user.
function accountBridge(account: string | (() => string)) {
  const scoped = <T>(command: string, args: JsonRecord = {}) =>
    invoke<T>(command, {
      ...args,
      expectedAccountId: typeof account === "function" ? account() : account,
    });
  return {
    acknowledgeCollaboration: (
      workspaceId: string,
      documentId: string,
      expectedState: string,
      expectedName: string,
      payload: CollaborationDraft,
    ) =>
      scoped<boolean>("desktop_ack_collaboration", {
        workspaceId,
        documentId,
        expectedState,
        expectedName,
        payload,
      }),
    refreshResource: (scope: MediaScope) =>
      scoped<WorkspaceData>("desktop_refresh_resource", scope),
    readMedia: (scope: MediaScope & { attachmentId: string }) =>
      scoped<{ mimeType: string; data: string } | null>(
        "desktop_read_media",
        scope,
      ),
    cacheMedia: (scope: MediaScope & { attachmentId: string }) =>
      scoped<{ mimeType: string; data: string } | null>(
        "desktop_cache_media",
        scope,
      ),
    exportAttachment: (scope: MediaScope & { attachmentId: string }) =>
      scoped<boolean>("desktop_export_attachment", scope),
    takeNavigation: () =>
      scoped<{ workspaceId: string; kind: string; resourceId?: string } | null>(
        "desktop_take_navigation",
      ),
    markNotificationRead: (
      workspaceId: string,
      notificationId: string,
      observedCreatedAt: string,
    ) =>
      scoped<WorkspaceData>("desktop_mark_notification_read", {
        workspaceId,
        notificationId,
        observedCreatedAt,
      }),
    pendingCollaborations: () =>
      scoped<
        Array<{
          workspaceId: string;
          documentId: string;
          payload: CollaborationDraft;
        }>
      >("desktop_pending_collaborations"),
    openWorkspacePage: (workspaceId: string, path: string) =>
      scoped<void>("desktop_open_workspace_page", { workspaceId, path }),
    readSettings: (workspaceId: string) =>
      scoped<WorkspaceSettings>("desktop_read_settings", { workspaceId }),
    syncSettings: (workspaceId: string) =>
      scoped<WorkspaceSettings>("desktop_sync_settings", { workspaceId }),
    collaborationSession: (workspaceId: string, documentId: string) =>
      scoped<{ token: string; name: string; url: string }>(
        "desktop_collaboration_session",
        { workspaceId, documentId },
      ),
    exportLocalWork: () => scoped<boolean>("desktop_export_local_work"),
    selectWorkspace: (workspaceId: string) =>
      scoped<WorkspaceData>("desktop_select_workspace", { workspaceId }),
    readWorkspace: (workspaceId: string) =>
      scoped<WorkspaceData>("desktop_read_workspace", { workspaceId }),
    sync: (workspaceId: string, retryCredentials = false) =>
      scoped<WorkspaceData>("desktop_sync", { workspaceId, retryCredentials }),
    saveOperation: (input: OperationInput) =>
      scoped<WorkspaceData>("desktop_save_operation", input),
    readDraft: (workspaceId: string, resourceId: string, kind: DraftKind) =>
      scoped<{ payload: JsonRecord } | null>("desktop_read_draft", {
        workspaceId,
        resourceId,
        kind,
      }),
    saveDraft: (
      workspaceId: string,
      resourceId: string,
      kind: DraftKind,
      payload: JsonRecord,
    ) =>
      scoped<void>("desktop_save_draft", {
        workspaceId,
        resourceId,
        kind,
        payload,
      }),
    deleteDraft: (workspaceId: string, resourceId: string, kind: DraftKind) =>
      scoped<void>("desktop_delete_draft", { workspaceId, resourceId, kind }),
    resolveOperation: (
      workspaceId: string,
      operationId: string,
      resolution: "retry" | "discard",
      payload?: JsonRecord,
      baseVersion?: string,
    ) =>
      scoped<WorkspaceData>("desktop_resolve_operation", {
        workspaceId,
        operationId,
        resolution,
        payload,
        baseVersion,
      }),
    stageAttachment: (workspaceId: string, resourceId: string) =>
      scoped<WorkspaceData>("desktop_stage_attachment", {
        workspaceId,
        resourceId,
      }),
    logout: (discardLocalWork = false) =>
      scoped<void>("desktop_logout", { discardLocalWork }),
  };
}

export const native = {
  ...accountBridge(getBoundAccountId),
  forAccount: (accountId: string) => accountBridge(accountId),
  bootstrap: () => invoke<Bootstrap>("desktop_bootstrap"),
  signIn: () => invoke<void>("desktop_sign_in"),
  confirmExit: () => invoke<void>("desktop_confirm_exit"),
  exportDocument: (title: string, contentHtml: string, contentText: string) =>
    invoke<boolean>("desktop_export_document", {
      title,
      contentHtml,
      contentText,
    }),
  settings: () => invoke<Settings>("desktop_settings"),
  updateSettings: (settings: Settings) =>
    invoke<Settings>("desktop_update_settings", { settings }),
  checkUpdate: () =>
    invoke<{ available: boolean; version?: string }>("desktop_check_update"),
  installUpdate: () => invoke<void>("desktop_install_update"),
  openExternal: (url: string) => invoke<void>("desktop_open_external", { url }),
  listen,
};

export function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
