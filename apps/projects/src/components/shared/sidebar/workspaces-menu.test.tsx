/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ButtonHTMLAttributes, HTMLAttributes, ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { WorkspacesMenu } from "./workspaces-menu";

const mockFlushDrafts = jest.fn();
let mockDraftScope: {
  accountId: string;
  workspaceId: string;
  workspaceSlug: string;
} | null = null;
jest.mock("@/shell/desktop/workspace-scope", () => ({
  useDesktopWorkspaceScope: () => mockDraftScope,
}));
jest.mock("@/shell/desktop/drafts", () => ({
  retryDesktopDrafts: (...args: unknown[]) => mockFlushDrafts(...args),
}));

const mockChangeWorkspace = jest.fn();
const mockAssign = jest.fn();
const mockStop = jest.fn();
const mockBeginWorkspaceSwitch = jest.fn();
const mockWorkspaceReady = jest.fn();
const mockCancelWorkspaceSwitch = jest.fn();
const currentWorkspace = {
  id: "one",
  name: "Acme",
  slug: "acme",
  userRole: "admin",
};
const nextWorkspace = {
  id: "two",
  name: "Product",
  slug: "product",
  userRole: "member",
};
let mockCurrentWorkspace: typeof currentWorkspace | undefined =
  currentWorkspace;

jest.mock("./actions", () => ({
  changeWorkspace: (...args: unknown[]) => mockChangeWorkspace(...args),
  logOut: jest.fn(),
}));
jest.mock("./utils", () => ({ clearAllStorage: jest.fn() }));
jest.mock("next/navigation", () => ({ usePathname: () => "/acme/my-work" }));
jest.mock("@/hooks", () => ({
  useAnalytics: () => ({ analytics: { logout: jest.fn() } }),
  useLocalStorage: () => ["", jest.fn()],
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: "admin" }),
}));
jest.mock("@/lib/hooks/workspaces", () => ({
  useWorkspaces: () => ({ data: [currentWorkspace, nextWorkspace] }),
  useCurrentWorkspace: () => ({ workspace: mockCurrentWorkspace }),
}));
jest.mock("@/utils/workspace-url", () => ({
  buildWorkspaceUrl: (slug: string) => `https://${slug}.fortyone.app/my-work`,
}));
jest.mock("ui", () => {
  const Container = ({
    children,
    ...props
  }: HTMLAttributes<HTMLDivElement>) => <div {...props}>{children}</div>;
  const Button = ({
    children,
    leftIcon,
    rightIcon,
    color: _color,
    size: _size,
    variant: _variant,
    ...props
  }: ButtonHTMLAttributes<HTMLButtonElement> & {
    leftIcon?: ReactNode;
    rightIcon?: ReactNode;
    color?: string;
    size?: string;
    variant?: string;
  }) => (
    <button type="button" {...props}>
      {leftIcon}
      {children}
      {rightIcon}
    </button>
  );
  const Item = ({
    children,
    onSelect,
    ...props
  }: Omit<ButtonHTMLAttributes<HTMLButtonElement>, "onSelect"> & {
    onSelect?: () => void;
  }) => (
    <button onClick={onSelect} role="menuitem" type="button" {...props}>
      {children}
    </button>
  );
  const Menu = Object.assign(Container, {
    Button: Container,
    Items: ({ children }: { children: ReactNode }) => <div>{children}</div>,
    Group: Container,
    Item,
    SubMenu: Container,
    SubTrigger: Container,
    SubItems: Container,
    Separator: Container,
  });
  return {
    Button,
    Menu,
    Avatar: () => <span aria-hidden="true" />,
    Badge: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  };
});

const flush = async () => {
  await act(async () => {
    await Promise.resolve();
  });
};
const selectProduct = () =>
  fireEvent.click(screen.getByRole("menuitem", { name: /Product/ }));

