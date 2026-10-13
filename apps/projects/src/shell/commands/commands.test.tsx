/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type * as DesktopBridgeModule from "@/shell/desktop/bridge";
import { DESKTOP_ACTION_EVENT } from "@/shell/desktop/bridge";
import { DesktopProvider } from "@/shell/desktop/desktop-provider";
import { Commands } from "./commands";

let mockPathname = "/acme/my-work";
let mockWorkspaceName: string | undefined = "Acme";
const mockSetQuery = jest.fn();
const mockSyncChrome = jest.fn().mockResolvedValue(undefined);
const installDesktopBridge = () => {
  window.__FORTYONE_DESKTOP__ = {
    version: 2,
    platform: "macos",
    syncChrome: mockSyncChrome,
    navigationState: jest
      .fn()
      .mockResolvedValue({ canGoBack: false, canGoForward: false }),
    navigate: jest.fn().mockResolvedValue(undefined),
    windowAction: jest.fn().mockResolvedValue(undefined),
  };
};

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: jest.fn() }),
  usePathname: () => mockPathname,
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
  useUserRole: () => ({ userRole: "admin" }),
}));
jest.mock("@/lib/hooks/workspaces", () => ({
  useCurrentWorkspace: () => ({
    workspace: mockWorkspaceName ? { name: mockWorkspaceName } : undefined,
  }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/components/shared/keyboard-shortcuts", () => ({
  KeyboardShortcuts: () => null,
}));
jest.mock("@/components/ui/invite-members", () => ({
  InviteMembersDialog: () => null,
}));
jest.mock("@/components/ui/new-objective", () => ({
  NewObjectiveDialog: () => null,
}));
jest.mock("@/components/ui/new-sprint-dialog", () => ({
  NewSprintDialog: () => null,
}));
jest.mock("@/components/ui/new-story-dialog", () => ({
  NewStoryDialog: () => null,
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
    mockSetQuery.mockClear();
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
});
