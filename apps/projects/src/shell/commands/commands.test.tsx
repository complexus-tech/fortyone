/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { toast } from "sonner";
import { useHotkeys } from "react-hotkeys-hook";
import type * as DesktopBridgeModule from "@/shell/desktop/bridge";
import { DESKTOP_ACTION_EVENT } from "@/shell/desktop/bridge";
import { DesktopProvider } from "@/shell/desktop/desktop-provider";
import { Commands } from "./commands";

let mockPathname = "/acme/my-work";
let mockWorkspaceName: string | undefined = "Acme";
let mockWorkspacePresent = true;
let mockRole: string | undefined = "admin";
let mockObjectivesEnabled = true;
let mockFeaturesPending = false;
let mockHasSidebar = true;
let mockSidebarCollapsed = false;
let mockTourSteps = [{}];
let mockParams: Record<string, string> = {};
const mockPush = jest.fn();
const mockLogout = jest.fn().mockResolvedValue(undefined);
const mockAnalyticsLogout = jest.fn();
const mockClearStorage = jest.fn();
const mockToggleSidebar = jest.fn();
const mockStartTour = jest.fn();
const mockSyncMenu = jest.fn().mockResolvedValue(undefined);
const mockNavigateOffline = jest.fn((_href: string, fallback: () => void) => {
  fallback();
  return Promise.resolve();
});
const mockSetQuery = jest.fn();
const mockSyncChrome = jest.fn().mockResolvedValue(undefined);
const installDesktopBridge = (withMenus = false) => {
  window.__FORTYONE_DESKTOP__ = {
    version: 2,
    platform: "macos",
    syncChrome: mockSyncChrome,
    navigationState: jest
      .fn()
      .mockResolvedValue({ canGoBack: false, canGoForward: false }),
    navigate: jest.fn().mockResolvedValue(undefined),
    windowAction: jest.fn().mockResolvedValue(undefined),
    ...(withMenus ? { syncMenu: mockSyncMenu } : {}),
  };
};

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
  usePathname: () => mockPathname,
  useParams: () => mockParams,
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("@/hooks/features", () => ({
  useFeatures: () => ({
    objectiveEnabled: mockObjectivesEnabled,
    isPending: mockFeaturesPending,
  }),
}));
jest.mock("@/hooks/tracking", () => ({
  useAnalytics: () => ({ analytics: { logout: mockAnalyticsLogout } }),
}));
jest.mock("@/components/shared/sidebar/actions", () => ({
  logOut: () => mockLogout(),
}));
jest.mock("@/components/shared/sidebar/utils", () => ({
  clearAllStorage: () => mockClearStorage(),
}));
jest.mock("@/components/shared/sidebar/sidebar-context", () => ({
  useOptionalSidebar: () =>
    mockHasSidebar
      ? { isCollapsed: mockSidebarCollapsed, toggleSidebar: mockToggleSidebar }
      : null,
}));
jest.mock("@/components/walkthrough/walkthrough-provider", () => ({
  useWalkthrough: () => ({
    steps: mockTourSteps,
    startWalkthrough: mockStartTour,
  }),
}));
jest.mock("@/shell/desktop/offline-navigation", () => ({
  navigateWithDesktopOffline: (href: string, fallback: () => void) =>
    mockNavigateOffline(href, fallback),
}));
jest.mock("next-themes", () => ({
  useTheme: () => ({ resolvedTheme: "light" }),
}));
jest.mock("react-hotkeys-hook", () => ({ useHotkeys: jest.fn() }));
jest.mock("nuqs", () => ({ useQueryState: () => ["existing", mockSetQuery] }));
jest.mock("ui", () => ({
  Kbd: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  Tooltip: ({ children }: { children: ReactNode }) => <>{children}</>,
}));
jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: (term: string) =>
      term === "storyTerm" ? "tasks" : "objectives",
  }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockRole }),
}));
jest.mock("@/lib/hooks/workspaces", () => ({
  useCurrentWorkspace: () => ({
    workspace: mockWorkspacePresent
      ? { id: "workspace-1", name: mockWorkspaceName }
      : undefined,
  }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/components/shared/keyboard-shortcuts", () => ({
  KeyboardShortcuts: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div>Keyboard shortcuts dialog</div> : null,
}));
jest.mock("@/components/ui/invite-members", () => ({
  InviteMembersDialog: () => null,
}));
jest.mock("@/components/ui/new-objective", () => ({
  NewObjectiveDialog: ({
    isOpen,
    teamId,
  }: {
    isOpen: boolean;
    teamId?: string;
  }) => (isOpen ? <div data-team-id={teamId}>New objective dialog</div> : null),
}));
jest.mock("@/components/ui/new-sprint-dialog", () => ({
  NewSprintDialog: () => null,
}));
jest.mock("@/components/ui/new-story-dialog", () => ({
  NewStoryDialog: ({
    isOpen,
    teamId,
    objectiveId,
    sprintId,
  }: {
    isOpen: boolean;
    teamId?: string;
    objectiveId?: string;
    sprintId?: string;
  }) =>
    isOpen ? (
      <div
        data-objective-id={objectiveId}
        data-sprint-id={sprintId}
        data-team-id={teamId}
      >
        New story dialog
      </div>
    ) : null,
}));
jest.mock("./command-bar", () => ({
  CommandBar: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div>Command palette</div> : null,
}));
jest.mock("@/shell/desktop/bridge", () => ({
  ...jest.requireActual<typeof DesktopBridgeModule>("@/shell/desktop/bridge"),
  readDesktopBackground: () => [251, 250, 248],
}));

describe("Commands desktop actions", () => {
  beforeEach(() => {
    mockPathname = "/acme/my-work";
    mockWorkspaceName = "Acme";
    mockWorkspacePresent = true;
    mockRole = "admin";
    mockObjectivesEnabled = true;
    mockFeaturesPending = false;
    mockHasSidebar = true;
    mockSidebarCollapsed = false;
    mockTourSteps = [{}];
    mockParams = {};
    jest.clearAllMocks();
    mockLogout.mockResolvedValue(undefined);
    localStorage.clear();
    sessionStorage.clear();
    window.history.replaceState(null, "", "/acme/my-work");
    delete window.__FORTYONE_DESKTOP__;
  });
  afterEach(() => {
    delete window.__FORTYONE_DESKTOP__;
  });

  it("retains the normal browser search trigger", () => {
    render(<Commands showTrigger />);
    fireEvent.click(
      screen.getByRole("button", {
        name: "Search tasks, objectives, or commands",
      }),
    );
    expect(screen.getByText("Command palette")).toBeInTheDocument();
  });

  it("opens the existing palette from native search and preserves the search page input", async () => {
    installDesktopBridge();
    const { rerender } = render(
      <DesktopProvider>
        <Commands showTrigger />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(
        screen.getByRole("button", {
          name: "Search tasks, objectives, or commands",
        }),
      ).toHaveClass("h-[30px]");
    });
    expect(screen.getByText("Search Acme")).toBeInTheDocument();
    const searchButton = screen.getByRole("button", {
      name: "Search tasks, objectives, or commands",
    });
    expect(searchButton.querySelector("svg")).toHaveClass(
      "h-[16px]",
      "w-[16px]",
    );
    expect(searchButton).toHaveTextContent("⌘ K");
    fireEvent(
      window,
      new CustomEvent(DESKTOP_ACTION_EVENT, { detail: { action: "search" } }),
    );
    expect(screen.getByText("Command palette")).toBeInTheDocument();

    mockPathname = "/acme/search";
    rerender(
      <DesktopProvider>
        <Commands showTrigger />
      </DesktopProvider>,
    );
    const input = screen.getByRole("searchbox", {
      name: "Search tasks and objectives",
    });
    expect(input).toHaveValue("existing");
    expect(input.closest("search")).toHaveClass("h-[30px]");
    expect(
      screen.getByRole("button", { name: "Open command menu" }),
    ).toHaveClass("h-[30px]");
    fireEvent.change(input, { target: { value: "new query" } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(mockSetQuery).toHaveBeenCalledWith("new query");
  });

  it("toggles the palette on repeated native search shortcuts while the search button always opens it", async () => {
    installDesktopBridge();
    mockWorkspaceName = undefined;
    render(
      <DesktopProvider>
        <Commands showTrigger />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(document.documentElement.dataset.desktopPlatform).toBe("macos");
    });
    expect(screen.getByText("Search workspace")).toBeInTheDocument();

    const dispatchNativeSearch = () => {
      fireEvent(
        window,
        new CustomEvent(DESKTOP_ACTION_EVENT, { detail: { action: "search" } }),
      );
    };
    dispatchNativeSearch();
    expect(screen.getByText("Command palette")).toBeInTheDocument();
    dispatchNativeSearch();
    expect(screen.queryByText("Command palette")).not.toBeInTheDocument();

    const searchButton = screen.getByRole("button", {
      name: "Search tasks, objectives, or commands",
    });
    fireEvent.click(searchButton);
    fireEvent.click(searchButton);
    expect(screen.getByText("Command palette")).toBeInTheDocument();
    dispatchNativeSearch();
    expect(screen.queryByText("Command palette")).not.toBeInTheDocument();
  });

  const dispatchAction = (action: string) => {
    fireEvent(
      window,
      new CustomEvent(DESKTOP_ACTION_EVENT, { detail: { action } }),
    );
  };
  const renderDesktop = async (withMenus = true) => {
    installDesktopBridge(withMenus);
    const view = render(
      <DesktopProvider>
        <Commands />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(document.documentElement.dataset.desktopPlatform).toBe("macos");
    });
    return view;
  };

  it("shares app creation callbacks instead of opening a second dialog", async () => {
    installDesktopBridge(true);
    const createStory = jest.fn();
    const createObjective = jest.fn();
    const showShortcuts = jest.fn();
    render(
      <DesktopProvider>
        <Commands
          onCreateObjective={createObjective}
          onCreateStory={createStory}
          onShowKeyboardShortcuts={showShortcuts}
        />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(mockSyncMenu).toHaveBeenLastCalledWith(
        expect.objectContaining({
          ready: true,
          canCreateStory: true,
          canCreateObjective: true,
        }),
      );
    });
    dispatchAction("create-story");
    dispatchAction("create-objective");
    dispatchAction("keyboard-shortcuts");
    expect(createStory).toHaveBeenCalledTimes(1);
    expect(createObjective).toHaveBeenCalledTimes(1);
    expect(showShortcuts).toHaveBeenCalledTimes(1);
    expect(screen.queryByText("New story dialog")).not.toBeInTheDocument();
    expect(screen.queryByText("New objective dialog")).not.toBeInTheDocument();
    expect(
      screen.queryByText("Keyboard shortcuts dialog"),
    ).not.toBeInTheDocument();
  });

  it("opens contextual dialogs and keyboard help when no app callbacks are mounted", async () => {
    mockHasSidebar = false;
    mockParams = {
      teamId: "team-1",
      objectiveId: "objective-1",
      sprintId: "sprint-1",
    };
    await renderDesktop();
    dispatchAction("create-story");
    const story = screen.getByText("New story dialog");
    expect(story).toHaveAttribute("data-team-id", "team-1");
    expect(story).toHaveAttribute("data-objective-id", "objective-1");
    expect(story).toHaveAttribute("data-sprint-id", "sprint-1");
    dispatchAction("create-objective");
    expect(screen.getByText("New objective dialog")).toHaveAttribute(
      "data-team-id",
      "team-1",
    );
    dispatchAction("keyboard-shortcuts");
    expect(screen.getByText("Keyboard shortcuts dialog")).toBeInTheDocument();
    expect(mockSyncMenu).toHaveBeenLastCalledWith(
      expect.objectContaining({
        canToggleSidebar: false,
      }),
    );
  });

  it.each(["guest", undefined])(
    "blocks creation for role %s in native state and event handlers",
    async (role) => {
      mockRole = role;
      await renderDesktop();
      dispatchAction("create-story");
      dispatchAction("create-objective");
      expect(screen.queryByText("New story dialog")).not.toBeInTheDocument();
      expect(
        screen.queryByText("New objective dialog"),
      ).not.toBeInTheDocument();
      expect(mockSyncMenu).toHaveBeenLastCalledWith(
        expect.objectContaining({
          canCreateStory: false,
          canCreateObjective: false,
        }),
      );
    },
  );

  it.each(["disabled", "pending"])(
    "blocks objective creation when the feature is %s",
    async (state) => {
      mockObjectivesEnabled = state !== "disabled";
      mockFeaturesPending = state === "pending";
      await renderDesktop();
      dispatchAction("create-objective");
      expect(
        screen.queryByText("New objective dialog"),
      ).not.toBeInTheDocument();
      expect(mockSyncMenu).toHaveBeenLastCalledWith(
        expect.objectContaining({ canCreateObjective: false }),
      );
    },
  );

  it("ignores native commands in a browser or without an active workspace", async () => {
    const browser = render(<Commands />);
    dispatchAction("create-story");
    dispatchAction("settings");
    expect(screen.queryByText("New story dialog")).not.toBeInTheDocument();
    expect(mockPush).not.toHaveBeenCalled();
    browser.unmount();
    mockWorkspacePresent = false;
    await renderDesktop();
    dispatchAction("create-story");
    dispatchAction("search");
    dispatchAction("settings");
    dispatchAction("logout");
    expect(mockSyncMenu).toHaveBeenLastCalledWith(
      expect.objectContaining({ ready: false }),
    );
    expect(screen.queryByText("New story dialog")).not.toBeInTheDocument();
    expect(screen.queryByText("Command palette")).not.toBeInTheDocument();
    expect(mockPush).not.toHaveBeenCalled();
    expect(mockLogout).not.toHaveBeenCalled();
  });

  it("remembers the current page for settings and preserves it when already in settings", async () => {
    window.history.replaceState(null, "", "/acme/my-work?view=assigned");
    const { rerender } = await renderDesktop();
    dispatchAction("settings");
    await waitFor(() => {
      expect(JSON.parse(localStorage.getItem("pathBeforeSettings")!)).toBe(
        "/acme/my-work?view=assigned",
      );
    });
    expect(mockNavigateOffline).toHaveBeenCalledWith(
      "/acme/settings/account/preferences",
      expect.any(Function),
    );
    expect(mockPush).toHaveBeenCalledWith("/acme/settings/account/preferences");
    mockPathname = "/acme/settings/account/notifications";
    rerender(
      <DesktopProvider>
        <Commands />
      </DesktopProvider>,
    );
    dispatchAction("settings");
    expect(JSON.parse(localStorage.getItem("pathBeforeSettings")!)).toBe(
      "/acme/my-work?view=assigned",
    );
  });

  it("toggles the existing sidebar and publishes its current state", async () => {
    const { rerender } = await renderDesktop();
    dispatchAction("toggle-sidebar");
    expect(mockToggleSidebar).toHaveBeenCalledTimes(1);
    mockSidebarCollapsed = true;
    rerender(
      <DesktopProvider>
        <Commands />
      </DesktopProvider>,
    );
    await waitFor(() => {
      expect(mockSyncMenu).toHaveBeenLastCalledWith(
        expect.objectContaining({ sidebarCollapsed: true }),
      );
    });
    mockHasSidebar = false;
    rerender(
      <DesktopProvider>
        <Commands />
      </DesktopProvider>,
    );
    dispatchAction("toggle-sidebar");
    expect(mockToggleSidebar).toHaveBeenCalledTimes(1);
  });

  it.each([false, true])(
    "keeps shortcut ownership compatible with native menus=%s",
    async (withMenus) => {
      await renderDesktop(withMenus);
      const hotkeys = jest.mocked(useHotkeys).mock.calls;
      const commandK = hotkeys.filter(([keys]) => keys === "mod+k").at(-1);
      const shortcuts = hotkeys.filter(([keys]) => keys === "mod+/").at(-1);
      expect(commandK?.[2]).toEqual({ enabled: !withMenus });
      expect(shortcuts?.[2]).toEqual({ enabled: !withMenus });
    },
  );

  it("starts the existing product tour from the app", async () => {
    await renderDesktop();
    dispatchAction("product-tour");
    expect(mockStartTour).toHaveBeenCalledTimes(1);
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("carries a tour request from settings until the app layout has initialized its steps", async () => {
    mockHasSidebar = false;
    mockTourSteps = [];
    const { rerender } = await renderDesktop();
    dispatchAction("product-tour");
    expect(mockPush).toHaveBeenCalledWith("/acme/my-work");
    expect(mockStartTour).not.toHaveBeenCalled();
    expect(
      sessionStorage.getItem("fortyone:desktop-product-tour:workspace-1"),
    ).toBe("pending");
    mockHasSidebar = true;
    rerender(
      <DesktopProvider>
        <Commands />
      </DesktopProvider>,
    );
    expect(mockStartTour).not.toHaveBeenCalled();
    mockTourSteps = [{}];
    rerender(
      <DesktopProvider>
        <Commands />
      </DesktopProvider>,
    );
    expect(mockStartTour).toHaveBeenCalledTimes(1);
    expect(
      sessionStorage.getItem("fortyone:desktop-product-tour:workspace-1"),
    ).toBeNull();
    rerender(
      <DesktopProvider>
        <Commands />
      </DesktopProvider>,
    );
    expect(mockStartTour).toHaveBeenCalledTimes(1);
  });

  it("waits for current editor snapshots, rejects repeated logout, and retains storage when draft flushing fails", async () => {
    let resolveSnapshot: () => void = () => {};
    const snapshot = new Promise<void>((resolve) => {
      resolveSnapshot = resolve;
    });
    const listener = (event: Event) => {
      (
        event as CustomEvent<{ waitUntil: (write: Promise<void>) => void }>
      ).detail.waitUntil(snapshot);
    };
    window.addEventListener("fortyone:desktop-before-close", listener);
    mockLogout.mockRejectedValue(new Error("Local draft could not be saved"));
    try {
      await renderDesktop();
      const wasInert = document.body.inert;
      dispatchAction("logout");
      dispatchAction("logout");
      dispatchAction("create-story");
      dispatchAction("settings");
      expect(document.body.inert).toBe(true);
      expect(screen.queryByText("New story dialog")).not.toBeInTheDocument();
      expect(mockPush).not.toHaveBeenCalled();
      expect(mockLogout).not.toHaveBeenCalled();
      await act(async () => {
        resolveSnapshot();
      });
      await waitFor(() => {
        expect(toast.error).toHaveBeenCalledWith(
          "Local draft could not be saved",
        );
      });
      expect(mockLogout).toHaveBeenCalledTimes(1);
      expect(mockAnalyticsLogout).not.toHaveBeenCalled();
      expect(mockClearStorage).not.toHaveBeenCalled();
      expect(window.location.pathname).toBe("/acme/my-work");
      expect(document.body.inert).toBe(wasInert);
    } finally {
      window.removeEventListener("fortyone:desktop-before-close", listener);
    }
  });

  it("blocks logout entirely when an active editor cannot save its latest snapshot", async () => {
    const listener = (event: Event) => {
      (
        event as CustomEvent<{ waitUntil: (write: Promise<void>) => void }>
      ).detail.waitUntil(Promise.reject(new Error("Editor has unsaved work")));
    };
    window.addEventListener("fortyone:desktop-before-close", listener);
    try {
      await renderDesktop();
      dispatchAction("logout");
      await waitFor(() => {
        expect(toast.error).toHaveBeenCalledWith("Editor has unsaved work");
      });
      expect(mockLogout).not.toHaveBeenCalled();
      expect(mockClearStorage).not.toHaveBeenCalled();
      expect(mockAnalyticsLogout).not.toHaveBeenCalled();
    } finally {
      window.removeEventListener("fortyone:desktop-before-close", listener);
    }
  });
});
