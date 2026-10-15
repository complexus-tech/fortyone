import { BellIcon } from "icons/src/bell";
import { WorkIcon } from "icons/src/work";
import { StoriesIcon } from "icons/src/stories";
import { DocsIcon } from "icons/src/docs";
import { RefreshIcon } from "icons/src/refresh";
import { SettingsIcon } from "icons/src/settings";
import type { Bootstrap, WorkspaceData } from "../types";
import { Spinner } from "./primitives";

export type Page =
  | "my-work"
  | "stories"
  | "documents"
  | "sync"
  | "settings"
  | "notifications";
const pages = [
  { id: "notifications", label: "Inbox", Icon: BellIcon },
  { id: "my-work", label: "My work", Icon: WorkIcon },
  { id: "stories", label: "All stories", Icon: StoriesIcon },
  { id: "documents", label: "Documents", Icon: DocsIcon },
] as const;

export function Sidebar({
  session,
  data,
  page,
  onPage,
  onWorkspace,
  switchingId,
  online,
  syncing,
  collaborativePending = 0,
  onOpenBrowser,
}: {
  session: Bootstrap;
  data: WorkspaceData | null;
  page: Page;
  onPage: (page: Page) => void;
  onWorkspace: (id: string) => void;
  switchingId: string | null;
  online: boolean;
  syncing: boolean;
  collaborativePending?: number;
  onOpenBrowser?: () => void;
}) {
  const workspace = session.workspaces.find(
    (item) => item.id === session.activeWorkspaceId,
  );
  const target = session.workspaces.find((item) => item.id === switchingId);
  const pending = (data?.pending.length ?? 0) + collaborativePending;
  const downloading =
    data?.downloadState &&
    !["ready", "paused"].includes(data.downloadState.stage);
  const preparing = Boolean(session.activeWorkspaceId) && !data?.lastSyncedAt;
  const paused = data?.downloadState?.stage === "paused";
  const downloadDetail =
    data?.downloadState?.stage === "settings"
      ? "Downloading settings"
      : data?.downloadState?.stage === "my-work"
        ? "Downloading My work"
        : data?.downloadState?.stage === "recent"
          ? "Downloading recent work"
          : data?.downloadState?.stage === "documents"
            ? "Downloading documents"
            : "Getting My work and settings ready";
  const attention = data?.pending.some(
    (operation) => operation.status !== "pending",
  );
  return (
    <aside className="sidebar">
      <div className="sidebar-titlebar" data-tauri-drag-region />
      <div className="workspace-switcher" aria-busy={Boolean(switchingId)}>
        <span className="workspace-avatar">
          {(workspace?.name ?? "41").slice(0, 2).toUpperCase()}
        </span>
        <label className="sr-only" htmlFor="workspace">
          Workspace
        </label>
        <select
          id="workspace"
          value={switchingId ?? session.activeWorkspaceId ?? ""}
          disabled={Boolean(switchingId)}
          onChange={(event) => onWorkspace(event.target.value)}
        >
          <option value="" disabled>
            Select a workspace
          </option>
          {session.workspaces.map((item) => (
            <option key={item.id} value={item.id}>
              {item.name}
            </option>
          ))}
        </select>
      </div>
      {target ? (
        <div className="workspace-progress" role="status">
          <Spinner />
          Switching to {target.name}…
        </div>
      ) : null}
      <nav aria-label="Workspace navigation">
        <span className="nav-caption">Workspace</span>
        {pages.map(({ id, label, Icon }) => (
          <button
            key={id}
            aria-current={page === id ? "page" : undefined}
            onClick={() => onPage(id)}
          >
            <Icon />
            {label}
            {id === "notifications" && data?.unreadNotificationCount ? (
              <span className="nav-badge">{data.unreadNotificationCount}</span>
            ) : null}
          </button>
        ))}
        {onOpenBrowser ? (
          <button
            onClick={onOpenBrowser}
            disabled={!online}
            className="browser-link"
          >
            Open full workspace ↗
          </button>
        ) : null}
      </nav>
      <div className="sidebar-bottom">
        <button
          className={`sync-summary ${attention ? "needs-attention" : ""}`}
          onClick={() => onPage("sync")}
        >
          {online && (syncing || downloading || (preparing && !paused)) ? (
            <Spinner />
          ) : (
            <span className={`connection-dot ${online ? "connected" : ""}`} />
          )}
          <span aria-live="polite">
            {attention
              ? "Needs your attention"
              : !online
                ? "Working offline"
                : syncing
                  ? "Syncing changes…"
                  : paused
                    ? "Download paused"
                    : downloading
                      ? "Downloading workspace…"
                      : preparing
                        ? "Preparing offline workspace…"
                        : pending
                          ? `${pending} waiting to sync`
                          : "All changes saved"}
            <small>
              {preparing && !online
                ? "Connect to download this workspace"
                : paused
                  ? "Downloaded items remain available"
                  : preparing || downloading
                    ? downloadDetail
                    : !online
                      ? "Changes stay on this device"
                      : data?.complete &&
                          Object.values(data.complete).some(
                            (complete) => !complete,
                          )
                        ? "Recent work downloaded"
                        : "Available on this device"}
            </small>
          </span>
          <RefreshIcon />
        </button>
        <button
          className="settings-link"
          aria-current={page === "settings" ? "page" : undefined}
          onClick={() => onPage("settings")}
        >
          <SettingsIcon />
          Settings
          <span className="user-initial">
            {(session.user?.fullName || session.user?.email || "U").slice(0, 1)}
          </span>
        </button>
      </div>
    </aside>
  );
}
