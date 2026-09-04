/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { toast } from "sonner";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { AppCommandBar } from "./app-command-bar";

const mockCompleteWalkthroughAction = jest.fn();
const mockStartWalkthrough = jest.fn();
const mockCloseWalkthrough = jest.fn();
const mockRouterPush = jest.fn();
let mockUserRole: "admin" | "member" | "guest" | undefined = "admin";
let mockMembersPending = false;
let mockTeamsPending = false;
let mockHasTeams = true;
let mockWalkthroughActive = false;

jest.mock("sonner", () => ({ toast: jest.fn() }));

jest.mock("next/navigation", () => ({
  useParams: () => ({}),
  usePathname: () => "/acme/my-work",
  useRouter: () => ({ push: mockRouterPush }),
  useSearchParams: () => new URLSearchParams(window.location.search),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user-1" } } }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({
    data: [{ id: "user-1", isActive: true, isSystem: false }],
    isPending: mockMembersPending,
    isError: false,
  }),
}));
jest.mock("@/modules/teams/public/queries", () => ({
  useJoinedTeams: () => ({
    data: mockHasTeams ? [{ id: "team-1" }] : [],
    isPending: mockTeamsPending,
    isError: false,
  }),
}));

jest.mock("ui", () => {
  const MockButton = ({
    children,
    disabled,
    href,
    leftIcon,
    onClick,
    "aria-label": ariaLabel,
    "data-walkthrough-create-kind": walkthroughCreateKind,
    "data-walkthrough-target": walkthroughTarget,
  }: {
    children?: ReactNode;
    disabled?: boolean;
    href?: string;
    leftIcon?: ReactNode;
    onClick?: () => void;
    "aria-label"?: string;
    "data-walkthrough-create-kind"?: string;
    "data-walkthrough-target"?: string;
  }) =>
    href ? (
      <a
        aria-label={ariaLabel}
        data-walkthrough-create-kind={walkthroughCreateKind}
        data-walkthrough-target={walkthroughTarget}
        href={href}
      >
        {leftIcon}
        {children}
      </a>
    ) : (
      <button
        aria-label={ariaLabel}
        data-walkthrough-create-kind={walkthroughCreateKind}
        data-walkthrough-target={walkthroughTarget}
        disabled={disabled}
        onClick={onClick}
        type="button"
      >
        {leftIcon}
        {children}
      </button>
    );
  const MockContainer = ({
    children,
    "data-walkthrough-target": walkthroughTarget,
  }: {
    children: ReactNode;
    "data-walkthrough-target"?: string;
  }) => <div data-walkthrough-target={walkthroughTarget}>{children}</div>;
  const MockMenuItem = ({
    children,
    onSelect,
  }: {
    children: ReactNode;
    onSelect?: () => void;
  }) => (
    <button onClick={onSelect} type="button">
      {children}
    </button>
  );

  const MockMenu = Object.assign(MockContainer, {
    Button: MockContainer,
    Group: MockContainer,
    Item: MockMenuItem,
    Items: MockContainer,
  });

  return {
    Badge: MockContainer,
    Box: MockContainer,
    Button: MockButton,
    Flex: MockContainer,
    Menu: MockMenu,
    Tooltip: MockContainer,
  };
});

jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "story",
  }),
}));

jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockUserRole }),
}));

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
    workspaceSlug: "acme",
  }),
}));

jest.mock("@/modules/notifications/hooks/unread", () => ({
  useUnreadNotifications: () => ({ data: 3 }),
}));

jest.mock("@/components/ui/new-objective", () => ({
  NewObjectiveDialog: () => null,
}));

jest.mock("@/components/ui/new-story-dialog", () => ({
  NewStoryDialog: ({
    isOpen,
    onCreated,
    setIsOpen,
  }: {
    isOpen: boolean;
    onCreated: () => void;
    setIsOpen: (isOpen: boolean) => void;
  }) =>
    isOpen ? (
      <div>
        <button onClick={onCreated} type="button">
          Simulate story created
        </button>
        <button
          onClick={() => {
            setIsOpen(false);
            onCreated();
          }}
          type="button"
        >
          Simulate story saved
        </button>
        <button
          onClick={() => {
            setIsOpen(false);
          }}
          type="button"
        >
          Close task dialog
        </button>
      </div>
    ) : null,
}));

