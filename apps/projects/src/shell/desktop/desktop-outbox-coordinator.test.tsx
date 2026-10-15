import { act, render, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DesktopOutboxCoordinator } from "./desktop-outbox-coordinator";
import { getDesktopBridge } from "./bridge";
import { bindDesktopSession } from "./session-binding";
import { useDesktopWorkspaceScope } from "./workspace-scope";
import { reconcileDesktopOutbox } from "./outbox-projection";
import { DESKTOP_OUTBOX_CHANGED } from "./outbox-types";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./session-binding", () => ({ bindDesktopSession: jest.fn() }));
jest.mock("./workspace-scope", () => ({ useDesktopWorkspaceScope: jest.fn() }));
jest.mock("./desktop-provider", () => ({
  useDesktop: () => ({ isDesktop: true }),
}));
jest.mock("./outbox-projection", () => ({ reconcileDesktopOutbox: jest.fn() }));

const scope = {
  accountId: "account",
  workspaceId: "workspace",
  workspaceSlug: "engineering",
};
const empty = { operations: [], completed: [] };
const pending = {
  operations: [
    {
      id: "operation",
      kind: "story.update" as const,
      resourceId: "story",
      payload: { title: "New title" },
      status: "pending" as const,
      attempted: false,
    },
  ],
  completed: [],
};
const readOutbox = jest.fn();
const syncOutbox = jest.fn();
let online = true;
const show = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <DesktopOutboxCoordinator />
    </QueryClientProvider>,
  );

beforeEach(() => {
  jest.clearAllMocks();
  online = true;
  delete window.__FORTYONE_OFFLINE_PAGE__;
  jest.spyOn(navigator, "onLine", "get").mockImplementation(() => online);
  jest.mocked(useDesktopWorkspaceScope).mockReturnValue(scope);
  jest.mocked(bindDesktopSession).mockResolvedValue(true);
  jest.mocked(getDesktopBridge).mockReturnValue({
    readOutbox,
    syncOutbox,
    resolveOperation: jest.fn(),
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
  readOutbox.mockResolvedValue(pending);
  syncOutbox.mockResolvedValue(empty);
});
afterEach(() => {
  jest.restoreAllMocks();
  jest.useRealTimers();
});

it("loads native pending work before replaying it and reconciling acknowledgments", async () => {
  show();
  await waitFor(() => {
    expect(syncOutbox).toHaveBeenCalledTimes(1);
  });
  expect(readOutbox).toHaveBeenCalledWith({
    accountId: "account",
    workspaceId: "workspace",
  });
  expect(reconcileDesktopOutbox).toHaveBeenCalledWith(
    expect.any(QueryClient),
    scope,
    pending,
  );
  await waitFor(() => {
    expect(reconcileDesktopOutbox).toHaveBeenCalledWith(
      expect.any(QueryClient),
      scope,
      empty,
    );
  });
});

it("restores offline work without replay and resumes after reconnect", async () => {
  online = false;
  show();
  await waitFor(() => {
    expect(reconcileDesktopOutbox).toHaveBeenCalled();
  });
  expect(syncOutbox).not.toHaveBeenCalled();
  online = true;
  act(() => {
    window.dispatchEvent(new Event("online"));
  });
  await waitFor(() => {
    expect(syncOutbox).toHaveBeenCalledTimes(1);
  });
});

it("drops a late read after the workspace owner unmounts", async () => {
  let finish!: (value: typeof pending) => void;
  readOutbox.mockReturnValue(
    new Promise((resolve) => {
      finish = resolve;
    }),
  );
  const view = show();
  await waitFor(() => {
    expect(readOutbox).toHaveBeenCalled();
  });
  view.unmount();
  await act(async () => {
    finish(pending);
  });
  expect(syncOutbox).not.toHaveBeenCalled();
  expect(reconcileDesktopOutbox).not.toHaveBeenCalled();
});

it("ignores enqueue events for a different account", async () => {
  readOutbox.mockResolvedValue(empty);
  show();
  await waitFor(() => {
    expect(reconcileDesktopOutbox).toHaveBeenCalled();
  });
  act(() => {
    window.dispatchEvent(
      new CustomEvent(DESKTOP_OUTBOX_CHANGED, {
        detail: { ...scope, accountId: "another-account" },
      }),
    );
  });
  expect(readOutbox).toHaveBeenCalledTimes(1);
});

it("coalesces overlapping replay triggers and retries pending work periodically", async () => {
  jest.useFakeTimers();
  let finish!: (value: typeof pending) => void;
  syncOutbox
    .mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    )
    .mockResolvedValue(pending);
  const view = show();
  await act(async () => {
    await Promise.resolve();
  });
  expect(syncOutbox).toHaveBeenCalledTimes(1);
  act(() => {
    window.dispatchEvent(new Event("online"));
    window.dispatchEvent(new Event("online"));
  });
  expect(syncOutbox).toHaveBeenCalledTimes(1);
  await act(async () => {
    finish(pending);
  });
  expect(syncOutbox).toHaveBeenCalledTimes(2);
  await act(async () => {
    jest.advanceTimersByTime(30_000);
  });
  expect(syncOutbox).toHaveBeenCalledTimes(3);
  view.unmount();
  jest.useRealTimers();
});
