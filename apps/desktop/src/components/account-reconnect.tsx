import type { BrowserSignIn } from "../hooks/use-browser-sign-in";
import { ErrorNotice, Spinner } from "./primitives";

export function SignInFeedback({ connection }: { connection: BrowserSignIn }) {
  return (
    <>
      {connection.opened ? (
        <p className="footnote" role="status">
          Finish signing in with this account in your browser. Your local work
          stays on this device.
        </p>
      ) : null}
      <ErrorNotice message={connection.error} />
    </>
  );
}

export function AccountReconnect({
  connection,
}: {
  connection: BrowserSignIn;
}) {
  return (
    <>
      <div className="setting-row">
        <span>
          <strong>Reconnect your account</strong>
          <small>
            Sign in again to restore synchronization. Downloaded work and
            unsynced drafts stay on this device.
          </small>
        </span>
        <button
          disabled={connection.busy || !navigator.onLine}
          onClick={() => {
            void connection.begin();
          }}
        >
          {connection.busy ? <Spinner /> : null}Reconnect account
        </button>
      </div>
      <SignInFeedback connection={connection} />
    </>
  );
}
