/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import type * as DesktopBridgeModule from "./bridge";
import {
  DESKTOP_ACTION_EVENT,
  DESKTOP_HISTORY_EVENT,
  EMPTY_DESKTOP_SHELL,
} from "./bridge";
import {
  DesktopProvider,
  useDesktop,
  useDesktopAction,
  useDesktopShell,
} from "./desktop-provider";

const mockSyncChrome = jest.fn().mockResolvedValue(undefined);
const mockNavigationState = jest.fn();
const mockNavigate = jest.fn().mockResolvedValue(undefined);
const mockWindowAction = jest.fn().mockResolvedValue(undefined);
const mockAction = jest.fn();
let mockTheme = "light";
let mockThemePreference = "system";

type DesktopShellState = DesktopBridgeModule.DesktopShellState;

jest.mock("next-themes", () => ({
  useTheme: () => ({ resolvedTheme: mockTheme, theme: mockThemePreference }),
}));

jest.mock("./bridge", () => ({
  ...jest.requireActual<typeof DesktopBridgeModule>("./bridge"),
  readDesktopBackground: () => [251, 250, 248],
}));

const Shell = ({ state }: { state: DesktopShellState }) => {
  const isDesktop = useDesktopShell(state);
  useDesktopAction("search", mockAction);
  return <span>{isDesktop ? "Native toolbar" : "Web toolbar"}</span>;
};

const Status = () => {
  const { isDesktop } = useDesktop();
  return <span>{isDesktop ? "Connected" : "Disconnected"}</span>;
};

const SHELL_STATE: DesktopShellState = {
  ready: true,
  canCreateStory: true,
  storyLabel: "task",
};

const installBridge = () => {
  window.__FORTYONE_DESKTOP__ = {
    version: 2,
    platform: "macos",
    syncChrome: mockSyncChrome,
    navigationState: mockNavigationState,
    navigate: mockNavigate,
    windowAction: mockWindowAction,
  };
};

