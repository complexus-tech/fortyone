import type { BrowserSignIn } from "../hooks/use-browser-sign-in";
import { AccountReconnect } from "./account-reconnect";
import { useAccountNative } from "../hooks/use-account-native";
import { useEffect, useState } from "react";
import type { Settings as DesktopSettings } from "../types";
import { native, errorMessage } from "../native";
import { applyTheme } from "../theme";
import { flushDrafts } from "../drafts";
import { CachedSettings } from "./cached-settings";
import { ErrorNotice, Spinner } from "./primitives";

export function Settings({
  pendingCount,
  workspaceId,
  onLogout,
  connection,
}: {
  pendingCount: number;
  connection: BrowserSignIn;
  workspaceId?: string;
  onLogout: (discardLocalWork?: boolean) => Promise<void>;
}) {
  const client = useAccountNative();
  const [settings, setSettings] = useState<DesktopSettings | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [update, setUpdate] = useState<{
    available: boolean;
    version?: string;
  } | null>(null);
  const [exported, setExported] = useState(false);
  const [discardLocalWork, setDiscardLocalWork] = useState(false);
  const [signedOutPrompt, setSignedOutPrompt] = useState(false);
  useEffect(() => {
    let active = true;
    void native
      .settings()
      .then((value) => {
        if (active) setSettings(value);
      })
      .catch((reason: unknown) => {
        if (active) setError(errorMessage(reason));
      });
    return () => {
      active = false;
    };
  }, []);
  const save = async (next: DesktopSettings) => {
    setBusy(true);
    setError(null);
    try {
      const saved = await native.updateSettings(next);
      setSettings(saved);
      applyTheme(saved.theme);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };
  const run = async (task: () => Promise<unknown>) => {
    setBusy(true);
    setError(null);
    try {
      await task();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };
  return (
    <section className="collection settings-page">
      <div className="page-heading">
        <div>
          <span className="eyebrow">Make yourself at home</span>
          <h1>Desktop settings</h1>
          <p>A quieter, more connected place to work.</p>
        </div>
      </div>
      <ErrorNotice message={error} />
      {!settings ? (
        <p>
          <Spinner />
          Loading preferences…
        </p>
      ) : (
        <>
          <section className="settings-group">
            <h2>Experience</h2>
            <label className="setting-row">
              <span>
                <strong>Appearance</strong>
                <small>Follow your device or choose a theme.</small>
              </span>
              <select
                aria-label="Appearance"
                value={settings.theme}
                disabled={busy}
                onChange={(event) => {
                  void save({
                    ...settings,
                    theme: event.target.value as DesktopSettings["theme"],
                  });
                }}
              >
                <option value="system">System</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </label>
            <label className="setting-row">
              <span>
                <strong>Desktop notifications</strong>
                <small>Receive notifications while FortyOne is running.</small>
              </span>
              <input
                type="checkbox"
                checked={settings.notifications}
                disabled={busy}
                onChange={(event) => {
                  void save({
                    ...settings,
                    notifications: event.target.checked,
                  });
                }}
              />
            </label>
            <label className="setting-row">
              <span>
                <strong>Launch at login</strong>
                <small>Have your workspace ready when your day starts.</small>
              </span>
              <input
                type="checkbox"
                checked={settings.launchAtLogin}
                disabled={busy}
                onChange={(event) => {
                  void save({
                    ...settings,
                    launchAtLogin: event.target.checked,
                  });
                }}
              />
            </label>
          </section>
          <section className="settings-group">
            <h2>Updates</h2>
            <div className="setting-row">
              <span>
                <strong>FortyOne {settings.appVersion}</strong>
                <small>
                  {!settings.updateConfigured
                    ? "Updates are not configured for this build."
                    : update?.available
                      ? `Version ${update.version} is ready to install.`
                      : update
                        ? "You’re using the latest available version."
                        : "Keep your desktop app up to date."}
                </small>
              </span>
              {update?.available ? (
                <button
                  disabled={busy || pendingCount > 0}
                  onClick={() => {
                    void run(async () => {
                      await flushDrafts();
                      await native.installUpdate();
                    });
                  }}
                >
                  Install & restart
                </button>
              ) : (
                <button
                  disabled={busy || !settings.updateConfigured}
                  onClick={() => {
                    void run(async () => setUpdate(await native.checkUpdate()));
                  }}
                >
                  {busy ? <Spinner /> : null}Check for updates
                </button>
              )}
            </div>
            {pendingCount > 0 ? (
              <p className="footnote">
                Sync or resolve pending changes before installing an update.
              </p>
            ) : null}
          </section>
          <section className="settings-group">
            <h2>Keyboard shortcuts</h2>
            <div className="shortcut-row">
              <span>Search your workspace</span>
              <kbd>⌘ / Ctrl K</kbd>
            </div>
            <div className="shortcut-row">
              <span>New story</span>
              <kbd>⌘ / Ctrl N</kbd>
            </div>
            <div className="shortcut-row">
              <span>Close an editor</span>
              <kbd>Esc</kbd>
            </div>
          </section>
          {workspaceId ? (
            <CachedSettings key={workspaceId} workspaceId={workspaceId} />
          ) : null}
          <section className="settings-group">
            <h2>Account</h2>
            <AccountReconnect connection={connection} />
            <div className="setting-row">
              <span>
                <strong>Export your local work</strong>
                <small>
                  Keep a backup of drafts, pending changes and downloads before
                  signing out.
                </small>
              </span>
              <button
                disabled={busy}
                onClick={() => {
                  void run(async () => {
                    await flushDrafts();
                    setExported(await client.exportLocalWork());
                  });
                }}
              >
                Export backup
              </button>
            </div>
            {exported ? (
              <p className="footnote" role="status">
                Your local work was exported.
              </p>
            ) : null}
            <div className="setting-row">
              <span>
                <strong>Sign out of this device</strong>
                <small>
                  {pendingCount
                    ? `${pendingCount} unsynced changes need to be resolved first.`
                    : "Local account data will be removed from this device."}
                </small>
              </span>
              <button disabled={busy} onClick={() => setSignedOutPrompt(true)}>
                Sign out
              </button>
            </div>
            {signedOutPrompt ? (
              <div className="logout-confirm">
                <p>
                  Sign out and remove downloaded workspace data from this
                  device?
                </p>
                <label className="discard-confirm">
                  <input
                    type="checkbox"
                    checked={discardLocalWork}
                    onChange={(event) =>
                      setDiscardLocalWork(event.target.checked)
                    }
                  />
                  Discard any unsynced changes and drafts. I have exported or
                  reviewed my local work.
                </label>
                <button
                  className="danger-button"
                  disabled={busy}
                  onClick={() => {
                    void run(() => onLogout(discardLocalWork));
                  }}
                >
                  Confirm sign out
                </button>
                <button onClick={() => setSignedOutPrompt(false)}>
                  Cancel
                </button>
              </div>
            ) : null}
          </section>
        </>
      )}
    </section>
  );
}
