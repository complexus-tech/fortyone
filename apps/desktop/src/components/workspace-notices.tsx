import type { DownloadState } from "../types";
import type { BrowserSignIn } from "../hooks/use-browser-sign-in";
import { authRecovery } from "../auth-recovery";
import { SignInFeedback } from "./account-reconnect";
import { ErrorNotice, Spinner } from "./primitives";

export function WorkspaceNotices({
  error,
  onDismissError,
  needsSignIn,
  download,
  syncing,
  onRetry,
  connection,
}: {
  error: string | null;
  onDismissError: () => void;
  needsSignIn: boolean;
  download?: DownloadState;
  syncing: boolean;
  onRetry: () => void;
  connection: BrowserSignIn;
}) {
  const errorRecovery = authRecovery(error);
  const downloadRecovery =
    download?.stage === "paused" ? authRecovery(download.error) : null;
  const recovery =
    errorRecovery ?? downloadRecovery ?? (needsSignIn ? "sign-in" : null);
  return (
    <>
      <ErrorNotice
        message={errorRecovery ? null : error}
        onDismiss={onDismissError}
      />
      {recovery ? (
        <div className="reauth-notice" role="alert">
          <span>
            {recovery === "credential-access"
              ? "Secure credential access is unavailable. Unlock or allow access to your system credential store, then retry. Your downloaded work and drafts are retained."
              : "Sign in again to synchronize. Your downloaded work and unsynced drafts are retained."}
          </span>
          {recovery === "credential-access" ? (
            <button disabled={syncing || !navigator.onLine} onClick={onRetry}>
              {syncing ? <Spinner /> : null}Sync now
            </button>
          ) : (
            <button
              disabled={connection.busy || !navigator.onLine}
              onClick={() => {
                void connection.begin();
              }}
            >
              {connection.busy ? <Spinner /> : null}Sign in to sync
            </button>
          )}
          {recovery === "sign-in" ? (
            <SignInFeedback connection={connection} />
          ) : null}
        </div>
      ) : null}
      {download && !["ready", "paused"].includes(download.stage) ? (
        <div className="download-progress" role="status">
          <Spinner />
          <span>
            Preparing offline workspace ·{" "}
            {download.stage === "my-work"
              ? "Downloading My work"
              : download.stage === "settings"
                ? "Downloading settings"
                : download.stage === "recent"
                  ? "Downloading recent work"
                  : "Downloading documents"}
            {download.downloaded ? ` (${download.downloaded})` : ""}
          </span>
        </div>
      ) : null}
      {download?.stage === "paused" && !downloadRecovery ? (
        <div className="download-progress" role="status">
          <span>
            Offline download paused. Downloaded work remains available.{" "}
            {download.error ?? "It will resume when a connection is available."}
          </span>
          <button disabled={syncing || !navigator.onLine} onClick={onRetry}>
            Retry download
          </button>
        </div>
      ) : null}
    </>
  );
}