describe("DesktopProvider", () => {
  beforeEach(() => {
    mockSyncChrome.mockReset().mockResolvedValue(undefined);
    mockNavigationState.mockReset().mockResolvedValue({
      canGoBack: false,
      canGoForward: false,
    });
    mockNavigate.mockReset().mockResolvedValue(undefined);
    mockWindowAction.mockReset().mockResolvedValue(undefined);
    mockAction.mockClear();
    mockTheme = "light";
    mockThemePreference = "system";
    delete window.__FORTYONE_DESKTOP__;
  });

  afterEach(() => {
    delete window.__FORTYONE_DESKTOP__;
    jest.restoreAllMocks();
  });

  it("keeps ordinary browser controls and ignores desktop events", () => {
    render(
      <DesktopProvider>
        <Shell state={SHELL_STATE} />
      </DesktopProvider>,
    );
    fireEvent(
      window,
      new CustomEvent(DESKTOP_ACTION_EVENT, {
        detail: { action: "search" },
      }),
    );
    expect(screen.getByText("Web toolbar")).toBeInTheDocument();
    expect(mockSyncChrome).not.toHaveBeenCalled();
    expect(mockAction).not.toHaveBeenCalled();
    expect(document.documentElement.dataset.desktopPlatform).toBeUndefined();
    expect(document.documentElement.dataset.desktopVibrancy).toBeUndefined();
  });

  it("enables vibrancy only after a capable native build connects and removes it on unmount", async () => {
    installBridge();
    window.__FORTYONE_DESKTOP__!.vibrancy = true;
    let finishHandshake: () => void = () => {};
    mockSyncChrome.mockReturnValue(
      new Promise<void>((resolve) => {
        finishHandshake = resolve;
      }),
    );
    const { unmount } = render(
      <DesktopProvider>
        <Status />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(mockSyncChrome).toHaveBeenCalled();
    });
    expect(document.documentElement.dataset.desktopVibrancy).toBeUndefined();

    await act(async () => {
      finishHandshake();
    });
    expect(document.documentElement.dataset.desktopVibrancy).toBe("true");
    unmount();
    expect(document.documentElement.dataset.desktopVibrancy).toBeUndefined();
    expect(document.documentElement.dataset.desktopPlatform).toBeUndefined();
  });

  it("publishes shell capabilities, synchronizes theme, and clears actions on unmount", async () => {
    window.__FORTYONE_DESKTOP__ = {
      version: 2,
      platform: "macos",
      syncChrome: mockSyncChrome,
      navigationState: mockNavigationState,
      navigate: mockNavigate,
      windowAction: mockWindowAction,
    };
    const { rerender } = render(
      <DesktopProvider>
        <Shell state={SHELL_STATE} />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(mockSyncChrome).toHaveBeenLastCalledWith({
        ...SHELL_STATE,
        background: [251, 250, 248],
        dark: false,
        systemTheme: true,
      });
    });
    expect(
      screen.queryByRole("banner", { name: "Desktop title bar" }),
    ).toBeNull();
    fireEvent(
      window,
      new CustomEvent(DESKTOP_ACTION_EVENT, {
        detail: { action: "search" },
      }),
    );
    expect(mockAction).toHaveBeenCalledTimes(1);
    fireEvent(window, new Event(DESKTOP_ACTION_EVENT));
    fireEvent(
      window,
      new CustomEvent(DESKTOP_ACTION_EVENT, {
        detail: { action: "run-command" },
      }),
    );
    expect(mockAction).toHaveBeenCalledTimes(1);

    mockTheme = "dark";
    mockThemePreference = "dark";
    rerender(
      <DesktopProvider>
        <Status />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(mockSyncChrome).toHaveBeenLastCalledWith({
        ...EMPTY_DESKTOP_SHELL,
        background: [251, 250, 248],
        dark: true,
        systemTheme: false,
      });
    });
    fireEvent(
      window,
      new CustomEvent(DESKTOP_ACTION_EVENT, {
        detail: { action: "search" },
      }),
    );
    expect(mockAction).toHaveBeenCalledTimes(1);
  });

  it("shows one fallback title bar and uses actual native back-forward availability", async () => {
    installBridge();
    render(
      <DesktopProvider>
        <Status />
      </DesktopProvider>,
    );
    const back = await screen.findByRole("button", { name: "Back" });
    const forward = screen.getByRole("button", { name: "Forward" });
    expect(back).toBeDisabled();
    expect(forward).toBeDisabled();
    expect(document.documentElement.dataset.desktopFallback).toBe("true");
    expect(document.documentElement.dataset.desktopVibrancy).toBeUndefined();

    mockNavigationState.mockResolvedValue({
      canGoBack: true,
      canGoForward: false,
    });
    fireEvent(window, new Event(DESKTOP_HISTORY_EVENT));
    await waitFor(() => {
      expect(back).toBeEnabled();
    });

    mockNavigationState.mockResolvedValue({
      canGoBack: false,
      canGoForward: true,
    });
    fireEvent.click(back);
    expect(mockNavigate).toHaveBeenCalledWith("back");
    await waitFor(() => {
      expect(forward).toBeEnabled();
    });
    expect(back).toBeDisabled();
    fireEvent.click(forward);
    expect(mockNavigate).toHaveBeenCalledWith("forward");
  });

  it("republishes the latest native state when a cached page is restored", async () => {
    installBridge();
    const { rerender } = render(
      <DesktopProvider>
        <Shell state={SHELL_STATE} />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(mockSyncChrome).toHaveBeenLastCalledWith(
        expect.objectContaining(SHELL_STATE),
      );
    });

    mockTheme = "dark";
    mockThemePreference = "dark";
    const updatedState = {
      ...SHELL_STATE,
      canCreateStory: false,
      storyLabel: "issue",
    };
    rerender(
      <DesktopProvider>
        <Shell state={updatedState} />
      </DesktopProvider>,
    );
    const expected = {
      ...updatedState,
      background: [251, 250, 248],
      dark: true,
      systemTheme: false,
    };
    await waitFor(() => {
      expect(mockSyncChrome).toHaveBeenLastCalledWith(expected);
    });

    mockSyncChrome.mockClear();
    fireEvent(window, new PageTransitionEvent("pageshow", { persisted: true }));
    await waitFor(() => {
      expect(mockSyncChrome).toHaveBeenCalledTimes(1);
    });
    expect(mockSyncChrome).toHaveBeenLastCalledWith(expected);
  });

  it("ignores stale history responses after a newer navigation", async () => {
    installBridge();
    let resolvePrevious: (
      value: DesktopBridgeModule.DesktopNavigationState,
    ) => void = () => {};
    mockNavigationState.mockReturnValueOnce(
      new Promise((resolve) => {
        resolvePrevious = resolve;
      }),
    );
    render(
      <DesktopProvider>
        <Status />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(mockNavigationState).toHaveBeenCalledTimes(1);
    });

    mockNavigationState.mockResolvedValue({
      canGoBack: true,
      canGoForward: false,
    });
    fireEvent(window, new PopStateEvent("popstate"));
    const back = screen.getByRole("button", { name: "Back" });
    await waitFor(() => {
      expect(back).toBeEnabled();
    });
    await act(async () => {
      resolvePrevious({ canGoBack: false, canGoForward: false });
    });
    expect(back).toBeEnabled();
  });

  it("drags or zooms blank chrome without intercepting interactive controls", async () => {
    installBridge();
    mockNavigationState.mockResolvedValue({
      canGoBack: true,
      canGoForward: false,
    });
    render(
      <DesktopProvider>
        <Status />
      </DesktopProvider>,
    );
    const header = await screen.findByRole("banner", {
      name: "Desktop title bar",
    });
    fireEvent.mouseDown(header, { button: 0, detail: 1 });
    expect(mockWindowAction).toHaveBeenLastCalledWith("drag");
    fireEvent.doubleClick(header, { button: 0, detail: 2 });
    expect(mockWindowAction).toHaveBeenLastCalledWith("toggle-maximize");

    fireEvent.mouseDown(header, { button: 2, detail: 1 });
    fireEvent.mouseDown(header, { button: 0, detail: 2 });
    fireEvent.mouseDown(screen.getByRole("button", { name: "Back" }), {
      button: 0,
      detail: 1,
    });
    fireEvent.doubleClick(screen.getByRole("button", { name: "Back" }), {
      button: 0,
      detail: 2,
    });
    expect(mockWindowAction).toHaveBeenCalledTimes(2);
  });

  it("retains web controls when the native handshake fails", async () => {
    const error = new Error("Desktop capability unavailable");
    mockSyncChrome.mockRejectedValue(error);
    const logError = jest.spyOn(console, "error").mockImplementation(() => {});
    window.__FORTYONE_DESKTOP__ = {
      version: 2,
      platform: "macos",
      syncChrome: mockSyncChrome,
      navigationState: mockNavigationState,
      navigate: mockNavigate,
      windowAction: mockWindowAction,
    };
    render(
      <DesktopProvider>
        <Shell state={SHELL_STATE} />
      </DesktopProvider>,
    );

    await waitFor(() => {
      expect(logError).toHaveBeenCalledWith(
        "Could not synchronize the desktop toolbar",
        error,
      );
    });
    expect(screen.getByText("Web toolbar")).toBeInTheDocument();
    fireEvent(
      window,
      new CustomEvent(DESKTOP_ACTION_EVENT, {
        detail: { action: "search" },
      }),
    );
    expect(mockAction).not.toHaveBeenCalled();
  });
});
