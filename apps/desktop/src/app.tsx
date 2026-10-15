import { useState } from "react";
import { useBrowserSignIn } from "./hooks/use-browser-sign-in";
import { WorkspaceNotices } from "./components/workspace-notices";
import { RefreshIcon } from "icons/src/refresh";
import { useResourceRefresh } from "./hooks/use-resource-refresh";
import { useDesktop } from "./hooks/use-desktop";
import { native, errorMessage } from "./native";
import { useCollaborativeSync } from "./hooks/use-collaborative-sync";
import { useDesktopEvents, type Selection } from "./hooks/use-desktop-events";
import type { Page } from "./components/sidebar";
import { Sidebar } from "./components/sidebar";
import { StoryList } from "./components/story-list";
import { StoryEditor } from "./components/story-editor";
import { Documents, DocumentEditor } from "./components/documents";
import { Settings } from "./components/settings";
import { Notifications } from "./components/notifications";
import { SyncPanel } from "./components/sync-panel";
import { Empty, ErrorNotice, Spinner } from "./components/primitives";

export function App() {
  const desktop = useDesktop();
  const { session, data } = desktop;
  const [page, setPage] = useState<Page>("my-work");
  const [selection, setSelection] = useState<Selection>(null);
  const connection = useBrowserSignIn(session?.user?.id ?? null);
  const collaborativeSync = useCollaborativeSync(session?.user?.id);
  useDesktopEvents(desktop, setPage, setSelection);
  useResourceRefresh(selection, desktop, setSelection);

  if (desktop.starting)
    return (
      <div className="launch-screen">
        <div className="brand-mark">41</div>
        <h1>FortyOne</h1>
        <p role="status">
          <Spinner />
          Opening your workspace…
        </p>
      </div>
    );
  if (!session?.user)
    return (
      <div className="launch-screen">
        <div className="launch-drag" data-tauri-drag-region />
        <div className="brand-mark">41</div>
        <span className="eyebrow">A little more focus. A lot more done.</span>
        <h1>
          Your work,
          <br />
          right here.
        </h1>
        <p>
          Bring your team’s stories and documents together.
          <br />
          Ready for wherever the day takes you.
        </p>
        <ErrorNotice message={connection.error ?? desktop.error} />
        <button
          className="primary-button"
          disabled={connection.busy || !navigator.onLine}
          onClick={() => {
            void connection.begin();
          }}
        >
          {connection.busy ? <Spinner /> : null}Continue in browser
        </button>
        {connection.opened ? (
          <p className="footnote" role="status">
            Finish signing in in your browser. FortyOne will open automatically.
          </p>
        ) : null}
        {!desktop.online ? (
          <p className="footnote">
            Connect to the internet to sign in and download your first
            workspace.
          </p>
        ) : null}
        {desktop.error ? (
          <button
            className="text-button"
            onClick={() => {
              void desktop.reload();
            }}
          >
            Try opening again
          </button>
        ) : null}
        <span className="launch-note">
          Saved on your device. Connected to your team.
        </span>
      </div>
    );
  const workspace = session.workspaces.find(
    (item) => item.id === session.activeWorkspaceId,
  );
  const target = session.workspaces.find(
    (item) => item.id === desktop.switchingId,
  );
  return (
    <div className="app-shell">
      <Sidebar
        session={session}
        data={data}
        page={page}
        onPage={setPage}
        onWorkspace={(id) => {
          void desktop.selectWorkspace(id).then((selected) => {
            if (selected) setSelection(null);
          });
        }}
        onOpenBrowser={
          data
            ? () => {
                void native
                  .forAccount(session.user!.id)
                  .openWorkspacePage(data.workspaceId, "/my-work")
                  .catch((reason: unknown) =>
                    desktop.setError(errorMessage(reason)),
                  );
              }
            : undefined
        }
        collaborativePending={collaborativeSync.pendingCount}
        switchingId={desktop.switchingId}
        online={desktop.online}
        syncing={desktop.syncing}
      />
      <main className="main-pane">
        <header className="app-header" data-tauri-drag-region>
          <span>{workspace?.name ?? "FortyOne"}</span>
          <span className="header-divider">/</span>
          <span>
            {page === "my-work"
              ? "My work"
              : page === "sync"
                ? "Sync & offline"
                : page === "settings"
                  ? "Settings"
                  : page === "stories"
                    ? "Stories"
                    : page === "notifications"
                      ? "Inbox"
                      : "Documents"}
          </span>
          <div className="header-actions">
            {!desktop.online ? (
              <span className="offline-pill">Offline</span>
            ) : null}
            <button
              className="icon-button"
              title="Sync workspace"
              aria-label="Sync workspace"
              disabled={!data || desktop.syncing || !navigator.onLine}
              onClick={() => {
                void desktop.sync(true);
              }}
            >
              {desktop.syncing ? <Spinner /> : <RefreshIcon />}
            </button>
          </div>
        </header>
        <WorkspaceNotices
          error={desktop.error}
          onDismissError={() => desktop.setError(null)}
          needsSignIn={session.needsSignIn}
          download={data?.downloadState}
          syncing={desktop.syncing}
          onRetry={() => {
            void desktop.sync(true);
          }}
          connection={connection}
        />
        <div className="main-content" inert={Boolean(desktop.switchingId)}>
          {page === "settings" ? (
            <Settings
              key={`${session.user.id}:${data?.workspaceId}`}
              workspaceId={data?.workspaceId}
              pendingCount={
                (data?.pending.length ?? 0) + collaborativeSync.pendingCount
              }
              onLogout={desktop.logout}
              connection={connection}
            />
          ) : !data ? (
            <Empty title="Choose your workspace">
              Select a workspace from the sidebar to download it and start
              working.
            </Empty>
          ) : page === "notifications" ? (
            <Notifications
              data={data}
              acceptData={desktop.acceptData}
              onStory={(story) =>
                setSelection({
                  accountId: session.user!.id,
                  workspaceId: data.workspaceId,
                  kind: "story",
                  story,
                })
              }
            />
          ) : page === "documents" ? (
            <Documents
              data={data}
              onOpen={(document) =>
                setSelection({
                  accountId: session.user!.id,
                  workspaceId: data!.workspaceId,
                  kind: "document",
                  document,
                })
              }
              onCreate={() =>
                setSelection({
                  accountId: session.user!.id,
                  workspaceId: data!.workspaceId,
                  kind: "document",
                  document: null,
                })
              }
            />
          ) : page === "sync" ? (
            <SyncPanel
              data={data}
              syncing={desktop.syncing}
              online={desktop.online}
              collaborativePending={collaborativeSync.pendingCount}
              collaborationError={collaborativeSync.error}
              onSync={() => {
                void desktop.sync(true);
              }}
              acceptData={desktop.acceptData}
            />
          ) : (
            <StoryList
              key={`${data.workspaceId}:${page}`}
              data={data}
              userId={session.user.id}
              mine={page === "my-work"}
              onOpen={(story) =>
                setSelection({
                  accountId: session.user!.id,
                  workspaceId: data!.workspaceId,
                  kind: "story",
                  story,
                })
              }
              onCreate={() =>
                setSelection({
                  accountId: session.user!.id,
                  workspaceId: data!.workspaceId,
                  kind: "story",
                  story: null,
                })
              }
            />
          )}
        </div>
        {target ? (
          <div className="workspace-curtain" role="status">
            <Spinner />
            <strong>Switching to {target.name}…</strong>
            <span>Getting your workspace ready</span>
          </div>
        ) : null}
      </main>
      {data &&
      selection?.accountId === session.user.id &&
      selection?.workspaceId === data.workspaceId &&
      selection?.kind === "story" ? (
        <StoryEditor
          key={`${session.user.id}:${data.workspaceId}:${selection.story?.id ?? "new"}`}
          story={selection.story}
          data={data}
          save={desktop.saveOperation}
          acceptData={desktop.acceptData}
          onClose={() => setSelection(null)}
        />
      ) : null}
      {data &&
      selection?.accountId === session.user.id &&
      selection?.workspaceId === data.workspaceId &&
      selection?.kind === "document" ? (
        <DocumentEditor
          key={`${session.user.id}:${data.workspaceId}:${selection.document?.id ?? "new"}`}
          document={selection.document}
          data={data}
          save={desktop.saveOperation}
          onClose={() => setSelection(null)}
        />
      ) : null}
    </div>
  );
}