describe("WorkspacesMenu switching", () => {
  const originalLocation = window.location;
  const originalStop = window.stop;

  beforeEach(() => {
    mockDraftScope = null;
    mockFlushDrafts.mockReset().mockResolvedValue(undefined);
    mockCurrentWorkspace = currentWorkspace;
    jest.useFakeTimers();
    mockChangeWorkspace.mockReset().mockResolvedValue("two");
    mockAssign.mockReset();
    mockStop.mockReset();
    mockBeginWorkspaceSwitch.mockReset().mockResolvedValue(undefined);
    mockWorkspaceReady.mockReset().mockResolvedValue(undefined);
    mockCancelWorkspaceSwitch.mockReset().mockResolvedValue(undefined);
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { ...originalLocation, assign: mockAssign },
    });
    window.stop = mockStop;
    window.__FORTYONE_DESKTOP__ = {
      version: 2,
      platform: "macos",
      syncChrome: jest.fn(),
      navigationState: jest.fn(),
      navigate: jest.fn(),
      windowAction: jest.fn(),
      beginWorkspaceSwitch: mockBeginWorkspaceSwitch,
      workspaceReady: mockWorkspaceReady,
      cancelWorkspaceSwitch: mockCancelWorkspaceSwitch,
    };
  });

  it("disables the current workspace item until workspace data is available", () => {
    mockCurrentWorkspace = undefined;
    render(<WorkspacesMenu />);
    expect(
      screen.getByRole("menuitem", { name: /Loading workspace/ }),
    ).toBeDisabled();
    expect(mockWorkspaceReady).not.toHaveBeenCalled();
  });

  afterEach(() => {
    jest.useRealTimers();
    Object.defineProperty(window, "location", {
      configurable: true,
      value: originalLocation,
    });
    window.stop = originalStop;
    delete window.__FORTYONE_DESKTOP__;
    jest.restoreAllMocks();
  });

  it("shows the target immediately, guards duplicate selections, and keeps progress through navigation", async () => {
    let finishPreference: () => void = () => {};
    mockChangeWorkspace.mockReturnValue(
      new Promise<void>((resolve) => {
        finishPreference = resolve;
      }),
    );
    render(<WorkspacesMenu />);
    selectProduct();
    selectProduct();
    expect(
      screen.getByRole("button", { name: "Switching to Product…" }),
    ).toHaveAttribute("aria-busy", "true");
    expect(screen.getByRole("status")).toHaveTextContent(
      "Switching to Product…",
    );
    expect(mockChangeWorkspace).toHaveBeenCalledTimes(1);
    expect(mockAssign).not.toHaveBeenCalled();
    await flush();
    expect(mockBeginWorkspaceSwitch).toHaveBeenCalledWith({
      id: "two",
      name: "Product",
      url: "https://product.fortyone.app/my-work",
    });
    await act(async () => {
      finishPreference();
    });
    expect(mockAssign).toHaveBeenCalledWith(
      "https://product.fortyone.app/my-work",
    );
    expect(screen.getByRole("status")).toHaveTextContent(
      "Switching to Product…",
    );
  });

  it("shows progress while native drafts flush and blocks navigation if storage fails", async () => {
    mockDraftScope = {
      accountId: "alice",
      workspaceId: "one",
      workspaceSlug: "acme",
    };
    let reject!: (reason: Error) => void;
    mockFlushDrafts.mockReturnValue(
      new Promise<void>((_resolve, fail) => {
        reject = fail;
      }),
    );
    render(<WorkspacesMenu />);
    selectProduct();
    expect(screen.getByRole("status")).toHaveTextContent(
      "Switching to Product…",
    );
    expect(mockChangeWorkspace).not.toHaveBeenCalled();
    expect(mockBeginWorkspaceSwitch).not.toHaveBeenCalled();
    await act(async () => {
      reject(new Error("Draft could not be saved on this device."));
    });
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Draft could not be saved",
    );
    expect(mockAssign).not.toHaveBeenCalled();
    mockFlushDrafts.mockResolvedValue(undefined);
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    await flush();
    expect(mockAssign).toHaveBeenCalledTimes(1);
  });

  it("bounds a stalled last-used preference and aborts it before navigating", async () => {
    mockChangeWorkspace.mockReturnValue(new Promise(() => {}));
    render(<WorkspacesMenu />);
    selectProduct();
    await act(async () => {
      jest.advanceTimersByTime(1_500);
    });
    expect(mockChangeWorkspace.mock.calls[0][1].aborted).toBe(true);
    expect(mockAssign).toHaveBeenCalledTimes(1);
  });

  it("continues to the destination when the last-used preference fails", async () => {
    mockChangeWorkspace.mockRejectedValue(new Error("Preference unavailable"));
    render(<WorkspacesMenu />);
    selectProduct();
    await flush();
    expect(mockAssign).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("retains loading feedback with older desktop builds and browsers", async () => {
    delete window.__FORTYONE_DESKTOP__;
    render(<WorkspacesMenu isCollapsed />);
    selectProduct();
    await flush();
    expect(
      screen.getByRole("button", { name: "Switching to Product…" }),
    ).toHaveAttribute("aria-describedby", screen.getByRole("status").id);
    expect(mockAssign).toHaveBeenCalledTimes(1);
  });

  it("bounds a stalled optional native bridge without blocking hosted navigation", async () => {
    mockBeginWorkspaceSwitch.mockReturnValue(new Promise(() => {}));
    render(<WorkspacesMenu />);
    selectProduct();
    await act(async () => {
      jest.advanceTimersByTime(1_000);
    });
    expect(mockAssign).toHaveBeenCalledTimes(1);
  });

  it("offers Retry and Return after a timeout, including a collapsed sidebar", async () => {
    render(<WorkspacesMenu isCollapsed />);
    selectProduct();
    await flush();
    await act(async () => {
      jest.advanceTimersByTime(20_000);
    });
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Product is taking longer to open.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    await flush();
    expect(mockAssign).toHaveBeenCalledTimes(2);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    await act(async () => {
      jest.advanceTimersByTime(20_000);
    });
    fireEvent.click(screen.getByRole("button", { name: "Return to Acme" }));
    expect(mockStop).toHaveBeenCalledTimes(1);
    expect(mockCancelWorkspaceSwitch).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Acme" })).not.toBeDisabled();
  });

  it("does not navigate when selecting the current workspace and reports resolved readiness", () => {
    render(<WorkspacesMenu />);
    fireEvent.click(screen.getByRole("menuitem", { name: /Acme/ }));
    expect(mockChangeWorkspace).not.toHaveBeenCalled();
    expect(mockAssign).not.toHaveBeenCalled();
    expect(mockWorkspaceReady).toHaveBeenCalledWith({ id: "one" });
  });

  it("resets stale loading state after a back-forward cache restoration", async () => {
    render(<WorkspacesMenu />);
    selectProduct();
    await flush();
    fireEvent(window, new PageTransitionEvent("pageshow", { persisted: true }));
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    selectProduct();
    await flush();
    expect(mockAssign).toHaveBeenCalledTimes(2);
  });
});