jest.mock("@/components/walkthrough/walkthrough-provider", () => ({
  useWalkthrough: () => ({
    completeWalkthroughAction: mockCompleteWalkthroughAction,
    closeWalkthrough: mockCloseWalkthrough,
    startWalkthrough: mockStartWalkthrough,
    state: { isActive: mockWalkthroughActive },
  }),
}));

jest.mock("@/components/shared/keyboard-shortcuts", () => ({
  KeyboardShortcuts: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div>Keyboard shortcuts dialog</div> : null,
}));

jest.mock("./commands", () => ({
  Commands: () => <div>Search</div>,
}));

jest.mock("@/components/shared/sidebar/profile-menu", () => ({
  ProfileMenu: () => <div>Profile</div>,
}));

jest.mock("@/components/shared/sidebar/workspaces-menu", () => ({
  WorkspacesMenu: () => <div>Workspace</div>,
}));

jest.mock("@/components/shared/sidebar/sidebar-context", () => ({
  useSidebar: () => ({ isCollapsed: false }),
}));

jest.mock("@/components/shared/app-command-action-context", () => ({
  useCurrentAppCommandAction: () => null,
}));

describe("AppCommandBar", () => {
  beforeEach(() => {
    mockCompleteWalkthroughAction.mockClear();
    mockStartWalkthrough.mockClear();
    mockCloseWalkthrough.mockClear();
    mockRouterPush.mockClear();
    jest.mocked(toast).mockClear();
    mockUserRole = "admin";
    mockMembersPending = false;
    mockTeamsPending = false;
    mockHasTeams = true;
    mockWalkthroughActive = false;
    window.history.replaceState(null, "", "/acme/my-work");
  });

  it("shows Help with the existing command controls", () => {
    const { container } = render(<AppCommandBar />);

    expect(screen.getByRole("button", { name: "Help" })).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /^Notifications/ }),
    ).toBeInTheDocument();
    expect(
      container.querySelector(
        `[data-walkthrough-target="${walkthroughTargets.notifications}"]`,
      ),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create story" }),
    ).toHaveAttribute("data-walkthrough-target", walkthroughTargets.create);
    expect(
      screen.getByRole("button", { name: "Create story" }),
    ).toHaveAttribute("data-walkthrough-create-kind", "story");
    expect(screen.getByRole("button", { name: "Help" })).toHaveAttribute(
      "data-walkthrough-target",
      walkthroughTargets.help,
    );
    expect(screen.getByText("Profile")).toBeInTheDocument();
  });

  it("opens keyboard shortcuts from Help", () => {
    render(<AppCommandBar />);

    fireEvent.click(screen.getByRole("button", { name: "Keyboard shortcuts" }));

    expect(screen.getByText("Keyboard shortcuts dialog")).toBeInTheDocument();
  });

  it("starts the product tour from Help", () => {
    render(<AppCommandBar />);

    fireEvent.click(screen.getByRole("button", { name: "Product tour" }));

    expect(mockStartWalkthrough).toHaveBeenCalledTimes(1);
  });

  it("completes the walkthrough task action after the task dialog closes", () => {
    render(<AppCommandBar />);

    fireEvent.click(screen.getByRole("button", { name: "Create story" }));
    fireEvent.click(
      screen.getByRole("button", { name: "Simulate story created" }),
    );

    expect(mockCompleteWalkthroughAction).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Close task dialog" }));

    expect(mockCompleteWalkthroughAction).toHaveBeenCalledWith("story-created");
  });

  it("completes after a normally saved task closes its dialog", async () => {
    render(<AppCommandBar />);

    fireEvent.click(screen.getByRole("button", { name: "Create story" }));
    fireEvent.click(
      screen.getByRole("button", { name: "Simulate story saved" }),
    );

    await waitFor(() => {
      expect(mockCompleteWalkthroughAction).toHaveBeenCalledWith(
        "story-created",
      );
    });
  });

  it("waits for workspace membership and teams before consuming the first-task intent", () => {
    window.history.replaceState(
      { retained: true },
      "",
      "/acme/my-work?filter=assigned&onboarding=task&tag=a&tag=b#today",
    );
    mockUserRole = undefined;
    mockMembersPending = true;
    mockTeamsPending = true;
    const { rerender } = render(<AppCommandBar />);
    expect(
      screen.queryByRole("button", { name: "Close task dialog" }),
    ).not.toBeInTheDocument();
    expect(window.location.search).toContain("onboarding=task");

    mockUserRole = "admin";
    mockMembersPending = false;
    rerender(<AppCommandBar />);
    expect(
      screen.queryByRole("button", { name: "Close task dialog" }),
    ).not.toBeInTheDocument();

    mockTeamsPending = false;
    mockWalkthroughActive = true;
    rerender(<AppCommandBar />);
    expect(
      screen.getByRole("button", { name: "Close task dialog" }),
    ).toBeInTheDocument();
    expect(window.location.search).toBe("?filter=assigned&tag=a&tag=b");
    expect(window.location.hash).toBe("#today");
    expect(window.history.state).toEqual({ retained: true });
    expect(mockCloseWalkthrough).toHaveBeenCalled();
    expect(mockStartWalkthrough).not.toHaveBeenCalled();
    expect(mockCompleteWalkthroughAction).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Close task dialog" }));
    rerender(<AppCommandBar />);
    expect(
      screen.queryByRole("button", { name: "Close task dialog" }),
    ).not.toBeInTheDocument();
    expect(toast).not.toHaveBeenCalled();
    expect(mockCompleteWalkthroughAction).not.toHaveBeenCalled();
  });

  it("keeps a delayed automatic tour from covering the onboarding task dialog", () => {
    window.history.replaceState(null, "", "/acme/my-work?onboarding=task");
    const { rerender } = render(<AppCommandBar />);
    mockCloseWalkthrough.mockClear();

    mockWalkthroughActive = true;
    rerender(<AppCommandBar />);

    expect(mockCloseWalkthrough).toHaveBeenCalledTimes(1);
    expect(
      screen.getByRole("button", { name: "Close task dialog" }),
    ).toBeInTheDocument();
    expect(mockCompleteWalkthroughAction).not.toHaveBeenCalled();
    expect(mockStartWalkthrough).not.toHaveBeenCalled();
  });

  it("consumes a guest's intent without opening or replaying it after a role change", () => {
    window.history.replaceState(
      null,
      "",
      "/acme/my-work?onboarding=task&filter=all",
    );
    mockUserRole = "guest";
    const { rerender } = render(<AppCommandBar />);
    expect(window.location.search).toBe("?filter=all");
    expect(
      screen.queryByRole("button", { name: "Close task dialog" }),
    ).not.toBeInTheDocument();

    mockUserRole = "member";
    rerender(<AppCommandBar />);
    expect(
      screen.queryByRole("button", { name: "Close task dialog" }),
    ).not.toBeInTheDocument();
    expect(mockCloseWalkthrough).not.toHaveBeenCalled();
  });

  it("leaves the intent pending when the user has no joined team", () => {
    window.history.replaceState(null, "", "/acme/my-work?onboarding=task");
    mockHasTeams = false;
    render(<AppCommandBar />);
    expect(
      screen.queryByRole("button", { name: "Close task dialog" }),
    ).not.toBeInTheDocument();
    expect(window.location.search).toContain("onboarding=task");
  });

  it("offers optional calendar setup only after the onboarding task is created", async () => {
    window.history.replaceState(null, "", "/acme/my-work?onboarding=task");
    render(<AppCommandBar />);
    expect(toast).not.toHaveBeenCalled();

    fireEvent.click(
      screen.getByRole("button", { name: "Simulate story saved" }),
    );
    await waitFor(() => {
      expect(toast).toHaveBeenCalledTimes(1);
    });
    expect(mockCompleteWalkthroughAction).toHaveBeenCalledWith("story-created");
    expect(toast).toHaveBeenCalledWith(
      "Plan around your meetings",
      expect.objectContaining({
        description: expect.stringContaining("Optional"),
        action: expect.objectContaining({ label: "Connect calendar" }),
      }),
    );
    const action = jest.mocked(toast).mock.calls[0][1]?.action;
    if (!action || typeof action !== "object" || !("onClick" in action))
      throw new Error("Calendar action is missing");
    action.onClick({} as never);
    expect(mockRouterPush).toHaveBeenCalledWith(
      "/acme/settings/account/calendar",
    );
  });
});
