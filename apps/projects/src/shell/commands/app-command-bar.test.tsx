/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { AppCommandBar } from "./app-command-bar";

const mockCompleteWalkthroughAction = jest.fn();

jest.mock("next/navigation", () => ({
  useParams: () => ({}),
  usePathname: () => "/acme/my-work",
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
  useUserRole: () => ({ userRole: "admin" }),
}));

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
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
});
