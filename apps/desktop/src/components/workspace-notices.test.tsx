import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { WorkspaceNotices } from "./workspace-notices";
import { AccountReconnect } from "./account-reconnect";

beforeEach(() => {
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: true,
  });
});
const connection = () => ({
  busy: false,
  opened: false,
  error: null,
  begin: vi.fn().mockResolvedValue(undefined),
});
const paused = (error: string) => ({
  workspaceId: "workspace",
  stage: "paused" as const,
  downloaded: 0,
  error,
});

describe("account recovery actions", () => {
  it("offers browser reconnection for a paused authenticated download even when bootstrap still considers the user signed in", () => {
    const account = connection();
    const retry = vi.fn();
    render(
      <WorkspaceNotices
        error={null}
        onDismissError={vi.fn()}
        needsSignIn={false}
        download={paused(
          "Sign in again to synchronize. Your local work is retained.",
        )}
        syncing={false}
        onRetry={retry}
        connection={account}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Sign in to sync" }));
    expect(account.begin).toHaveBeenCalledOnce();
    expect(retry).not.toHaveBeenCalled();
    expect(
      screen.queryByRole("button", { name: "Retry download" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText(/unsynced drafts are retained/),
    ).toBeInTheDocument();
  });
  it("distinguishes a locked or denied credential store from an expired session and retries only on user request", () => {
    const account = connection();
    const retry = vi.fn();
    render(
      <WorkspaceNotices
        error="Secure credential access was not granted or is unavailable. Allow access, then retry."
        onDismissError={vi.fn()}
        needsSignIn={false}
        syncing={false}
        onRetry={retry}
        connection={account}
      />,
    );
    expect(retry).not.toHaveBeenCalled();
    expect(account.begin).not.toHaveBeenCalled();
    expect(screen.getByText(/Unlock or allow access/)).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Sign in to sync" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Sync now" }));
    expect(retry).toHaveBeenCalledOnce();
  });
  it("keeps ordinary connectivity failures on download retry instead of requesting authentication", () => {
    render(
      <WorkspaceNotices
        error={null}
        onDismissError={vi.fn()}
        needsSignIn={false}
        download={paused("Cannot reach the server.")}
        syncing={false}
        onRetry={vi.fn()}
        connection={connection()}
      />,
    );
    expect(
      screen.getByRole("button", { name: "Retry download" }),
    ).toBeEnabled();
    expect(
      screen.queryByRole("button", { name: "Sign in to sync" }),
    ).not.toBeInTheDocument();
  });
  it("keeps account reconnection available from Settings and shows browser handoff feedback", () => {
    const account = { ...connection(), opened: true };
    render(<AccountReconnect connection={account} />);
    fireEvent.click(screen.getByRole("button", { name: "Reconnect account" }));
    expect(account.begin).toHaveBeenCalledOnce();
    expect(screen.getByRole("status")).toHaveTextContent(
      "Finish signing in with this account",
    );
  });
});
